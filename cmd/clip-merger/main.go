package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/yash-sojitra-20/clip-merger/internal/api"
	"github.com/yash-sojitra-20/clip-merger/internal/config"
	"github.com/yash-sojitra-20/clip-merger/internal/observability"
)

func main() {
	portFlag := flag.Int("port", 0, "HTTP server port override")
	hostFlag := flag.String("host", "", "HTTP server host override")
	envFlag := flag.String("env", "", "Application environment (development, staging, production)")
	logLevelFlag := flag.String("log-level", "", "Logging level (debug, info, warn, error)")
	logFormatFlag := flag.String("log-format", "", "Logging format (text, json)")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}

	// Apply CLI overrides
	if *portFlag > 0 {
		cfg.Port = *portFlag
	}
	if *hostFlag != "" {
		cfg.Host = *hostFlag
	}
	if *envFlag != "" {
		cfg.AppEnv = *envFlag
	}
	if *logLevelFlag != "" {
		cfg.LogLevel = *logLevelFlag
	}
	if *logFormatFlag != "" {
		cfg.LogFormat = *logFormatFlag
	}

	// Initialize structured logger
	logger := observability.SetupLogger(cfg.LogLevel, cfg.LogFormat)
	logger.Info("Starting Clip Merger Service",
		slog.String("service", cfg.ServiceName),
		slog.String("version", cfg.ServiceVersion),
		slog.String("env", cfg.AppEnv),
		slog.String("address", cfg.Address()),
		slog.String("log_level", cfg.LogLevel),
		slog.String("log_format", cfg.LogFormat),
	)

	// Context listening for OS interrupt / terminate signals
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	server := api.NewServer(cfg, logger)

	if err := server.Start(ctx); err != nil {
		logger.Error("Server stopped with error", observability.ErrorAttr(err))
		os.Exit(1)
	}

	logger.Info("Clip Merger Service stopped successfully")
}

