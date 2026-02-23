package stream

import (
	"fmt"
	"image"
	"io"
	"os"
	"os/exec"

	"github.com/rs/zerolog/log"
)

// Encoder handles video encoding via FFmpeg.
type Encoder struct {
	width   int
	height  int
	fps     int
	bitrate int
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
}

// NewEncoder creates a new video encoder using FFmpeg.
func NewEncoder(width, height, fps, bitrate int) (*Encoder, error) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return nil, fmt.Errorf("ffmpeg not found in PATH")
	}

	// FFmpeg command for VP8 encoding from raw RGBA stream
	// -f rawvideo: input format
	// -pixel_format rgba: input pixel format
	// -video_size: input resolution
	// -i pipe:0: read from stdin
	// -c:v libvpx: use VP8 encoder
	// -b:v: bitrate
	// -deadline realtime: low latency encoding
	// -f ivf: output format (simplest for WebRTC packetization)
	args := []string{
		"-loglevel", "error", // Hide noise, show only errors
		"-f", "rawvideo",
		"-pixel_format", "rgba",
		"-video_size", fmt.Sprintf("%dx%d", width, height),
		"-framerate", fmt.Sprintf("%d", fps),
		"-i", "pipe:0",
		"-c:v", "libvpx",
		"-pix_fmt", "yuv420p",
		"-b:v", fmt.Sprintf("%dk", bitrate),
		"-deadline", "realtime",
		"-cpu-used", "5",
		"-threads", "4",
		"-g", "30",
		"-keyint_min", "30",
		"-f", "ivf",
		"-deadline", "realtime",
		"-cpu-used", "8",
		"pipe:1",
	}

	cmd := exec.Command("ffmpeg", args...)

	// Pipe FFmpeg stderr to our process stderr for debugging
	stderr, err := cmd.StderrPipe()
	if err == nil {
		go func() {
			io.Copy(os.Stderr, stderr)
		}()
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("ffmpeg start: %w", err)
	}

	log.Info().
		Int("width", width).
		Int("height", height).
		Int("fps", fps).
		Int("bitrate_kbps", bitrate).
		Msg("FFmpeg encoder started")

	return &Encoder{
		width:   width,
		height:  height,
		fps:     fps,
		bitrate: bitrate,
		cmd:     cmd,
		stdin:   stdin,
		stdout:  stdout,
	}, nil
}

// Encode writes a frame to FFmpeg's stdin.
func (e *Encoder) Encode(frame *image.RGBA) error {
	_, err := e.stdin.Write(frame.Pix)
	return err
}

// Read reads an encoded packet from FFmpeg's stdout.
func (e *Encoder) Read(p []byte) (int, error) {
	return e.stdout.Read(p)
}

// Close releases encoder resources.
func (e *Encoder) Close() error {
	log.Debug().Msg("closing FFmpeg encoder")
	e.stdin.Close()
	if e.cmd != nil && e.cmd.Process != nil {
		e.cmd.Process.Kill()
	}
	return nil
}
