package stream

import (
	"context"
	"fmt"
	"sync"

	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
	"github.com/rs/zerolog/log"
	"github.com/zulfikawr/vbrowser/internal/config"
	"github.com/zulfikawr/vbrowser/pkg/gst"
)

// Session represents a single WebRTC streaming session.
type Session struct {
	cfg            *config.Config
	peerConnection *webrtc.PeerConnection
	videoTrack     *webrtc.TrackLocalStaticSample
	audioTrack     *webrtc.TrackLocalStaticSample
	stopOnce       sync.Once
	done           chan struct{}
}

// NewSession creates a new streaming session.
func NewSession(cfg *config.Config) (*Session, error) {
	// Create PeerConnection
	webrtcCfg := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
		},
	}

	pc, err := webrtc.NewPeerConnection(webrtcCfg)
	if err != nil {
		return nil, err
	}

	return &Session{
		cfg:            cfg,
		peerConnection: pc,
		done:           make(chan struct{}),
	}, nil
}

// Start begins the GStreamer pipelines and WebRTC streaming.
func (s *Session) Start(ctx context.Context) error {
	// 1. Create WebRTC tracks
	videoTrack, err := webrtc.NewTrackLocalStaticSample(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8},
		"video", "vbrowser",
	)
	if err != nil {
		return err
	}
	s.videoTrack = videoTrack

	audioTrack, err := webrtc.NewTrackLocalStaticSample(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus},
		"audio", "vbrowser",
	)
	if err != nil {
		return err
	}
	s.audioTrack = audioTrack

	// 2. Add tracks to PeerConnection
	if _, err := s.peerConnection.AddTrack(videoTrack); err != nil {
		return err
	}
	if _, err := s.peerConnection.AddTrack(audioTrack); err != nil {
		return err
	}

	// 3. Start GStreamer Video Pipeline (Neko-style)
	videoPipelineStr := fmt.Sprintf(
		"ximagesrc display-name=:%d show-pointer=true use-damage=false ! "+
			"video/x-raw,framerate=%d/1 ! videoconvert ! queue ! "+
			"vp8enc target-bitrate=%d cpu-used=4 end-usage=cbr threads=4 deadline=1 undershoot=95 ! "+
			"appsink name=appsink emit-signals=true sync=false drop=true max-buffers=1",
		s.cfg.Display.DisplayNum,
		s.cfg.Stream.TargetFPS,
		s.cfg.Stream.MaxBitrateKbps*650, // Neko bitrate mapping
	)

	videoPipeline, err := gst.CreatePipeline(videoPipelineStr)
	if err != nil {
		return err
	}

	// 4. Start GStreamer Audio Pipeline (Neko-style)
	// captures from the .monitor of our null sink
	sinkName := fmt.Sprintf("vbrowser-%d.monitor", s.cfg.Display.DisplayNum)
	audioPipelineStr := fmt.Sprintf(
		"pulsesrc device=%s ! audio/x-raw,channels=2 ! audioconvert ! opusenc bitrate=128000 ! "+
			"appsink name=appsink emit-signals=true sync=false",
		sinkName,
	)

	audioPipeline, err := gst.CreatePipeline(audioPipelineStr)
	if err != nil {
		return err
	}

	// 5. Start pipelines
	videoPipeline.Play()
	audioPipeline.Play()

	// 6. Handle samples in background
	go func() {
		for {
			select {
			case <-s.done:
				videoPipeline.Destroy()
				audioPipeline.Destroy()
				return
			case <-ctx.Done():
				s.Stop()
				return
			case sample := <-videoPipeline.Sample():
				_ = s.videoTrack.WriteSample(media.Sample{
					Data:     sample.Data,
					Duration: sample.Duration,
				})
			case sample := <-audioPipeline.Sample():
				_ = s.audioTrack.WriteSample(media.Sample{
					Data:     sample.Data,
					Duration: sample.Duration,
				})
			}
		}
	}()

	return nil
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
		log.Info().Msg("stopping streaming session")
		close(s.done)
		if s.peerConnection != nil {
			err = s.peerConnection.Close()
		}
	})
	return err
}
