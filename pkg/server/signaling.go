package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"
	"github.com/rs/zerolog/log"
	"github.com/zulfikawr/vbrowser/internal/platform"
	"github.com/zulfikawr/vbrowser/internal/stream"
	"github.com/zulfikawr/vbrowser/pkg/xorg"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

// SignalingMessage represents a WebSocket signaling message.
type SignalingMessage struct {
	Type      string                     `json:"type"`
	SDP       *webrtc.SessionDescription `json:"sdp,omitempty"`
	Candidate *webrtc.ICECandidateInit   `json:"candidate,omitempty"`
	Input     *InputMessage              `json:"input,omitempty"`
	Clipboard string                     `json:"clipboard,omitempty"`
	Cursor    string                     `json:"cursor,omitempty"`
	Config    *ConfigMessage             `json:"config,omitempty"`
}

// ConfigMessage represents a dynamic configuration change.
type ConfigMessage struct {
	Width   int `json:"width"`
	Height  int `json:"height"`
	FPS     int `json:"fps"`
	Bitrate int `json:"bitrate"`
}

// InputMessage represents a user input event.
type InputMessage struct {
	Type   string  `json:"type"`
	X      int     `json:"x,omitempty"`
	Y      int     `json:"y,omitempty"`
	DeltaX float64 `json:"deltaX,omitempty"`
	DeltaY float64 `json:"deltaY,omitempty"`
	Button int     `json:"button,omitempty"`
	Keysym uint32  `json:"keysym,omitempty"`
}

// handleWebSocket handles WebSocket connections for WebRTC signaling.
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("websocket upgrade failed")
		return
	}
	defer conn.Close()

	log.Info().Str("remote", r.RemoteAddr).Msg("websocket connected")

	// Start a fresh session with current config
	session, err := stream.NewSession(s.cfg)
	if err != nil {
		log.Error().Err(err).Msg("failed to create session")
		s.sendError(conn, "Failed to create session")
		return
	}
	defer func() {
		if err := session.Stop(); err != nil {
			log.Warn().Err(err).Msg("failed to stop session")
		}
	}()

	// Store current session for config hot-reloading
	s.mu.Lock()
	s.currentSession = session
	s.mu.Unlock()

	// Send current config to client so UI matches server state
	s.wsWriteMu.Lock()
	_ = conn.WriteJSON(SignalingMessage{
		Type: "config",
		Config: &ConfigMessage{
			Width:   s.cfg.Browser.WindowWidth,
			Height:  s.cfg.Browser.WindowHeight,
			FPS:     s.cfg.Stream.TargetFPS,
			Bitrate: s.cfg.Stream.MaxBitrateKbps,
		},
	})
	s.wsWriteMu.Unlock()

	// Set up ICE candidate handler
	session.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate == nil {
			return
		}

		candidateInit := candidate.ToJSON()
		msg := SignalingMessage{
			Type:      "candidate",
			Candidate: &candidateInit,
		}

		s.wsWriteMu.Lock()
		if err := conn.WriteJSON(msg); err != nil {
			log.Error().Err(err).Msg("failed to send ICE candidate")
		}
		s.wsWriteMu.Unlock()
	})

	// Set up connection state handler
	session.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		log.Info().Str("state", state.String()).Msg("connection state changed")
	})

	// Start capture loop
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := session.Start(ctx); err != nil {
		log.Error().Err(err).Msg("failed to start session")
		s.sendError(conn, "Failed to start session")
		return
	}

	// Start clipboard watcher
	log.Debug().Msg("launching clipboard watcher goroutine")
	go platform.WatchClipboard(ctx, s.cfg.Display.DisplayNum, func(content string) {
		log.Debug().Int("len", len(content)).Msg("pushing clipboard to client")
		s.wsWriteMu.Lock()
		_ = conn.WriteJSON(SignalingMessage{
			Type:      "clipboard",
			Clipboard: content,
		})
		s.wsWriteMu.Unlock()
	})

	// Handle incoming messages
	for {
		var msg SignalingMessage
		if err := conn.ReadJSON(&msg); err != nil {
			log.Debug().Err(err).Msg("websocket read error")
			break
		}

		if err := s.handleSignalingMessage(session, conn, &msg); err != nil {
			log.Error().Err(err).Str("type", msg.Type).Msg("failed to handle message")
			s.sendError(conn, fmt.Sprintf("Failed to handle %s", msg.Type))
			break
		}
	}

	log.Info().Msg("websocket disconnected")
}

func (s *Server) handleSignalingMessage(session *stream.Session, conn *websocket.Conn, msg *SignalingMessage) error {
	switch msg.Type {
	case "ping":
		s.wsWriteMu.Lock()
		_ = conn.WriteJSON(SignalingMessage{Type: "pong"})
		s.wsWriteMu.Unlock()
		return nil
	case "offer":
		if msg.SDP == nil {
			return fmt.Errorf("offer missing SDP")
		}

		// Set remote description
		if err := session.SetRemoteDescription(*msg.SDP); err != nil {
			return fmt.Errorf("set remote description: %w", err)
		}

		// Create answer
		answer, err := session.CreateAnswer()
		if err != nil {
			return fmt.Errorf("create answer: %w", err)
		}

		// Send answer
		response := SignalingMessage{
			Type: "answer",
			SDP:  &answer,
		}

		s.wsWriteMu.Lock()
		err = conn.WriteJSON(response)
		s.wsWriteMu.Unlock()
		if err != nil {
			return fmt.Errorf("send answer: %w", err)
		}

		log.Info().Msg("sent answer to client")

	case "candidate":
		if msg.Candidate == nil {
			return fmt.Errorf("candidate message missing candidate")
		}

		if err := session.AddICECandidate(*msg.Candidate); err != nil {
			log.Debug().Err(err).Msg("failed to add ICE candidate, likely remote description not set yet")
		}
		return nil

	case "input":
		if msg.Input == nil {
			return fmt.Errorf("input message missing input")
		}
		s.handleInput(msg.Input)
		return nil

	case "config":
		if msg.Config == nil {
			return fmt.Errorf("config message missing data")
		}
		s.handleConfigChange(msg.Config)

	case "pli":
		log.Debug().Msg("received manual PLI request from client")
		if s.currentSession != nil {
			s.currentSession.RequestKeyframe()
		}

	case "clipboard":
		log.Debug().Msg("received clipboard message from client")
		if msg.Clipboard != "" {
			log.Debug().Msg("writing to browser clipboard")
			platform.SetLastContent(msg.Clipboard)
			if err := platform.WriteClipboard(s.cfg.Display.DisplayNum, msg.Clipboard); err != nil {
				log.Error().Err(err).Msg("failed to write clipboard")
			}
		} else {
			// Request clipboard from browser
			content, err := platform.ReadClipboard(s.cfg.Display.DisplayNum)
			if err != nil {
				log.Error().Err(err).Msg("failed to read clipboard")
				return nil
			}
			s.wsWriteMu.Lock()
			_ = conn.WriteJSON(SignalingMessage{
				Type:      "clipboard",
				Clipboard: content,
			})
			s.wsWriteMu.Unlock()
		}

	default:
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}

	return nil
}

func (s *Server) handleConfigChange(cfg *ConfigMessage) {
	log.Info().
		Int("width", cfg.Width).
		Int("height", cfg.Height).
		Int("fps", cfg.FPS).
		Int("bitrate", cfg.Bitrate).
		Msg("updating vbrowser configuration")

	// 1. Stop input batcher to prevent X11 calls during restart
	s.inputBatcher.Stop()

	// 2. Stop the current WebRTC/GStreamer session
	s.mu.Lock()
	if s.currentSession != nil {
		log.Info().Msg("stopping current session for config change")
		_ = s.currentSession.Stop()
		s.currentSession = nil
	}
	s.mu.Unlock()

	// 3. Update global config
	s.cfg.Browser.WindowWidth = cfg.Width
	s.cfg.Browser.WindowHeight = cfg.Height
	s.cfg.Stream.TargetFPS = cfg.FPS
	s.cfg.Stream.MaxBitrateKbps = cfg.Bitrate

	// Save config to persist changes
	if s.configPath != "" {
		if err := s.cfg.Save(s.configPath); err != nil {
			log.Warn().Err(err).Msg("failed to save config")
		}
	}

	// 4. Restart the browser manager with new resolution
	go func() {
		if err := s.mgr.Restart(""); err != nil {
			log.Error().Err(err).Msg("failed to restart browser manager")
		}
		// Recreate input batcher after restart
		s.inputBatcher = NewInputBatcher(func(x, y int) {
			xorg.Move(x, y)
		})
	}()

	log.Info().Msg("Configuration applied. Please refresh the page to start a new stream with the new resolution.")
}

func (s *Server) handleInput(input *InputMessage) {
	switch input.Type {
	case "mousemove":
		s.inputBatcher.AddMouseMove(input.X, input.Y)
	case "mousedown":
		s.inputBatcher.Flush() // Flush pending moves before click
		_ = xorg.ButtonDown(uint32(input.Button + 1))
	case "mouseup":
		_ = xorg.ButtonUp(uint32(input.Button + 1))
	case "wheel":
		s.inputBatcher.Flush() // Flush pending moves before scroll
		xorg.Scroll(input.DeltaX, input.DeltaY, false)
	case "keydown":
		if input.Keysym != 0 {
			_ = xorg.KeyDown(input.Keysym)
		}
	case "keyup":
		if input.Keysym != 0 {
			_ = xorg.KeyUp(input.Keysym)
		}
	}
}

func (s *Server) sendError(conn *websocket.Conn, message string) {
	msg := map[string]string{
		"type":  "error",
		"error": message,
	}
	s.wsWriteMu.Lock()
	defer s.wsWriteMu.Unlock()
	if err := conn.WriteJSON(msg); err != nil {
		log.Error().Err(err).Msg("failed to send error message")
	}
}
