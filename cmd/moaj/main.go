package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/MOAJ-Mentor-of-all-Jacks/MOAJ/internal/api"
	"github.com/MOAJ-Mentor-of-all-Jacks/MOAJ/internal/config"
	"github.com/MOAJ-Mentor-of-all-Jacks/MOAJ/internal/enterprise"
	"github.com/MOAJ-Mentor-of-all-Jacks/MOAJ/internal/observability"
)

var version = "dev"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, os.Args[1:], os.Stdout, os.Stderr); err != nil {
		observability.NewLogger(os.Stderr, slog.LevelInfo).Error("startup or server failed", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	command := "serve"
	if len(args) > 0 {
		command = args[0]
	}
	if len(args) > 1 {
		return errors.New("usage: moaj [serve|version|help]")
	}
	switch command {
	case "help", "-h", "--help":
		_, err := fmt.Fprintln(stdout, "Usage: moaj [serve|version|help]\n\nserve starts the HTTP runtime (default). Database, workers and UI arrive in later phases.")
		return err
	case "version":
		_, err := fmt.Fprintf(stdout, "moaj %s (%s)\n", version, enterprise.Edition())
		return err
	case "serve":
		// Continue below. Never echo arbitrary arguments into logs.
	default:
		return errors.New("unknown command; usage: moaj [serve|version|help]")
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("configuration: %w", err)
	}
	logger := observability.NewLogger(stderr, cfg.LogLevel)
	metrics := observability.NewMetrics(version, enterprise.Edition())
	app := api.New(api.Options{
		Logger: logger, Metrics: metrics, MetricsToken: cfg.MetricsToken,
		MaxBodyBytes: cfg.MaxBodyBytes,
		Enterprise:   enterprise.New(),
	})
	server := api.NewHTTPServer(cfg, app, logger)
	listener, err := net.Listen("tcp", cfg.HTTPAddr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	logger.Info("HTTP runtime listening", "address", listener.Addr().String(), "edition", enterprise.Edition(), "version", version, "readiness_scope", "http-runtime")
	return api.Serve(ctx, server, listener, cfg.ShutdownTimeout, app.BeginShutdown, logger)
}
