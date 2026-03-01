// Package server provides the HTTP server for vbrowser.
package server

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/zulfikawr/vbrowser/internal/browser"
	"github.com/zulfikawr/vbrowser/internal/config"
	"github.com/zulfikawr/vbrowser/internal/stream"
	"github.com/zulfikawr/vbrowser/pkg/xorg"
)

//go:embed ui
var uiFiles embed.FS

// Server manages the HTTP server.
type Server struct {
	cfg            *config.Config
	configPath     string
	server         *http.Server
	mgr            *browser.Manager
	currentSession *stream.Session
	inputBatcher   *InputBatcher
	mu             sync.Mutex
	wsWriteMu      sync.Mutex
}

// New creates a new HTTP server.
func New(cfg *config.Config, mgr *browser.Manager, configPath string) *Server {
	s := &Server{
		cfg:        cfg,
		configPath: configPath,
		mgr:        mgr,
	}
	s.inputBatcher = NewInputBatcher(func(x, y int) {
		xorg.Move(x, y)
	})
	return s
}

// Start starts the HTTP server.
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Serve embedded UI
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/login", s.handleLogin)
	mux.HandleFunc("/client.js", s.handleClientJS)
	mux.HandleFunc("/guacamole-keyboard.js", s.handleGuacamoleJS)
	mux.HandleFunc("/health", s.handleHealth)

	// WebSocket signaling endpoint
	mux.HandleFunc("/ws", s.handleWebSocket)

	addr := fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port)
	s.server = &http.Server{
		Addr:         addr,
		Handler:      s.authMiddleware(s.logMiddleware(mux)),
		ReadTimeout:  1 * time.Hour,
		WriteTimeout: 1 * time.Hour,
		IdleTimeout:  1 * time.Hour,
	}

	log.Info().Str("addr", addr).Msg("HTTP server starting")

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("HTTP server error")
		}
	}()

	return nil
}

// Stop gracefully shuts down the HTTP server.
func (s *Server) Stop(ctx context.Context) error {
	if s.server == nil {
		return nil
	}

	log.Info().Msg("stopping HTTP server")
	return s.server.Shutdown(ctx)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	data, err := uiFiles.ReadFile("ui/index.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Error().Err(err).Msg("failed to read index.html")
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0")
	if _, err := w.Write(data); err != nil {
		log.Error().Err(err).Msg("failed to write response")
	}
}

func (s *Server) handleClientJS(w http.ResponseWriter, r *http.Request) {
	data, err := uiFiles.ReadFile("ui/client.js")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Error().Err(err).Msg("failed to read client.js")
		return
	}

	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0")
	if _, err := w.Write(data); err != nil {
		log.Error().Err(err).Msg("failed to write response")
	}
}

func (s *Server) handleGuacamoleJS(w http.ResponseWriter, r *http.Request) {
	data, err := uiFiles.ReadFile("ui/guacamole-keyboard.js")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Error().Err(err).Msg("failed to read guacamole-keyboard.js")
		return
	}

	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0")
	if _, err := w.Write(data); err != nil {
		log.Error().Err(err).Msg("failed to write response")
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		data, err := uiFiles.ReadFile("ui/login.html")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(data)
		return
	}

	if r.Method == http.MethodPost {
		password := r.FormValue("password")
		if password == s.cfg.Server.Auth.Token {
			http.SetCookie(w, &http.Cookie{
				Name:     "vbrowser_auth",
				Value:    s.cfg.Server.Auth.Token,
				Path:     "/",
				HttpOnly: true,
				MaxAge:   86400 * 30, // 30 days
			})
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, "/login?error=1", http.StatusSeeOther)
	}
}

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.cfg.Server.Auth.Enabled || r.URL.Path == "/login" || r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		cookie, err := r.Cookie("vbrowser_auth")
		if err != nil || cookie.Value != s.cfg.Server.Auth.Token {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Debug().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("remote", r.RemoteAddr).
			Dur("duration", time.Since(start)).
			Msg("HTTP request")
	})
}
