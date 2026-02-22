//go:build linux
// +build linux

package capture

import (
	"fmt"
	"image"
	"os"

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
		Msg("initializing Xvfb capturer")

	return &XvfbCapturer{
		displayNum: displayNum,
		width:      width,
		height:     height,
		bounds:     bounds,
	}, nil
}

// Capture captures a single frame from the Xvfb display.
func (c *XvfbCapturer) Capture() (*image.RGBA, error) {
	// Use screenshot library which handles X11 display capture
	img, err := screenshot.CaptureDisplay(0)
	if err != nil {
		return nil, fmt.Errorf("capture display: %w", err)
	}

	// Convert to RGBA
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}

	return rgba, nil
}

// Close releases resources held by the capturer.
func (c *XvfbCapturer) Close() error {
	log.Debug().Msg("closing Xvfb capturer")
	return nil
}
