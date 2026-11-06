package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds the application configuration.
type Config struct {
	ServiceName     string
	ServiceVersion string
	AppEnv          string
	Host            string
	Port            int
	LogLevel        string
	LogFormat       string
	ShutdownTimeout time.Duration
}

// Default values
const (
	DefaultServiceName     = "clip-merger"
	DefaultServiceVersion = "0.1.0"
	DefaultAppEnv          = "development"
	DefaultHost            = "0.0.0.0"
	DefaultPort            = 8080
	DefaultLogLevel        = "info"
	DefaultLogFormat       = "text"
	DefaultShutdownTimeout = 10 * time.Second
)

// Load reads configuration from environment variables and an optional .env file.
func Load() (*Config, error) {
	// Best-effort load from .env file in current directory or parent
	loadDotEnv(".env")

	cfg := &Config{
		ServiceName:     getEnv("SERVICE_NAME", DefaultServiceName),
		ServiceVersion: getEnv("SERVICE_VERSION", DefaultServiceVersion),
		AppEnv:          getEnv("APP_ENV", DefaultAppEnv),
		Host:            getEnv("HOST", DefaultHost),
		LogLevel:        getEnv("LOG_LEVEL", DefaultLogLevel),
		LogFormat:       getEnv("LOG_FORMAT", DefaultLogFormat),
	}

	portStr := getEnv("PORT", strconv.Itoa(DefaultPort))
	port, err := strconv.Atoi(portStr)
	if err != nil || port <= 0 || port > 65535 {
		return nil, fmt.Errorf("invalid PORT value %q: must be a valid port number between 1 and 65535", portStr)
	}
	cfg.Port = port

	timeoutStr := getEnv("SHUTDOWN_TIMEOUT", DefaultShutdownTimeout.String())
	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		return nil, fmt.Errorf("invalid SHUTDOWN_TIMEOUT value %q: %w", timeoutStr, err)
	}
	cfg.ShutdownTimeout = timeout

	return cfg, nil
}

// Address returns the formatted host:port string.
func (c *Config) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// getEnv gets an environment variable or falls back to a default value.
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// loadDotEnv loads key-value pairs from a .env file into the environment
// without overwriting existing environment variables.
func loadDotEnv(filepath string) {
	file, err := os.Open(filepath)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		// Strip quotes if present
		if len(val) >= 2 && ((val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'')) {
			val = val[1 : len(val)-1]
		}

		if os.Getenv(key) == "" {
			_ = os.Setenv(key, val)
		}
	}
}

