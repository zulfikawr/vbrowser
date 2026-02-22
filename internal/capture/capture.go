// Package capture provides screen capture functionality for vbrowser.
package capture

import "image"

// Capturer defines the interface for screen capture implementations.
type Capturer interface {
	// Capture captures a single frame from the display.
	Capture() (*image.RGBA, error)
	// Close releases any resources held by the capturer.
	Close() error
}
