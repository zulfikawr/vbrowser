package stream

import (
	"context"
	"fmt"
	"sync"

	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
	"github.com/zulfikawr/vbrowser/internal/config"
	"github.com/zulfikawr/vbrowser/pkg/gst"
)

type Session struct {
	id             string
	cfg            *config.Config
	peerConnection *webrtc.PeerConnection
	videoTrack     *webrtc.TrackLocalStaticSample
	pipeline       gst.Pipeline
	stopChan       chan struct{}
	stopOnce       sync.Once
	wg             sync.WaitGroup
}

func NewSession(id string, cfg *config.Config) (*Session, error) {
	webrtcConfig := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{{URLs: []string{"stun:stun.l.google.com:19302"}}},
	}

	peerConnection, err := webrtc.NewPeerConnection(webrtcConfig)
	if err != nil {
		return nil, err
	}

	videoTrack, err := webrtc.NewTrackLocalStaticSample(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8},
		"video", "vbrowser",
	)
	if err != nil {
		return nil, err
	}

	if _, err := peerConnection.AddTrack(videoTrack); err != nil {
		return nil, err
	}

	// GStreamer pipeline string (Neko-style)
	// captures from ximagesrc (Xvfb) and encodes to VP8
	pipelineStr := fmt.Sprintf(
		"ximagesrc display-name=:%d show-pointer=true use-damage=false ! "+
			"video/x-raw,framerate=%d/1 ! videoconvert ! queue ! "+
			"vp8enc target-bitrate=%d deadline=1 cpu-used=8 threads=4 keyframe-max-dist=30 name=encoder ! "+
			"appsink name=appsink",
		cfg.Display.DisplayNum,
		cfg.Stream.TargetFPS,
		cfg.Stream.MaxBitrateKbps*1000,
	)

	pipeline, err := gst.CreatePipeline(pipelineStr)
	if err != nil {
		return nil, fmt.Errorf("gst pipeline: %w", err)
	}

	return &Session{
		id:             id,
		cfg:            cfg,
		peerConnection: peerConnection,
		videoTrack:     videoTrack,
		pipeline:       pipeline,
		stopChan:       make(chan struct{}),
	}, nil
}

func (s *Session) Start(ctx context.Context) error {
	s.pipeline.Play()
	s.wg.Add(1)
	go s.streamLoop(ctx)
	return nil
}

func (s *Session) streamLoop(ctx context.Context) {
	defer s.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case sample, ok := <-s.pipeline.Sample():
			if !ok {
				return
			}
			_ = s.videoTrack.WriteSample(media.Sample{
				Data:     sample.Data,
				Duration: sample.Duration,
			})
		}
	}
}

func (s *Session) Stop() error {
	s.stopOnce.Do(func() {
		close(s.stopChan)
		s.pipeline.Destroy()
		s.peerConnection.Close()
	})
	s.wg.Wait()
	return nil
}

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

func (s *Session) SetRemoteDescription(sdp webrtc.SessionDescription) error {
	return s.peerConnection.SetRemoteDescription(sdp)
}

func (s *Session) AddICECandidate(candidate webrtc.ICECandidateInit) error {
	return s.peerConnection.AddICECandidate(candidate)
}

func (s *Session) OnICECandidate(handler func(*webrtc.ICECandidate)) {
	s.peerConnection.OnICECandidate(handler)
}

func (s *Session) OnConnectionStateChange(handler func(webrtc.PeerConnectionState)) {
	s.peerConnection.OnConnectionStateChange(handler)
}
