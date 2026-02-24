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
	// Create PeerConnection with low-latency settings
	settingEngine := webrtc.SettingEngine{}
	settingEngine.SetSRTPReplayProtectionWindow(512)

	api := webrtc.NewAPI(webrtc.WithSettingEngine(settingEngine))

	webrtcCfg := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
		},
	}

	pc, err := api.NewPeerConnection(webrtcCfg)
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

	// 3. Start GStreamer Video Pipeline (Neko-style with optimizations)
	videoPipelineStr := fmt.Sprintf(
		"ximagesrc display-name=:%d show-pointer=true use-damage=false ! "+
			"video/x-raw,framerate=%d/1 ! videoconvert ! queue max-size-buffers=2 ! "+
			"vp8enc target-bitrate=%d cpu-used=6 end-usage=cbr threads=4 deadline=1 lag-in-frames=0 error-resilient=1 keyframe-max-dist=%d ! "+
			"appsink name=appsink emit-signals=true sync=false drop=false max-buffers=2",
		s.cfg.Display.DisplayNum,
		s.cfg.Stream.TargetFPS,
		s.cfg.Stream.MaxBitrateKbps*650, // Neko bitrate mapping
		s.cfg.Stream.TargetFPS,          // 1 keyframe per second
	)

	videoPipeline, err := gst.CreatePipeline(videoPipelineStr)
	if err != nil {
		return err
	}

	// 4. Start GStreamer Audio Pipeline (Neko-style)
	// captures from the .monitor of our null sink
	sinkName := fmt.Sprintf("vbrowser-%d", s.cfg.Display.DisplayNum)
	audioPipelineStr := fmt.Sprintf(
		"pulsesrc device=%s.monitor ! audio/x-raw,channels=2 ! audioconvert ! opusenc bitrate=128000 ! "+
			"appsink name=appsink emit-signals=true sync=false",
		sinkName,
	)

	log.Info().Str("audio_pipeline", audioPipelineStr).Msg("creating audio pipeline")
	audioPipeline, err := gst.CreatePipeline(audioPipelineStr)
	if err != nil {
		log.Error().Err(err).Msg("failed to create audio pipeline")
		return err
	}

	// 5. Start pipelines
	videoPipeline.Play()
	audioPipeline.Play()
	log.Info().Msg("video and audio pipelines started")

	// 6. Handle samples in background
	go func() {
		videoSamples := 0
		audioSamples := 0
		for {
			select {
			case <-s.done:
				log.Info().Int("video_samples", videoSamples).Int("audio_samples", audioSamples).Msg("stopping pipelines")
				videoPipeline.Destroy()
				audioPipeline.Destroy()
				return
			case <-ctx.Done():
				_ = s.Stop()
				return
			case sample := <-videoPipeline.Sample():
				videoSamples++
				_ = s.videoTrack.WriteSample(media.Sample{
					Data:     sample.Data,
					Duration: sample.Duration,
				})
			case sample := <-audioPipeline.Sample():
				audioSamples++
				if audioSamples%100 == 0 {
					log.Debug().Int("count", audioSamples).Msg("audio samples received")
				}
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
