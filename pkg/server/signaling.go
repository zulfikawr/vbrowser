package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"
	"github.com/rs/zerolog/log"
	"github.com/zulfikawr/vbrowser/internal/capture"
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

	// Create capturer (placeholder - will be integrated with browser manager)
	capturer, err := s.createCapturer()
	if err != nil {
		log.Error().Err(err).Msg("failed to create capturer")
		s.sendError(conn, "Failed to create capturer")
		return
	}
	defer capturer.Close()

	// Create WebRTC session
	session, err := stream.NewSession("session-1", s.cfg, capturer)
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

		if err := conn.WriteJSON(msg); err != nil {
			log.Error().Err(err).Msg("failed to send ICE candidate")
		}
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

		if err := conn.WriteJSON(response); err != nil {
			return fmt.Errorf("send answer: %w", err)
		}

		log.Info().Msg("sent answer to client")

	case "candidate":
		if msg.Candidate == nil {
			return fmt.Errorf("candidate message missing candidate")
		}

		if err := session.AddICECandidate(*msg.Candidate); err != nil {
			return fmt.Errorf("add ICE candidate: %w", err)
		}

		log.Debug().Msg("added ICE candidate")

	default:
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}

	return nil
}

func (s *Server) sendError(conn *websocket.Conn, message string) {
	msg := map[string]string{
		"type":  "error",
		"error": message,
	}
	if err := conn.WriteJSON(msg); err != nil {
		log.Error().Err(err).Msg("failed to send error message")
	}
}

func (s *Server) createCapturer() (capture.Capturer, error) {
	// For now, create a capturer based on config
	// This will be integrated with the browser manager in the next phase
	return capture.NewXvfbCapturer(
		s.cfg.Display.DisplayNum,
		s.cfg.Browser.WindowWidth,
		s.cfg.Browser.WindowHeight,
	)
}
