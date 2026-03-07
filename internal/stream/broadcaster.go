package stream

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
	"github.com/rs/zerolog/log"
	"github.com/zulfikawr/vbrowser/internal/config"
	"github.com/zulfikawr/vbrowser/pkg/gst"
)

// Broadcaster manages a single set of GStreamer pipelines and WebRTC tracks,
// allowing multiple clients to subscribe to the same stream.
type Broadcaster struct {
	cfg           *config.Config
	videoTrack    *webrtc.TrackLocalStaticSample
	audioTrack    *webrtc.TrackLocalStaticSample
	videoPipeline gst.Pipeline
	audioPipeline gst.Pipeline
	done          chan struct{}
	pipelinesDone chan struct{}
	mu            sync.Mutex
	running       bool
}

func NewBroadcaster(cfg *config.Config) (*Broadcaster, error) {
	videoMime := webrtc.MimeTypeVP8
	if cfg.Stream.VideoCodec == "h264" {
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
		return nil, err
	}

	audioTrack, err := webrtc.NewTrackLocalStaticSample(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus},
		"audio", "vbrowser",
	)
	if err != nil {
		return nil, err
	}

	return &Broadcaster{
		cfg:        cfg,
		videoTrack: videoTrack,
		audioTrack: audioTrack,
		done:       make(chan struct{}),
	}, nil
}

func (b *Broadcaster) Start(ctx context.Context) error {
	b.mu.Lock()
	if b.running {
		b.mu.Unlock()
		return nil
	}
	b.running = true
	b.mu.Unlock()

	return b.StartPipelines(ctx)
}

func (b *Broadcaster) StartPipelines(ctx context.Context) error {
	b.pipelinesDone = make(chan struct{})

	// Build Video Pipeline
	var encoder string
	if b.cfg.Stream.VideoCodec == "h264" {
		encoder = fmt.Sprintf("openh264enc bitrate=%d complexity=0 multi-thread=2 ! video/x-h264,profile=baseline", b.cfg.Stream.MaxBitrateKbps*1000)
	} else {
		switch b.cfg.Stream.Encoder {
		case "vaapi":
			encoder = fmt.Sprintf("vaapivp8enc bitrate=%d keyframe-period=25", b.cfg.Stream.MaxBitrateKbps)
		case "nvenc":
			encoder = fmt.Sprintf("nvvp8enc bitrate=%d gop-size=25", b.cfg.Stream.MaxBitrateKbps*1000)
		default:
			// Ultra-low latency VP8 tuning
			encoder = fmt.Sprintf("vp8enc target-bitrate=%d cpu-used=16 deadline=1 end-usage=cbr threads=4 static-threshold=0 lag-in-frames=0 undershoot=95 buffer-size=%d buffer-initial-size=%d buffer-optimal-size=%d error-resilient=partitions keyframe-max-dist=10 auto-alt-ref=true min-quantizer=4 max-quantizer=20",
				b.cfg.Stream.MaxBitrateKbps*650,
				b.cfg.Stream.MaxBitrateKbps*4,
				b.cfg.Stream.MaxBitrateKbps*2,
				b.cfg.Stream.MaxBitrateKbps*3)
		}
	}

	videoPipelineStr := fmt.Sprintf(
		"ximagesrc display-name=:%d show-pointer=true use-damage=false ! "+
			"video/x-raw,framerate=%d/1 ! videoconvert ! queue max-size-buffers=1 ! "+
			"%s ! appsink name=appsink emit-signals=true sync=false drop=true max-buffers=1",
		b.cfg.Display.DisplayNum,
		b.cfg.Stream.TargetFPS,
		encoder,
	)

	log.Info().Str("codec", b.cfg.Stream.VideoCodec).Str("encoder", b.cfg.Stream.Encoder).Msg("starting video pipeline")
	videoPipeline, err := gst.CreatePipeline(videoPipelineStr)
	if err != nil {
		return err
	}
	b.videoPipeline = videoPipeline

	// Audio Pipeline with queue to prevent drift
	sinkName := fmt.Sprintf("vbrowser-%d", b.cfg.Display.DisplayNum)
	audioPipelineStr := fmt.Sprintf(
		"pulsesrc device=%s.monitor ! audio/x-raw,channels=2 ! audioconvert ! queue max-size-buffers=100 max-size-time=0 max-size-bytes=0 ! opusenc bitrate=128000 ! "+
			"appsink name=appsink emit-signals=true sync=false",
		sinkName,
	)

	audioPipeline, err := gst.CreatePipeline(audioPipelineStr)
	if err != nil {
		videoPipeline.Destroy()
		return err
	}
	b.audioPipeline = audioPipeline

	videoPipeline.Play()
	audioPipeline.Play()
	log.Info().Msg("video and audio pipelines started")

	go func() {
		for {
			select {
			case <-b.done:
				b.StopPipelines()
				return
			case <-b.pipelinesDone:
				return
			case <-ctx.Done():
				b.Stop()
				return
			case sample := <-videoPipeline.Sample():
				_ = b.videoTrack.WriteSample(media.Sample{Data: sample.Data, Duration: sample.Duration})
			case sample := <-audioPipeline.Sample():
				_ = b.audioTrack.WriteSample(media.Sample{Data: sample.Data, Duration: sample.Duration})
			}
		}
	}()

	return nil
}

func (b *Broadcaster) StopPipelines() {
	if b.pipelinesDone != nil {
		select {
		case <-b.pipelinesDone:
		default:
			close(b.pipelinesDone)
		}
	}
	if b.videoPipeline != nil {
		b.videoPipeline.Destroy()
		b.videoPipeline = nil
	}
	if b.audioPipeline != nil {
		b.audioPipeline.Destroy()
		b.audioPipeline = nil
	}
}

func (b *Broadcaster) RestartPipelines(ctx context.Context) {
	b.StopPipelines()
	time.Sleep(500 * time.Millisecond)
	if err := b.StartPipelines(ctx); err != nil {
		log.Error().Err(err).Msg("failed to restart pipelines")
	}
}

func (b *Broadcaster) RequestKeyframe() {
	if b.videoPipeline != nil {
		b.videoPipeline.EmitVideoKeyframe()
	}
}

func (b *Broadcaster) Stop() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.running {
		return
	}
	b.running = false
	close(b.done)
}

func (b *Broadcaster) VideoTrack() *webrtc.TrackLocalStaticSample {
	return b.videoTrack
}

func (b *Broadcaster) AudioTrack() *webrtc.TrackLocalStaticSample {
	return b.audioTrack
}

func (b *Broadcaster) HandlePLI(receiver *webrtc.RTPReceiver) {
	go func() {
		for {
			pkts, _, err := receiver.ReadRTCP()
			if err != nil {
				return
			}
			for _, pkt := range pkts {
				if _, ok := pkt.(*rtcp.PictureLossIndication); ok {
					b.RequestKeyframe()
				}
			}
		}
	}()
}
