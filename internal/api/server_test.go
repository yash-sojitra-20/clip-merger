package api

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yash-sojitra-20/clip-merger/internal/config"
)

func newTestServer() *Server {
	cfg := &config.Config{
		ServiceName:     "clip-merger",
		ServiceVersion: "0.1.0",
		AppEnv:          "testing",
		Host:            "127.0.0.1",
		Port:            0,
		ShutdownTimeout: 1 * time.Second,
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewServer(cfg, logger)
}

func TestServerRoutes(t *testing.T) {
	srv := newTestServer()
	handler := srv.Handler()

	routes := []struct {
		method         string
		path           string
		expectedStatus int
		expectJSON     bool
	}{
		{"GET", "/health", http.StatusOK, true},
		{"GET", "/healthz", http.StatusOK, true},
		{"GET", "/livez", http.StatusOK, true},
		{"GET", "/readyz", http.StatusOK, true},
		{"GET", "/", http.StatusOK, true},
		{"GET", "/non-existent", http.StatusNotFound, false},
	}

	for _, rt := range routes {
		t.Run(rt.method+" "+rt.path, func(t *testing.T) {
			req := httptest.NewRequest(rt.method, rt.path, nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != rt.expectedStatus {
				t.Errorf("expected status %d, got %d", rt.expectedStatus, rr.Code)
			}

			if rt.expectJSON {
				var data map[string]any
				if err := json.NewDecoder(rr.Body).Decode(&data); err != nil {
					t.Fatalf("expected valid JSON response: %v", err)
				}
				if rt.path == "/health" || rt.path == "/healthz" || rt.path == "/" {
					if data["status"] != "ok" {
						t.Errorf("expected status 'ok', got %v", data["status"])
					}
					if data["service"] != "clip-merger" {
						t.Errorf("expected service 'clip-merger', got %v", data["service"])
					}
				}
			}
		})
	}
}

func TestServerGracefulShutdown(t *testing.T) {
	cfg := &config.Config{
		ServiceName:     "clip-merger",
		ServiceVersion: "0.1.0",
		AppEnv:          "testing",
		Host:            "127.0.0.1",
		Port:            18080,
		ShutdownTimeout: 1 * time.Second,
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := NewServer(cfg, logger)

	ctx, cancel := context.WithCancel(context.Background())

	errChan := make(chan error, 1)
	go func() {
		errChan <- srv.Start(ctx)
	}()

	// Allow server time to bind and listen
	time.Sleep(100 * time.Millisecond)

	// Trigger shutdown
	cancel()

	select {
	case err := <-errChan:
		if err != nil {
			t.Fatalf("expected clean shutdown, got error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server shutdown timed out")
	}
}
