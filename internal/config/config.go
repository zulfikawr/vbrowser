// Package config provides configuration management for vbrowser.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config represents the complete vbrowser configuration.
type Config struct {
	Server  ServerConfig  `json:"server"`
	Browser BrowserConfig `json:"browser"`
	Display DisplayConfig `json:"display"`
	Stream  StreamConfig  `json:"stream"`
	Session SessionConfig `json:"session"`
	Logging LoggingConfig `json:"logging"`
	PIDFile string        `json:"pid_file"`
}

// ServerConfig contains HTTP server settings.
type ServerConfig struct {
	Host string     `json:"host"`
	Port int        `json:"port"`
	Auth AuthConfig `json:"auth"`
}

// AuthConfig contains authentication settings.
type AuthConfig struct {
	Enabled bool   `json:"enabled"`
	Token   string `json:"token"`
}

// BrowserConfig contains Chromium browser settings.
type BrowserConfig struct {
	ChromiumPath string   `json:"chromium_path"`
	AutoDownload bool     `json:"auto_download"`
	DownloadDir  string   `json:"download_dir"`
	ProfileDir   string   `json:"profile_dir"`
	WindowWidth  int      `json:"window_width"`
	WindowHeight int      `json:"window_height"`
	ExtraArgs    []string `json:"extra_args"`
}

// DisplayConfig contains virtual display settings.
type DisplayConfig struct {
	VirtualDisplay bool `json:"virtual_display"`
	DisplayNum     int  `json:"display_num"`
	Depth          int  `json:"depth"`
}

// StreamConfig contains video streaming settings.
type StreamConfig struct {
	VideoCodec     string `json:"video_codec"`
	TargetFPS      int    `json:"target_fps"`
	MaxBitrateKbps int    `json:"max_bitrate_kbps"`
	QualityPreset  string `json:"quality_preset"`
}

// SessionConfig contains session management settings.
type SessionConfig struct {
	MaxSessions        int `json:"max_sessions"`
	IdleTimeoutSeconds int `json:"idle_timeout_seconds"`
}

// LoggingConfig contains logging settings.
type LoggingConfig struct {
	Level string `json:"level"`
	File  string `json:"file"`
}

// Defaults returns a Config with sensible default values.
func Defaults() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		Server: ServerConfig{
			Host: "127.0.0.1",
			Port: 7070,
			Auth: AuthConfig{
				Enabled: false,
				Token:   "",
			},
		},
		Browser: BrowserConfig{
			ChromiumPath: "",
			AutoDownload: true,
			DownloadDir:  filepath.Join(homeDir, ".local/share/vbrowser/chromium"),
			ProfileDir:   filepath.Join(homeDir, ".local/share/vbrowser/profile"),
			WindowWidth:  1920,
			WindowHeight: 1080,
			ExtraArgs:    []string{},
		},
		Display: DisplayConfig{
			VirtualDisplay: true,
			DisplayNum:     99,
			Depth:          24,
		},
		Stream: StreamConfig{
			VideoCodec:     "vp8",
			TargetFPS:      30,
			MaxBitrateKbps: 4000,
			QualityPreset:  "balanced",
		},
		Session: SessionConfig{
			MaxSessions:        1,
			IdleTimeoutSeconds: 0,
		},
		Logging: LoggingConfig{
			Level: "info",
			File:  "",
		},
		PIDFile: filepath.Join(homeDir, ".local/share/vbrowser/vbrowser.pid"),
	}
}

// Load reads and parses the configuration file from the given path.
// If path is empty, returns default configuration.
// Expands ~ in paths and applies validation.
func Load(path string) (*Config, error) {
	cfg := Defaults()

	if path == "" {
		return cfg, nil
	}

	path = expandPath(path)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	cfg.Browser.DownloadDir = expandPath(cfg.Browser.DownloadDir)
	cfg.Browser.ProfileDir = expandPath(cfg.Browser.ProfileDir)
	cfg.PIDFile = expandPath(cfg.PIDFile)
	if cfg.Logging.File != "" {
		cfg.Logging.File = expandPath(cfg.Logging.File)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks if the configuration values are valid and returns an error if not.
func (c *Config) Validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("server.port must be between 1 and 65535, got %d", c.Server.Port)
	}

	if c.Browser.WindowWidth < 640 || c.Browser.WindowWidth > 7680 {
		return fmt.Errorf("browser.window_width must be between 640 and 7680, got %d", c.Browser.WindowWidth)
	}

	if c.Browser.WindowHeight < 480 || c.Browser.WindowHeight > 4320 {
		return fmt.Errorf("browser.window_height must be between 480 and 4320, got %d", c.Browser.WindowHeight)
	}

	if c.Stream.TargetFPS < 1 || c.Stream.TargetFPS > 60 {
		return fmt.Errorf("stream.target_fps must be between 1 and 60, got %d", c.Stream.TargetFPS)
	}

	if c.Stream.MaxBitrateKbps < 100 || c.Stream.MaxBitrateKbps > 50000 {
		return fmt.Errorf("stream.max_bitrate_kbps must be between 100 and 50000, got %d", c.Stream.MaxBitrateKbps)
	}

	if c.Display.DisplayNum < 0 || c.Display.DisplayNum > 999 {
		return fmt.Errorf("display.display_num must be between 0 and 999, got %d", c.Display.DisplayNum)
	}

	return nil
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		homeDir, _ := os.UserHomeDir()
		return filepath.Join(homeDir, path[2:])
	}
	return path
}

// ApplyEnv applies environment variable overrides to the configuration.
func (c *Config) ApplyEnv() {
	if port := os.Getenv("VBROWSER_PORT"); port != "" {
		var p int
		if _, err := fmt.Sscanf(port, "%d", &p); err == nil {
			c.Server.Port = p
		}
	}
	if level := os.Getenv("VBROWSER_LOG_LEVEL"); level != "" {
		c.Logging.Level = level
	}
}
