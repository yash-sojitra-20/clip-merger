# Clip Merger

An automated video processing pipeline that combines, ranks, edits, and publishes videos using **Go**, **FFmpeg**, **Google Drive**, and the **YouTube Data API**.

---

## Overview

Clip Merger automates the workflow of creating and publishing combined social media videos:
1. **Cloud Ingestion**: Fetches raw video clips from daily directories in Google Drive.
2. **Video Processing**: Integrates with FFmpeg to rank, concatenate, and apply dynamic styled text overlays (custom colors, typography, background boxes).
3. **Social Publishing**: Publishes final compiled videos to YouTube (Shorts/Videos) using the YouTube Data API.
4. **Automation & Orchestration**: Scheduled via GitHub Actions and exposed via clean Go backend API endpoints.

---

## Architecture & Layout

```
clip-merger/
├── cmd/
│   └── clip-merger/
│       └── main.go                  # Service entrypoint & graceful shutdown
├── internal/
│   ├── api/
│   │   ├── handlers/
│   │   │   ├── health.go            # Health check, readiness & liveness handlers
│   │   │   └── health_test.go
│   │   ├── middleware/
│   │   │   ├── middleware.go        # Logging, recovery & CORS middlewares
│   │   │   └── middleware_test.go
│   │   ├── server.go                # HTTP Server configuration & router
│   │   └── server_test.go
│   ├── config/
│   │   ├── config.go                # Environment & .env configuration
│   │   └── config_test.go
│   └── observability/
│       ├── logger.go                # Structured slog logger setup
│       └── logger_test.go
├── .env.example                     # Sample environment configuration
├── .gitignore                       # Standard Go & video artifact ignores
├── go.mod                           # Go module definition
├── Makefile                         # Build, run, and test targets
└── README.md
```

---

## Prerequisites

- **Go 1.22+** (Go 1.26 installed)
- **Make** (optional, for convenience targets)

---

## Getting Started

### 1. Configuration

Copy the example environment configuration:

```bash
cp .env.example .env
```

Available environment variables:

| Variable | Description | Default |
|---|---|---|
| `HOST` | Bind address | `0.0.0.0` |
| `PORT` | Listening port | `8080` |
| `APP_ENV` | Environment (`development`, `staging`, `production`) | `development` |
| `LOG_LEVEL` | Log level (`debug`, `info`, `warn`, `error`) | `info` |
| `LOG_FORMAT` | Log format (`text`, `json`) | `text` |
| `SHUTDOWN_TIMEOUT` | Graceful shutdown timeout | `10s` |

### 2. Run the Application

Using `make`:
```bash
make run
```

Or using `go run`:
```bash
go run ./cmd/clip-merger
```

CLI flags can also override settings:
```bash
go run ./cmd/clip-merger -port 9090 -log-format json -env production
```

### 3. Build

```bash
make build
# Produces bin/clip-merger
./bin/clip-merger
```

### 4. Run Tests

```bash
make test
# or verbose with race detector:
make test-v
```

---

## API Endpoints

### Health Check

- **URL**: `GET /health` or `GET /healthz` (also mounted at root `GET /`)
- **Description**: Returns the operational status and uptime of the Clip Merger service.
- **Response**: `200 OK`

```json
{
  "status": "ok",
  "service": "clip-merger",
  "version": "0.1.0",
  "environment": "development",
  "timestamp": "2026-11-06T11:34:20Z",
  "uptime": "12s",
  "checks": {
    "api": "healthy"
  }
}
```

### Liveness Probe

- **URL**: `GET /livez`
- **Description**: Kubernetes / orchestrator liveness check.
- **Response**: `200 OK`

```json
{
  "status": "alive"
}
```

### Readiness Probe

- **URL**: `GET /readyz`
- **Description**: Kubernetes / orchestrator readiness check.
- **Response**: `200 OK`

```json
{
  "status": "ready"
}
```

---

## Incremental Roadmap

- [x] **Phase 1: Project Initialization & Health API**
  - Go module initialization
  - Structured logging (`log/slog`)
  - Middleware (structured request logging, panic recovery, CORS)
  - Health check (`/health`, `/healthz`, `/livez`, `/readyz`)
  - Graceful shutdown handling
  - Makefile and unit tests
- [ ] **Phase 2: Storage Integration**
  - Google Drive client to fetch daily source video batches
- [ ] **Phase 3: FFmpeg Pipeline**
  - Video ranking, concatenation, and dynamic text overlays with background boxes
- [ ] **Phase 4: Social Media Publishing**
  - YouTube Data API upload and metadata automation
- [ ] **Phase 5: Automation & CI/CD**
  - GitHub Actions scheduler and automated workflow execution
