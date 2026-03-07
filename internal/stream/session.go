package stream

import (
	"sync"

	"github.com/pion/webrtc/v4"
	"github.com/rs/zerolog/log"
	"github.com/zulfikawr/vbrowser/internal/config"
)

// Session represents a single WebRTC connection to a client.
type Session struct {
	ID             string
	cfg            *config.Config
	peerConnection *webrtc.PeerConnection
	stopOnce       sync.Once
}

// NewSession creates a new WebRTC session and subscribes it to the broadcaster.
func NewSession(cfg *config.Config, b *Broadcaster, id string) (*Session, error) {
	settingEngine := webrtc.SettingEngine{}
	settingEngine.SetSRTPReplayProtectionWindow(512)
	_ = settingEngine.SetAnsweringDTLSRole(webrtc.DTLSRoleServer)

	m := &webrtc.MediaEngine{}
	if err := m.RegisterDefaultCodecs(); err != nil {
		return nil, err
	}

	m.RegisterFeedback(webrtc.RTCPFeedback{Type: "ccm", Parameter: "fir"}, webrtc.RTPCodecTypeVideo)
	m.RegisterFeedback(webrtc.RTCPFeedback{Type: "nack", Parameter: "pli"}, webrtc.RTPCodecTypeVideo)

	api := webrtc.NewAPI(webrtc.WithSettingEngine(settingEngine), webrtc.WithMediaEngine(m))

	webrtcCfg := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{
				"stun:stun.l.google.com:19302",
				"stun:stun1.l.google.com:19302",
				"stun:stun2.l.google.com:19302",
				"stun:stun3.l.google.com:19302",
				"stun:stun4.l.google.com:19302",
			}},
		},
	}

	pc, err := api.NewPeerConnection(webrtcCfg)
	if err != nil {
		return nil, err
	}

	// Add the shared tracks from the Broadcaster
	videoTrack := b.VideoTrack()
	audioTrack := b.AudioTrack()
	log.Debug().Str("session", id).Str("kind", "video").Msg("adding shared track to peer connection")
	if _, err := pc.AddTrack(videoTrack); err != nil {
		log.Error().Err(err).Msg("failed to add video track")
		return nil, err
	}
	log.Debug().Str("session", id).Str("kind", "audio").Msg("adding shared track to peer connection")
	if _, err := pc.AddTrack(audioTrack); err != nil {
		log.Error().Err(err).Msg("failed to add audio track")
		return nil, err
	}

	// Handle PLI feedback
	for _, receiver := range pc.GetReceivers() {
		if receiver.Track() != nil && receiver.Track().Kind() == webrtc.RTPCodecTypeVideo {
			b.HandlePLI(receiver)
		}
	}

	return &Session{
		ID:             id,
		cfg:            cfg,
		peerConnection: pc,
	}, nil
}

// SetRemoteDescription sets the session's remote description.
func (s *Session) SetRemoteDescription(desc webrtc.SessionDescription) error {
	return s.peerConnection.SetRemoteDescription(desc)
}

// CreateAnswer creates a WebRTC answer for the session.
func (s *Session) CreateAnswer() (webrtc.SessionDescription, error) {
	answer, err := s.peerConnection.CreateAnswer(nil)
	if err != nil {
		return webrtc.SessionDescription{}, err
	}

	if err := s.peerConnection.SetLocalDescription(answer); err != nil {
		return webrtc.SessionDescription{}, err
	}

	return answer, nil
}

// AddICECandidate adds an ICE candidate to the session.
func (s *Session) AddICECandidate(candidate webrtc.ICECandidateInit) error {
	return s.peerConnection.AddICECandidate(candidate)
}

// OnICECandidate sets the ICE candidate handler.
func (s *Session) OnICECandidate(f func(*webrtc.ICECandidate)) {
	s.peerConnection.OnICECandidate(f)
}

// OnConnectionStateChange sets the connection state handler.
func (s *Session) OnConnectionStateChange(f func(webrtc.PeerConnectionState)) {
	s.peerConnection.OnConnectionStateChange(f)
}

// Stop terminates the session.
func (s *Session) Stop() error {
	var err error
	s.stopOnce.Do(func() {
		log.Info().Str("session", s.ID).Msg("stopping webrtc session")
		if s.peerConnection != nil {
			err = s.peerConnection.Close()
		}
	})
	return err
}
