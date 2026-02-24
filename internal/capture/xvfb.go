//go:build linux
// +build linux

package capture

import (
	"fmt"
	"image"
	"os"

	"golang.org/x/sys/unix"
	"time"

	"github.com/kbinani/screenshot"
	"github.com/rs/zerolog/log"
)

// XvfbCapturer captures frames from an Xvfb display.
type XvfbCapturer struct {
	displayNum int
	width      int
	height     int
	bounds     image.Rectangle
}

// NewXvfbCapturer creates a new Xvfb screen capturer.
func NewXvfbCapturer(displayNum, width, height int) (*XvfbCapturer, error) {
	bounds := image.Rect(0, 0, width, height)

	// Set DISPLAY environment variable for screenshot library
	displayStr := fmt.Sprintf(":%d", displayNum)
	if err := os.Setenv("DISPLAY", displayStr); err != nil {
		return nil, fmt.Errorf("set DISPLAY env: %w", err)
	}

	log.Info().
		Int("display", displayNum).
		Int("width", width).
		Int("height", height).
		Str("DISPLAY", os.Getenv("DISPLAY")).
		Msg("initializing Xvfb capturer")

	// Try to wait for X11 to be ready (best effort, don't fail if not available)
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		n := screenshot.NumActiveDisplays()
		if n > 0 {
			log.Debug().Int("displays", n).Msg("X11 display ready")
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	return &XvfbCapturer{
		displayNum: displayNum,
		width:      width,
		height:     height,
		bounds:     bounds,
	}, nil
}

// Capture captures a single frame from the Xvfb display.
func (c *XvfbCapturer) Capture() (*image.RGBA, error) {
	// Suppress XGB warnings by redirecting stderr to /dev/null at fd level
	devNull, _ := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if devNull != nil {
		oldStderr, _ := unix.Dup(2)
		_ = unix.Dup2(int(devNull.Fd()), 2)
		defer func() {
			_ = unix.Dup2(oldStderr, 2)
			_ = unix.Close(oldStderr)
			_ = devNull.Close()
		}()
	}

	// Use screenshot library which handles X11 display capture
	img, err := screenshot.CaptureDisplay(0)
	if err != nil {
		return nil, fmt.Errorf("capture display: %w", err)
	}

	// Convert to RGBA
	bounds := img.Bounds()
	rgba := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			rgba.Set(x, y, img.At(bounds.Min.X+x, bounds.Min.Y+y))
		}
	}

	return rgba, nil
}

// Close releases resources held by the capturer.
func (c *XvfbCapturer) Close() error {
	log.Debug().Msg("closing Xvfb capturer")
	return nil
}
