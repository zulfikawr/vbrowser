package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"
	"github.com/rs/zerolog/log"
	"github.com/zulfikawr/vbrowser/internal/browser"
	"github.com/zulfikawr/vbrowser/internal/stream"
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

var lastX, lastY int

// InputMessage represents a user input event.
type InputMessage struct {
	Type   string `json:"type"`
	X      int    `json:"x,omitempty"`
	Y      int    `json:"y,omitempty"`
	DeltaX int    `json:"deltaX,omitempty"`
	DeltaY int    `json:"deltaY,omitempty"`
	Button int    `json:"button,omitempty"`
	Key    string `json:"key,omitempty"`
	Ctrl   bool   `json:"ctrl,omitempty"`
	Alt    bool   `json:"alt,omitempty"`
	Shift  bool   `json:"shift,omitempty"`
	Meta   bool   `json:"meta,omitempty"`
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
		// Store last mouse position
		if msg.Input.Type == "mousemove" {
			lastX = msg.Input.X
			lastY = msg.Input.Y
		}
		return nil

	case "config":
		if msg.Config == nil {
			return fmt.Errorf("config message missing data")
		}
		s.handleConfigChange(msg.Config)

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

	// 1. Stop the current WebRTC/GStreamer session first
	s.mu.Lock()
	if s.currentSession != nil {
		log.Info().Msg("stopping current session for config change")
		_ = s.currentSession.Stop()
		s.currentSession = nil
	}
	s.mu.Unlock()

	// 2. Update global config
	s.cfg.Browser.WindowWidth = cfg.Width
	s.cfg.Browser.WindowHeight = cfg.Height
	s.cfg.Stream.TargetFPS = cfg.FPS
	s.cfg.Stream.MaxBitrateKbps = cfg.Bitrate

	// 3. Restart the browser manager with new resolution
	chromiumPath, _ := browser.GetChromiumPath(s.cfg.Browser.DownloadDir)
	go func() {
		if err := s.mgr.Restart(chromiumPath); err != nil {
			log.Error().Err(err).Msg("failed to restart browser manager")
		}
	}()

	log.Info().Msg("Configuration applied. Please refresh the page to start a new stream with the new resolution.")
}

func (s *Server) handleInput(input *InputMessage) {
	displayStr := fmt.Sprintf(":%d", s.cfg.Display.DisplayNum)
	var args []string

	switch input.Type {
	case "mousemove":
		args = []string{"mousemove", fmt.Sprintf("%d", input.X), fmt.Sprintf("%d", input.Y)}
	case "mousedown":
		button := input.Button + 1 // xdotool uses 1-based buttons
		args = []string{"mousedown", fmt.Sprintf("%d", button)}
	case "mouseup":
		button := input.Button + 1
		args = []string{"mouseup", fmt.Sprintf("%d", button)}
	case "wheel":
		displayStr := fmt.Sprintf(":%d", s.cfg.Display.DisplayNum)
		button := 4
		if input.DeltaY > 0 {
			button = 5
		}

		// Run a single combined command
		var err error
		for i := 0; i < 3; i++ {
			cmd := exec.Command("xdotool", "mousemove", fmt.Sprintf("%d", input.X), fmt.Sprintf("%d", input.Y), "click", fmt.Sprintf("%d", button))
			cmd.Env = append(os.Environ(), "DISPLAY="+displayStr)
			err = cmd.Run()
			if err == nil {
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
		if err != nil {
			log.Warn().Err(err).Msg("failed to execute xdotool wheel after retries")
		}
		return
	case "keydown":
		args = []string{"key", "--clearmodifiers"}
		if input.Ctrl {
			args = append(args, "ctrl")
		}
		if input.Alt {
			args = append(args, "alt")
		}
		if input.Shift {
			args = append(args, "shift")
		}
		if input.Meta {
			args = append(args, "meta")
		}

		key := input.Key
		if len(key) > 1 {
			switch key {
			case "Enter":
				key = "Return"
			case " ":
				key = "space"
			case "Backspace":
				key = "BackSpace"
			case "ArrowLeft":
				key = "Left"
			case "ArrowRight":
				key = "Right"
			case "ArrowUp":
				key = "Up"
			case "ArrowDown":
				key = "Down"
			case "Control":
				return
			case "Alt":
				return
			case "Shift":
				return
			case "Meta":
				return
			}
		} else if key == " " {
			key = "space"
		}

		args = append(args, key)
	case "keyup":
		return
	default:
		return
	}

	cmd := exec.Command("xdotool", args...)
	cmd.Env = append(os.Environ(), "DISPLAY="+displayStr)

	var err error
	for i := 0; i < 3; i++ {
		err = cmd.Run()
		if err == nil {
			return
		}
		cmd = exec.Command("xdotool", args...)
		cmd.Env = append(os.Environ(), "DISPLAY="+displayStr)
		time.Sleep(100 * time.Millisecond)
	}

	if err != nil {
		log.Warn().Err(err).Str("type", input.Type).Msg("failed to execute xdotool after retries")
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
