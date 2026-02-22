package stream

import (
	"image"
	"image/color"
	"testing"
)

func TestNewEncoder(t *testing.T) {
	encoder, err := NewEncoder(1920, 1080, 30, 4000)
	if err != nil {
		t.Fatalf("NewEncoder failed: %v", err)
	}

	if encoder == nil {
		t.Fatal("encoder is nil")
	}

	if encoder.width != 1920 {
		t.Errorf("expected width 1920, got %d", encoder.width)
	}

	if encoder.height != 1080 {
		t.Errorf("expected height 1080, got %d", encoder.height)
	}

	if encoder.fps != 30 {
		t.Errorf("expected fps 30, got %d", encoder.fps)
	}

	if encoder.bitrate != 4000 {
		t.Errorf("expected bitrate 4000, got %d", encoder.bitrate)
	}
}

func TestEncode(t *testing.T) {
	encoder, err := NewEncoder(640, 480, 30, 1000)
	if err != nil {
		t.Fatalf("NewEncoder failed: %v", err)
	}
	defer encoder.Close()

	// Create a test frame (solid color)
	frame := image.NewRGBA(image.Rect(0, 0, 640, 480))
	for y := 0; y < 480; y++ {
		for x := 0; x < 640; x++ {
			frame.SetRGBA(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}

	encoded, err := encoder.Encode(frame)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if len(encoded) == 0 {
		t.Error("encoded data is empty")
	}

	t.Logf("encoded frame size: %d bytes", len(encoded))
}

func TestClose(t *testing.T) {
	encoder, err := NewEncoder(640, 480, 30, 1000)
	if err != nil {
		t.Fatalf("NewEncoder failed: %v", err)
	}

	if err := encoder.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}
}
