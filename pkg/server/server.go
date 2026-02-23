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
)

//go:embed ui
var uiFiles embed.FS

// Server manages the HTTP server.
type Server struct {
	cfg            *config.Config
	server         *http.Server
	mgr            *browser.Manager
	currentSession *stream.Session
	mu             sync.Mutex
	wsWriteMu      sync.Mutex
}

// New creates a new HTTP server.
func New(cfg *config.Config, mgr *browser.Manager) *Server {
	return &Server{
		cfg: cfg,
		mgr: mgr,
	}
}

// Start starts the HTTP server.
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Serve embedded UI
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/client.js", s.handleClientJS)

	// WebSocket signaling endpoint
	mux.HandleFunc("/ws", s.handleWebSocket)

	addr := fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port)
	s.server = &http.Server{
		Addr:         addr,
		Handler:      s.logMiddleware(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
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
	if _, err := w.Write(data); err != nil {
		log.Error().Err(err).Msg("failed to write response")
	}
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
