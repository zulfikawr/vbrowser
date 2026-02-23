package server

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/zulfikawr/vbrowser/internal/config"
)

func TestNew(t *testing.T) {
	cfg := config.Defaults()
	srv := New(cfg, nil)

	if srv == nil {
		t.Fatal("New returned nil")
	}

	if srv.cfg != cfg {
		t.Error("config not set correctly")
	}
}

func TestStartStop(t *testing.T) {
	cfg := config.Defaults()
	cfg.Server.Port = 17070 // Use different port for testing
	srv := New(cfg, nil)

	if err := srv.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Test that server is responding
	resp, err := http.Get("http://127.0.0.1:17070/")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Stop server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Stop(ctx); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
}

func TestHandleIndex(t *testing.T) {
	cfg := config.Defaults()
	cfg.Server.Port = 17071
	srv := New(cfg, nil)

	if err := srv.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer func() {
		if err := srv.Stop(context.Background()); err != nil {
			t.Logf("Stop error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	resp, err := http.Get("http://127.0.0.1:17071/")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("expected text/html content type, got %s", contentType)
	}
}

func TestHandle404(t *testing.T) {
	cfg := config.Defaults()
	cfg.Server.Port = 17072
	srv := New(cfg, nil)

	if err := srv.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer func() {
		if err := srv.Stop(context.Background()); err != nil {
			t.Logf("Stop error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	resp, err := http.Get("http://127.0.0.1:17072/nonexistent")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", resp.StatusCode)
	}
}
