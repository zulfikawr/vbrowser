// Package stream provides video streaming functionality for vbrowser.
package stream

import (
	"image"

	"github.com/rs/zerolog/log"
)

// Encoder handles video encoding (placeholder for VP8).
type Encoder struct {
	width   int
	height  int
	fps     int
	bitrate int
}

// NewEncoder creates a new video encoder.
func NewEncoder(width, height, fps, bitrate int) (*Encoder, error) {
	log.Info().
		Int("width", width).
		Int("height", height).
		Int("fps", fps).
		Int("bitrate_kbps", bitrate).
		Msg("initializing encoder")

	return &Encoder{
		width:   width,
		height:  height,
		fps:     fps,
		bitrate: bitrate,
	}, nil
}

// Encode encodes a single frame (placeholder - returns raw frame data for now).
func (e *Encoder) Encode(frame *image.RGBA) ([]byte, error) {
	// For now, return raw pixel data
	// In Phase 2.3, this will be replaced with actual VP8 encoding via WebRTC track
	return frame.Pix, nil
}

// Close releases encoder resources.
func (e *Encoder) Close() error {
	log.Debug().Msg("encoder closed")
	return nil
}
