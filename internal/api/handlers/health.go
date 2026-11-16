package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

// HealthResponse represents the payload returned by health endpoints.
type HealthResponse struct {
	Status      string            `json:"status"`
	Service     string            `json:"service"`
	Version     string            `json:"version"`
	Environment string            `json:"environment"`
	Timestamp   string            `json:"timestamp"`
	Uptime      string            `json:"uptime"`
	Checks      map[string]string `json:"checks,omitempty"`
}

// HealthHandler manages health and readiness checks.
type HealthHandler struct {
	serviceName string
	version     string
	environment string
	startTime   time.Time
}

// NewHealthHandler creates a new instance of HealthHandler.
func NewHealthHandler(serviceName, version, environment string) *HealthHandler {
	return &HealthHandler{
		serviceName: serviceName,
		version:     version,
		environment: environment,
		startTime:   time.Now(),
	}
}

// Health handles requests to /health and /healthz.
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	uptime := time.Since(h.startTime).Round(time.Second).String()

	resp := HealthResponse{
		Status:      "ok",
		Service:     h.serviceName,
		Version:     h.version,
		Environment: h.environment,
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Uptime:      uptime,
		Checks: map[string]string{
			"api": "healthy",
		},
	}

	writeJSON(w, http.StatusOK, resp)
}

// Liveness handles kubernetes/container liveness probe at /livez.
func (h *HealthHandler) Liveness(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	resp := map[string]string{
		"status": "alive",
	}
	writeJSON(w, http.StatusOK, resp)
}

// Readiness handles kubernetes/container readiness probe at /readyz.
func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	resp := map[string]string{
		"status": "ready",
	}
	writeJSON(w, http.StatusOK, resp)
}

// writeJSON writes an object as JSON to the response writer.
func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}

