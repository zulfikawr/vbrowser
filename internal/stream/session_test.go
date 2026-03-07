package stream

import (
	"context"
	"testing"
	"time"

	"github.com/pion/webrtc/v4"
	"github.com/zulfikawr/vbrowser/internal/config"
)

func TestNewSession(t *testing.T) {
	cfg := config.Defaults()

	// Just mock the broadcaster for tests
	b, err := NewBroadcaster(cfg)
	if err != nil {
		t.Fatalf("NewBroadcaster failed: %v", err)
	}

	session, err := NewSession(cfg, b, "test-id")
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

	if session.peerConnection == nil {
		t.Error("peer connection is nil")
	}
}

func TestCreateAnswer(t *testing.T) {
	cfg := config.Defaults()

	b, err := NewBroadcaster(cfg)
	if err != nil {
		t.Fatalf("NewBroadcaster failed: %v", err)
	}

	session, err := NewSession(cfg, b, "test-id")
	if err != nil {
		t.Fatalf("NewSession failed: %v", err)
	}
	defer func() {
		if err := session.Stop(); err != nil {
			t.Logf("Stop error: %v", err)
		}
	}()

	err = session.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  "v=0\r\no=- 0 0 IN IP4 127.0.0.1\r\ns=-\r\nt=0 0\r\na=fingerprint:sha-256 00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00\r\na=setup:actpass\r\nm=video 9 UDP/TLS/RTP/SAVPF 96\r\na=mid:0\r\na=rtpmaps:96 VP8/90000\r\n",
	})
	if err != nil {
		t.Logf("SetRemoteDescription expected error (invalid SDP) or success: %v", err)
	}

	answer, err := session.CreateAnswer()
	if err != nil {
		t.Fatalf("CreateAnswer failed: %v", err)
	}

	if answer.Type != webrtc.SDPTypeAnswer {
		t.Errorf("expected answer type, got %s", answer.Type)
	}
}

func TestBroadcasterStartStop(t *testing.T) {
	cfg := config.Defaults()

	b, err := NewBroadcaster(cfg)
	if err != nil {
		t.Fatalf("NewBroadcaster failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// This will fail if X11/Pulse is not running, but we just check it doesn't panic
	_ = b.Start(ctx)

	time.Sleep(50 * time.Millisecond)
	b.Stop()
}
