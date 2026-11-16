package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/yash-sojitra-20/clip-merger/internal/api/handlers"
	"github.com/yash-sojitra-20/clip-merger/internal/api/middleware"
	"github.com/yash-sojitra-20/clip-merger/internal/config"
)

// Server wraps the HTTP server and routing.
type Server struct {
	cfg        *config.Config
	logger     *slog.Logger
	httpServer *http.Server
}

// NewServer configures a new Server instance.
func NewServer(cfg *config.Config, logger *slog.Logger) *Server {
	mux := http.NewServeMux()

	healthHandler := handlers.NewHealthHandler(cfg.ServiceName, cfg.ServiceVersion, cfg.AppEnv)

	// Register Health & Probe routes
	mux.HandleFunc("GET /health", healthHandler.Health)
	mux.HandleFunc("GET /healthz", healthHandler.Health)
	mux.HandleFunc("GET /livez", healthHandler.Liveness)
	mux.HandleFunc("GET /readyz", healthHandler.Readiness)

	// Root fallback / info
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		healthHandler.Health(w, r)
	})

	// Apply Middlewares
	wrappedHandler := middleware.Chain(
		middleware.CORS,
		middleware.Logging(logger),
		middleware.Recovery(logger),
	)(mux)

	httpServer := &http.Server{
		Addr:         cfg.Address(),
		Handler:      wrappedHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		cfg:        cfg,
		logger:     logger,
		httpServer: httpServer,
	}
}

// Handler returns the configured HTTP handler (useful for testing).
func (s *Server) Handler() http.Handler {
	return s.httpServer.Handler
}

// Start runs the HTTP server and blocks until the context is canceled or an error occurs.
func (s *Server) Start(ctx context.Context) error {
	errChan := make(chan error, 1)

	go func() {
		s.logger.Info("Starting HTTP server",
			slog.String("address", s.cfg.Address()),
			slog.String("service", s.cfg.ServiceName),
			slog.String("version", s.cfg.ServiceVersion),
			slog.String("env", s.cfg.AppEnv),
		)

		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- fmt.Errorf("http server failed: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
		s.logger.Info("Shutdown signal received, draining active connections...",
			slog.Duration("timeout", s.cfg.ShutdownTimeout),
		)

		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
		defer cancel()

		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("server graceful shutdown failed: %w", err)
		}

		s.logger.Info("HTTP server stopped gracefully")
		return nil

	case err := <-errChan:
		return err
	}
}
