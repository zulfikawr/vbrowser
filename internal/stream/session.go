package stream

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
	"github.com/rs/zerolog/log"
	"github.com/zulfikawr/vbrowser/internal/capture"
	"github.com/zulfikawr/vbrowser/internal/config"
)

// Session represents a WebRTC streaming session.
type Session struct {
	id             string
	cfg            *config.Config
	peerConnection *webrtc.PeerConnection
	videoTrack     *webrtc.TrackLocalStaticSample
	capturer       capture.Capturer
	encoder        *Encoder
	stopChan       chan struct{}
	wg             sync.WaitGroup
}

// NewSession creates a new WebRTC streaming session.
func NewSession(id string, cfg *config.Config, capturer capture.Capturer) (*Session, error) {
	// Create WebRTC configuration
	webrtcConfig := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	// Create peer connection
	peerConnection, err := webrtc.NewPeerConnection(webrtcConfig)
	if err != nil {
		return nil, fmt.Errorf("create peer connection: %w", err)
	}

	// Create video track
	videoTrack, err := webrtc.NewTrackLocalStaticSample(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8},
		"video",
		"vbrowser",
	)
	if err != nil {
		peerConnection.Close()
		return nil, fmt.Errorf("create video track: %w", err)
	}

	// Add track to peer connection
	if _, err := peerConnection.AddTrack(videoTrack); err != nil {
		peerConnection.Close()
		return nil, fmt.Errorf("add track: %w", err)
	}

	// Create encoder
	encoder, err := NewEncoder(
		cfg.Browser.WindowWidth,
		cfg.Browser.WindowHeight,
		cfg.Stream.TargetFPS,
		cfg.Stream.MaxBitrateKbps,
	)
	if err != nil {
		peerConnection.Close()
		return nil, fmt.Errorf("create encoder: %w", err)
	}

	log.Info().Str("session_id", id).Msg("WebRTC session created")

	return &Session{
		id:             id,
		cfg:            cfg,
		peerConnection: peerConnection,
		videoTrack:     videoTrack,
		capturer:       capturer,
		encoder:        encoder,
		stopChan:       make(chan struct{}),
	}, nil
}

// Start begins capturing and streaming frames.
func (s *Session) Start(ctx context.Context) error {
	log.Info().Str("session_id", s.id).Msg("starting capture loop")

	s.wg.Add(1)
	go s.captureLoop(ctx)

	return nil
}

// Stop stops the streaming session.
func (s *Session) Stop() error {
	log.Info().Str("session_id", s.id).Msg("stopping session")

	close(s.stopChan)
	s.wg.Wait()

	if s.encoder != nil {
		s.encoder.Close()
	}

	if s.peerConnection != nil {
		s.peerConnection.Close()
	}

	return nil
}

// CreateAnswer creates a WebRTC answer in response to an offer.
func (s *Session) CreateAnswer() (webrtc.SessionDescription, error) {
	answer, err := s.peerConnection.CreateAnswer(nil)
	if err != nil {
		return webrtc.SessionDescription{}, fmt.Errorf("create answer: %w", err)
	}

	if err := s.peerConnection.SetLocalDescription(answer); err != nil {
		return webrtc.SessionDescription{}, fmt.Errorf("set local description: %w", err)
	}

	return answer, nil
}

// CreateOffer creates a WebRTC offer.
func (s *Session) CreateOffer() (webrtc.SessionDescription, error) {
	offer, err := s.peerConnection.CreateOffer(nil)
	if err != nil {
		return webrtc.SessionDescription{}, fmt.Errorf("create offer: %w", err)
	}

	if err := s.peerConnection.SetLocalDescription(offer); err != nil {
		return webrtc.SessionDescription{}, fmt.Errorf("set local description: %w", err)
	}

	return offer, nil
}

// SetRemoteDescription sets the remote SDP description.
func (s *Session) SetRemoteDescription(sdp webrtc.SessionDescription) error {
	return s.peerConnection.SetRemoteDescription(sdp)
}

// AddICECandidate adds an ICE candidate.
func (s *Session) AddICECandidate(candidate webrtc.ICECandidateInit) error {
	return s.peerConnection.AddICECandidate(candidate)
}

// OnICECandidate sets the ICE candidate callback.
func (s *Session) OnICECandidate(handler func(*webrtc.ICECandidate)) {
	s.peerConnection.OnICECandidate(handler)
}

// OnConnectionStateChange sets the connection state change callback.
func (s *Session) OnConnectionStateChange(handler func(webrtc.PeerConnectionState)) {
	s.peerConnection.OnConnectionStateChange(handler)
}

func (s *Session) captureLoop(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(time.Second / time.Duration(s.cfg.Stream.TargetFPS))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			if err := s.captureAndSend(); err != nil {
				if err != io.ErrClosedPipe {
					log.Error().Err(err).Msg("capture and send failed")
				}
				return
			}
		}
	}
}

func (s *Session) captureAndSend() error {
	// Capture frame
	frame, err := s.capturer.Capture()
	if err != nil {
		return fmt.Errorf("capture: %w", err)
	}

	// For now, we'll send raw frames
	// WebRTC will handle VP8 encoding internally
	if err := s.videoTrack.WriteSample(media.Sample{
		Data:     frame.Pix,
		Duration: time.Second / time.Duration(s.cfg.Stream.TargetFPS),
	}); err != nil {
		return fmt.Errorf("write sample: %w", err)
	}

	return nil
}
