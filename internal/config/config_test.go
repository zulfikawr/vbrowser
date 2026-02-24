package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaults(t *testing.T) {
	cfg := Defaults()

	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("expected host 127.0.0.1, got %s", cfg.Server.Host)
	}

	if cfg.Server.Port != 7070 {
		t.Errorf("expected port 7070, got %d", cfg.Server.Port)
	}

	if cfg.Browser.WindowWidth != 1280 {
		t.Errorf("expected width 1280, got %d", cfg.Browser.WindowWidth)
	}

	if cfg.Stream.TargetFPS != 60 {
		t.Errorf("expected fps 60, got %d", cfg.Stream.TargetFPS)
	}
}

func TestLoad(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	configJSON := `{
		"server": {
			"host": "0.0.0.0",
			"port": 8080
		},
		"browser": {
			"window_width": 1280,
			"window_height": 720
		}
	}`

	if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("expected host 0.0.0.0, got %s", cfg.Server.Host)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Server.Port)
	}

	if cfg.Browser.WindowWidth != 1280 {
		t.Errorf("expected width 1280, got %d", cfg.Browser.WindowWidth)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(*Config)
		wantErr bool
	}{
		{
			name:    "valid config",
			modify:  func(c *Config) {},
			wantErr: false,
		},
		{
			name: "invalid port low",
			modify: func(c *Config) {
				c.Server.Port = 0
			},
			wantErr: true,
		},
		{
			name: "invalid port high",
			modify: func(c *Config) {
				c.Server.Port = 70000
			},
			wantErr: true,
		},
		{
			name: "invalid width",
			modify: func(c *Config) {
				c.Browser.WindowWidth = 100
			},
			wantErr: true,
		},
		{
			name: "invalid fps",
			modify: func(c *Config) {
				c.Stream.TargetFPS = 0
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Defaults()
			tt.modify(cfg)
			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExpandPath(t *testing.T) {
	homeDir, _ := os.UserHomeDir()

	tests := []struct {
		input    string
		expected string
	}{
		{"~/test", filepath.Join(homeDir, "test")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
	}

	for _, tt := range tests {
		result := expandPath(tt.input)
		if result != tt.expected {
			t.Errorf("expandPath(%s) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}
