//go:build windows
// +build windows

package capture

import (
	"fmt"
	"image"
)

// WindowsCapturer is a stub for Windows screen capture.
type WindowsCapturer struct{}

// NewWindowsCapturer creates a new Windows screen capturer (stub).
func NewWindowsCapturer(width, height int) (*WindowsCapturer, error) {
	return nil, fmt.Errorf("Windows screen capture not yet implemented")
}

// Capture is a stub.
func (c *WindowsCapturer) Capture() (*image.RGBA, error) {
	return nil, fmt.Errorf("not implemented")
}

// Close is a stub.
func (c *WindowsCapturer) Close() error {
	return nil
}
