package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestConfigLoadDefaults(t *testing.T) {
	// Clear relevant env vars
	keys := []string{"PORT", "HOST", "APP_ENV", "LOG_LEVEL", "LOG_FORMAT", "SHUTDOWN_TIMEOUT", "SERVICE_NAME", "SERVICE_VERSION"}
	for _, k := range keys {
		orig := os.Getenv(k)
		_ = os.Unsetenv(k)
		defer func(key, val string) {
			if val != "" {
				_ = os.Setenv(key, val)
			}
		}(k, orig)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error loading defaults, got: %v", err)
	}

	if cfg.Port != DefaultPort {
		t.Errorf("expected port %d, got %d", DefaultPort, cfg.Port)
	}
	if cfg.Host != DefaultHost {
		t.Errorf("expected host %s, got %s", DefaultHost, cfg.Host)
	}
	if cfg.AppEnv != DefaultAppEnv {
		t.Errorf("expected app env %s, got %s", DefaultAppEnv, cfg.AppEnv)
	}
	if cfg.ShutdownTimeout != DefaultShutdownTimeout {
		t.Errorf("expected shutdown timeout %v, got %v", DefaultShutdownTimeout, cfg.ShutdownTimeout)
	}
	if cfg.Address() != "0.0.0.0:8080" {
		t.Errorf("expected address 0.0.0.0:8080, got %s", cfg.Address())
	}
}

func TestConfigLoadCustomEnv(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("HOST", "127.0.0.1")
	t.Setenv("APP_ENV", "production")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("LOG_FORMAT", "json")
	t.Setenv("SHUTDOWN_TIMEOUT", "15s")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error loading custom config, got: %v", err)
	}

	if cfg.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Port)
	}
	if cfg.Host != "127.0.0.1" {
		t.Errorf("expected host 127.0.0.1, got %s", cfg.Host)
	}
	if cfg.AppEnv != "production" {
		t.Errorf("expected app env production, got %s", cfg.AppEnv)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("expected log level debug, got %s", cfg.LogLevel)
	}
	if cfg.LogFormat != "json" {
		t.Errorf("expected log format json, got %s", cfg.LogFormat)
	}
	if cfg.ShutdownTimeout != 15*time.Second {
		t.Errorf("expected shutdown timeout 15s, got %v", cfg.ShutdownTimeout)
	}
	if cfg.Address() != "127.0.0.1:9090" {
		t.Errorf("expected address 127.0.0.1:9090, got %s", cfg.Address())
	}
}

func TestConfigInvalidPort(t *testing.T) {
	t.Setenv("PORT", "not-a-number")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid PORT, got nil")
	}

	t.Setenv("PORT", "-1")
	_, err = Load()
	if err == nil {
		t.Fatal("expected error for negative PORT, got nil")
	}
}

func TestConfigInvalidShutdownTimeout(t *testing.T) {
	t.Setenv("PORT", "8080")
	t.Setenv("SHUTDOWN_TIMEOUT", "invalid-duration")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid SHUTDOWN_TIMEOUT, got nil")
	}
}

func TestLoadDotEnv(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")
	content := `# Comment line
TEST_DOTENV_KEY=sample_value
TEST_DOTENV_QUOTED="quoted_value"
`
	if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp .env file: %v", err)
	}

	loadDotEnv(envPath)

	if val := os.Getenv("TEST_DOTENV_KEY"); val != "sample_value" {
		t.Errorf("expected 'sample_value', got %q", val)
	}
	if val := os.Getenv("TEST_DOTENV_QUOTED"); val != "quoted_value" {
		t.Errorf("expected 'quoted_value', got %q", val)
	}
}

