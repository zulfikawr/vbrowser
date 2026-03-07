package server

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"
	"github.com/rs/zerolog/log"
	"github.com/zulfikawr/vbrowser/internal/platform"
	"github.com/zulfikawr/vbrowser/internal/stream"
	"github.com/zulfikawr/vbrowser/pkg/utils"
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
	SessionID string                     `json:"session_id,omitempty"`
	Config    *ConfigMessage             `json:"config,omitempty"`
	Control   *ControlMessage            `json:"control,omitempty"`
	Sessions  []SessionInfo              `json:"sessions,omitempty"`
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

// ControlMessage represents control handover.
type ControlMessage struct {
	Action   string `json:"action,omitempty"` // "request", "give", "kick", "background"
	TargetID string `json:"target_id,omitempty"`
	IsHost   bool   `json:"is_host"`
	Value    bool   `json:"value,omitempty"` // Used for background state
}

type SessionInfo struct {
	ID     string `json:"id"`
	IsHost bool   `json:"is_host"`
	Remote string `json:"remote"`
}

type Client struct {
	ID           string
	Conn         *websocket.Conn
	Session      *stream.Session
	IsBackground bool
	mu           sync.Mutex
}

func (c *Client) WriteJSON(v interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Conn.WriteJSON(v)
}

// handleWebSocket handles WebSocket connections for WebRTC signaling.
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("websocket upgrade failed")
		return
	}

	sessionID, err := utils.NewUID(16)
	if err != nil {
		log.Error().Err(err).Msg("failed to generate session id")
		_ = conn.Close()
		return
	}
	log.Info().Str("remote", r.RemoteAddr).Str("id", sessionID).Msg("websocket connected")

	session, err := stream.NewSession(s.cfg, s.broadcaster, sessionID)
	if err != nil {
		log.Error().Err(err).Msg("failed to create session")
		_ = conn.Close()
		return
	}

	client := &Client{
		ID:      sessionID,
		Conn:    conn,
		Session: session,
	}

	// Register session
	s.mu.Lock()
	s.clients[sessionID] = client
	if s.hostID == "" {
		s.hostID = sessionID
		log.Info().Str("id", sessionID).Msg("session assigned as host")
	}
	s.mu.Unlock()

	defer func() {
		_ = session.Stop()
		_ = conn.Close()
		s.mu.Lock()
		delete(s.clients, sessionID)
		// Only promote if the host actually disconnected
		if s.hostID == sessionID {
			s.hostID = ""
			// Prefer non-background clients for host promotion
			for id, c := range s.clients {
				if !c.IsBackground {
					s.hostID = id
					break
				}
			}
			// Fallback to any client if all are backgrounded
			if s.hostID == "" {
				for id := range s.clients {
					s.hostID = id
					break
				}
			}
			if s.hostID != "" {
				log.Info().Str("id", s.hostID).Msg("host disconnected, promoted new host")
			}
		}
		s.mu.Unlock()
		s.broadcastHostStatus()
	}()

	// Send current config to client
	_ = client.WriteJSON(SignalingMessage{
		Type:      "config",
		SessionID: sessionID,
		Config: &ConfigMessage{
			Width:   s.cfg.Browser.WindowWidth,
			Height:  s.cfg.Browser.WindowHeight,
			FPS:     s.cfg.Stream.TargetFPS,
			Bitrate: s.cfg.Stream.MaxBitrateKbps,
		},
		Control: &ControlMessage{
			IsHost: s.hostID == sessionID,
		},
	})

	s.broadcastHostStatus()

	// Set up ICE candidate handler
	session.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate == nil {
			return
		}

		candidateInit := candidate.ToJSON()
		if candidateInit.Candidate == "" {
			return
		}

		msg := SignalingMessage{
			Type:      "candidate",
			Candidate: &candidateInit,
		}

		if err := client.WriteJSON(msg); err != nil {
			log.Error().Err(err).Msg("failed to send ICE candidate")
		}
	})

	// Set up connection state handler
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	session.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		log.Info().Str("state", state.String()).Str("id", sessionID).Msg("connection state changed")
		if state == webrtc.PeerConnectionStateClosed || state == webrtc.PeerConnectionStateFailed {
			cancel()
		}
	})

	// Push initial clipboard state if host
	go func() {
		time.Sleep(1 * time.Second)
		s.mu.Lock()
		isHost := s.hostID == sessionID
		s.mu.Unlock()
		if isHost {
			content, _ := platform.ReadClipboard(s.cfg.Display.DisplayNum)
			if content != "" {
				_ = client.WriteJSON(SignalingMessage{
					Type:      "clipboard",
					Clipboard: content,
				})
			}
		}
	}()

	var descriptionSet bool
	var pendingCandidates []webrtc.ICECandidateInit
	var candidatesMu sync.Mutex

	// Handle incoming messages
	for {
		var msg SignalingMessage
		if err := conn.ReadJSON(&msg); err != nil {
			log.Debug().Err(err).Msg("websocket read error")
			break
		}

		handleMsg := func() error {
			switch msg.Type {
			case "ping":
				_ = client.WriteJSON(SignalingMessage{Type: "pong"})

			case "offer":
				if msg.SDP == nil {
					return fmt.Errorf("offer missing SDP")
				}
				if err := session.SetRemoteDescription(*msg.SDP); err != nil {
					return fmt.Errorf("set remote description: %w", err)
				}
				answer, err := session.CreateAnswer()
				if err != nil {
					return fmt.Errorf("create answer: %w", err)
				}
				_ = client.WriteJSON(SignalingMessage{
					Type: "answer",
					SDP:  &answer,
				})

				candidatesMu.Lock()
				descriptionSet = true
				for _, c := range pendingCandidates {
					_ = session.AddICECandidate(c)
				}
				pendingCandidates = nil
				candidatesMu.Unlock()

			case "candidate":
				if msg.Candidate == nil {
					return fmt.Errorf("candidate message missing data")
				}
				candidatesMu.Lock()
				if !descriptionSet {
					pendingCandidates = append(pendingCandidates, *msg.Candidate)
					candidatesMu.Unlock()
					return nil
				}
				candidatesMu.Unlock()
				_ = session.AddICECandidate(*msg.Candidate)

			case "input":
				if msg.Input == nil {
					return fmt.Errorf("input message missing data")
				}
				s.mu.Lock()
				isHost := s.hostID == sessionID
				s.mu.Unlock()
				if isHost {
					s.handleInput(msg.Input)
				}

			case "config":
				if msg.Config == nil {
					return fmt.Errorf("config message missing data")
				}
				s.mu.Lock()
				isHost := s.hostID == sessionID
				s.mu.Unlock()
				if isHost {
					s.handleConfigChange(context.Background(), msg.Config)
				}

			case "control":
				if msg.Control != nil {
					switch msg.Control.Action {
					case "request":
						s.mu.Lock()
						s.hostID = sessionID
						log.Info().Str("id", sessionID).Msg("session took control")
						s.mu.Unlock()
						s.broadcastHostStatus()
					case "give":
						s.mu.Lock()
						if s.hostID == sessionID && msg.Control.TargetID != "" {
							s.hostID = msg.Control.TargetID
							log.Info().Str("from", sessionID).Str("to", msg.Control.TargetID).Msg("control transferred")
						}
						s.mu.Unlock()
						s.broadcastHostStatus()
					case "kick":
						s.mu.Lock()
						if s.hostID == sessionID && msg.Control.TargetID != "" && msg.Control.TargetID != sessionID {
							if target, ok := s.clients[msg.Control.TargetID]; ok {
								log.Info().Str("by", sessionID).Str("target", msg.Control.TargetID).Msg("session kicked")
								_ = target.Session.Stop()
								_ = target.Conn.Close()
							}
						}
						s.mu.Unlock()
						s.broadcastHostStatus()
					case "background":
						s.mu.Lock()
						if c, ok := s.clients[sessionID]; ok {
							c.IsBackground = msg.Control.Value
						}
						s.mu.Unlock()
					}
				}

			case "pli":
				s.broadcaster.RequestKeyframe()

			case "clipboard":
				s.mu.Lock()
				isHost := s.hostID == sessionID
				s.mu.Unlock()
				if isHost {
					if msg.Clipboard != "" {
						platform.SetLastContent(msg.Clipboard)
						_ = platform.WriteClipboard(s.cfg.Display.DisplayNum, msg.Clipboard)
					} else {
						content, err := platform.ReadClipboard(s.cfg.Display.DisplayNum)
						if err == nil {
							_ = client.WriteJSON(SignalingMessage{
								Type:      "clipboard",
								Clipboard: content,
							})
						}
					}
				}

			default:
				return fmt.Errorf("unknown message type: %s", msg.Type)
			}
			return nil
		}

		if err := handleMsg(); err != nil {
			log.Error().Err(err).Msg("failed to handle signaling message")
		}
	}

	log.Info().Str("id", sessionID).Msg("websocket disconnected")
}

func (s *Server) broadcastHostStatus() {
	s.mu.Lock()
	defer s.mu.Unlock()

	var sessions []SessionInfo
	for id, c := range s.clients {
		sessions = append(sessions, SessionInfo{
			ID:     id,
			IsHost: id == s.hostID,
			Remote: c.Conn.RemoteAddr().String(),
		})
	}

	for id, c := range s.clients {
		_ = c.WriteJSON(SignalingMessage{
			Type: "control",
			Control: &ControlMessage{
				IsHost: id == s.hostID,
			},
			Sessions: sessions,
		})
	}
}

func (s *Server) startClipboardWatcher() {
	go platform.WatchClipboard(context.Background(), s.cfg.Display.DisplayNum, func(content string) {
		s.mu.Lock()
		defer s.mu.Unlock()
		if s.hostID != "" {
			if client, ok := s.clients[s.hostID]; ok {
				_ = client.WriteJSON(SignalingMessage{
					Type:      "clipboard",
					Clipboard: content,
				})
			}
		}
	})
}

func (s *Server) handleConfigChange(ctx context.Context, msg *ConfigMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Info().Msg("applying new configuration")

	resChanged := s.cfg.Browser.WindowWidth != msg.Width || s.cfg.Browser.WindowHeight != msg.Height

	s.cfg.Browser.WindowWidth = msg.Width
	s.cfg.Browser.WindowHeight = msg.Height
	s.cfg.Stream.TargetFPS = msg.FPS
	s.cfg.Stream.MaxBitrateKbps = msg.Bitrate

	_ = s.cfg.Save(s.configPath)

	if resChanged {
		go func() {
			_ = s.mgr.Restart("")
			s.broadcaster.RestartPipelines(ctx)
		}()
	} else {
		go s.broadcaster.RestartPipelines(ctx)
	}
}

func (s *Server) handleInput(input *InputMessage) {
	switch input.Type {
	case "mousemove":
		s.inputBatcher.AddMouseMove(input.X, input.Y)
	case "mousedown":
		s.inputBatcher.Flush()
		_ = xorg.ButtonDown(uint32(input.Button + 1))
	case "mouseup":
		_ = xorg.ButtonUp(uint32(input.Button + 1))
	case "wheel":
		s.inputBatcher.Flush()
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
