package middleware

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLoggingMiddleware(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	handler := Logging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("created"))
	}))

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}

	logOutput := buf.String()
	if !strings.Contains(logOutput, "http request") {
		t.Errorf("expected log to contain 'http request', got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "status=201") {
		t.Errorf("expected log to contain 'status=201', got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "path=/test") {
		t.Errorf("expected log to contain 'path=/test', got: %s", logOutput)
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	panicHandler := Recovery(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("unexpected explosion")
	}))

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rr := httptest.NewRecorder()

	// Ensure test doesn't crash from panic
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic was not caught by Recovery middleware: %v", r)
		}
	}()

	panicHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Internal Server Error") {
		t.Errorf("expected response body to contain 'Internal Server Error', got %q", rr.Body.String())
	}
}

func TestCORSMiddleware(t *testing.T) {
	handler := CORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))

	// Test regular request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected Access-Control-Allow-Origin: *")
	}

	// Test OPTIONS preflight
	optReq := httptest.NewRequest(http.MethodOptions, "/test", nil)
	optRR := httptest.NewRecorder()
	handler.ServeHTTP(optRR, optReq)

	if optRR.Code != http.StatusNoContent {
		t.Errorf("expected OPTIONS response 204 No Content, got %d", optRR.Code)
	}
}

func TestChainMiddleware(t *testing.T) {
	var order []string

	m1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "m1_start")
			next.ServeHTTP(w, r)
			order = append(order, "m1_end")
		})
	}

	m2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "m2_start")
			next.ServeHTTP(w, r)
			order = append(order, "m2_end")
		})
	}

	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
		w.WriteHeader(http.StatusOK)
	})

	chained := Chain(m1, m2)(final)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	chained.ServeHTTP(rr, req)

	expected := []string{"m1_start", "m2_start", "handler", "m2_end", "m1_end"}
	if len(order) != len(expected) {
		t.Fatalf("expected %d elements, got %d: %v", len(expected), len(order), order)
	}
	for i, v := range expected {
		if order[i] != v {
			t.Errorf("index %d: expected %s, got %s", i, v, order[i])
		}
	}
}

