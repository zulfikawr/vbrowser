package stream

import (
	"context"
	"image"
	"testing"
	"time"

	"github.com/pion/webrtc/v4"
	"github.com/zulfikawr/vbrowser/internal/config"
)

// mockCapturer is a mock implementation of capture.Capturer for testing.
type mockCapturer struct {
	width  int
	height int
}

func (m *mockCapturer) Capture() (*image.RGBA, error) {
	return image.NewRGBA(image.Rect(0, 0, m.width, m.height)), nil
}

func (m *mockCapturer) Close() error {
	return nil
}

func TestNewSession(t *testing.T) {
	cfg := config.Defaults()
	capturer := &mockCapturer{width: 640, height: 480}

	session, err := NewSession("test-session", cfg, capturer)
	if err != nil {
		t.Fatalf("NewSession failed: %v", err)
	}
	defer func() {
		if err := session.Stop(); err != nil {
			t.Logf("Stop error: %v", err)
		}
	}()

	if session == nil {
		t.Fatal("session is nil")
	}

	if session.id != "test-session" {
		t.Errorf("expected id 'test-session', got %s", session.id)
	}

	if session.peerConnection == nil {
		t.Error("peer connection is nil")
	}

	if session.videoTrack == nil {
		t.Error("video track is nil")
	}
}

func TestCreateOffer(t *testing.T) {
	cfg := config.Defaults()
	capturer := &mockCapturer{width: 640, height: 480}

	session, err := NewSession("test-session", cfg, capturer)
	if err != nil {
		t.Fatalf("NewSession failed: %v", err)
	}
	defer func() {
		if err := session.Stop(); err != nil {
			t.Logf("Stop error: %v", err)
		}
	}()

	offer, err := session.CreateOffer()
	if err != nil {
		t.Fatalf("CreateOffer failed: %v", err)
	}

	if offer.Type != webrtc.SDPTypeOffer {
		t.Errorf("expected offer type, got %s", offer.Type)
	}

	if offer.SDP == "" {
		t.Error("offer SDP is empty")
	}
}

func TestStartStop(t *testing.T) {
	cfg := config.Defaults()
	capturer := &mockCapturer{width: 640, height: 480}

	session, err := NewSession("test-session", cfg, capturer)
	if err != nil {
		t.Fatalf("NewSession failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	if err := session.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	if err := session.Stop(); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
}
