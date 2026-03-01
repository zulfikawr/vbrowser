package stream

import (
	"context"
	"fmt"
	"sync"

	"github.com/pion/rtcp"
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
	videoPipeline  gst.Pipeline
	stopOnce       sync.Once
	done           chan struct{}
}

// NewSession creates a new streaming session.
func NewSession(cfg *config.Config) (*Session, error) {
	// Create PeerConnection with ultra-low-latency settings
	settingEngine := webrtc.SettingEngine{}
	settingEngine.SetSRTPReplayProtectionWindow(512)
	_ = settingEngine.SetAnsweringDTLSRole(webrtc.DTLSRoleServer)

	// Create API with TWCC and PLI support
	m := &webrtc.MediaEngine{}
	if err := m.RegisterDefaultCodecs(); err != nil {
		return nil, err
	}

	// Add TWCC and PLI feedback support
	m.RegisterFeedback(webrtc.RTCPFeedback{Type: "ccm", Parameter: "fir"}, webrtc.RTPCodecTypeVideo)
	m.RegisterFeedback(webrtc.RTCPFeedback{Type: "nack", Parameter: "pli"}, webrtc.RTPCodecTypeVideo)

	api := webrtc.NewAPI(webrtc.WithSettingEngine(settingEngine), webrtc.WithMediaEngine(m))

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
	videoMime := webrtc.MimeTypeVP8
	if s.cfg.Stream.VideoCodec == "h264" {
		videoMime = webrtc.MimeTypeH264
	}

	videoTrack, err := webrtc.NewTrackLocalStaticSample(
		webrtc.RTPCodecCapability{
			MimeType: videoMime,
			RTCPFeedback: []webrtc.RTCPFeedback{
				{Type: "nack", Parameter: "pli"},
				{Type: "ccm", Parameter: "fir"},
			},
		},
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

	// 3. Build Video Pipeline based on codec and encoder choice
	var encoder string
	if s.cfg.Stream.VideoCodec == "h264" {
		// Optimized H.264 for ARM64 using OpenH264 (Software but highly efficient)
		encoder = fmt.Sprintf("openh264enc bitrate=%d complexity=0 multi-thread=2 ! video/x-h264,profile=baseline", s.cfg.Stream.MaxBitrateKbps*1000)
	} else {
		switch s.cfg.Stream.Encoder {
		case "vaapi":
			encoder = fmt.Sprintf("vaapivp8enc bitrate=%d keyframe-period=25", s.cfg.Stream.MaxBitrateKbps)
		case "nvenc":
			encoder = fmt.Sprintf("nvvp8enc bitrate=%d gop-size=25", s.cfg.Stream.MaxBitrateKbps*1000)
		default:
			encoder = fmt.Sprintf("vp8enc target-bitrate=%d cpu-used=16 deadline=1 end-usage=cbr threads=4 static-threshold=0 lag-in-frames=0 undershoot=95 buffer-size=%d buffer-initial-size=%d buffer-optimal-size=%d error-resilient=1 keyframe-max-dist=25 min-quantizer=4 max-quantizer=20",
				s.cfg.Stream.MaxBitrateKbps*650,
				s.cfg.Stream.MaxBitrateKbps*4,
				s.cfg.Stream.MaxBitrateKbps*2,
				s.cfg.Stream.MaxBitrateKbps*3)
		}
	}

	videoPipelineStr := fmt.Sprintf(
		"ximagesrc display-name=:%d show-pointer=true use-damage=false ! "+
			"video/x-raw,framerate=%d/1 ! videoconvert ! queue max-size-buffers=1 ! "+
			"%s ! appsink name=appsink emit-signals=true sync=false drop=true max-buffers=1",
		s.cfg.Display.DisplayNum,
		s.cfg.Stream.TargetFPS,
		encoder,
	)

	log.Info().
		Str("codec", s.cfg.Stream.VideoCodec).
		Str("encoder", s.cfg.Stream.Encoder).
		Msg("starting video pipeline")
	videoPipeline, err := gst.CreatePipeline(videoPipelineStr)
	if err != nil {
		return err
	}
	s.videoPipeline = videoPipeline

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

	// 4. Handle RTCP Feedback (PLI) to force keyframes
	for _, receiver := range s.peerConnection.GetReceivers() {
		if receiver.Track() == nil {
			continue
		}

		go func(r *webrtc.RTPReceiver) {
			for {
				pkts, _, err := r.ReadRTCP()
				if err != nil {
					return
				}

				for _, pkt := range pkts {
					if _, ok := pkt.(*rtcp.PictureLossIndication); ok {
						log.Debug().Msg("received PLI, forcing keyframe")
						if videoPipeline != nil {
							videoPipeline.EmitVideoKeyframe()
						}
					}
				}
			}
		}(receiver)
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

// RequestKeyframe forces the video pipeline to emit a new keyframe.
func (s *Session) RequestKeyframe() {
	if s.videoPipeline != nil {
		s.videoPipeline.EmitVideoKeyframe()
	}
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
