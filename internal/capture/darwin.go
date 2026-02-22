//go:build darwin
// +build darwin

package capture

import (
	"fmt"
	"image"
)

// MacOSCapturer is a stub for macOS screen capture.
type MacOSCapturer struct{}

// NewMacOSCapturer creates a new macOS screen capturer (stub).
func NewMacOSCapturer(width, height int) (*MacOSCapturer, error) {
	return nil, fmt.Errorf("macOS screen capture not yet implemented")
}

// Capture is a stub.
func (c *MacOSCapturer) Capture() (*image.RGBA, error) {
	return nil, fmt.Errorf("not implemented")
}

// Close is a stub.
func (c *MacOSCapturer) Close() error {
	return nil
}
