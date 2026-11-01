package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/MOAJ-Mentor-of-all-Jacks/MOAJ/internal/config"
)

func NewHTTPServer(cfg config.Config, handler http.Handler, logger *slog.Logger) *http.Server {
	return &http.Server{
		Addr: cfg.HTTPAddr, Handler: handler,
		ReadTimeout: cfg.ReadTimeout, ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		WriteTimeout: cfg.WriteTimeout, IdleTimeout: cfg.IdleTimeout,
		MaxHeaderBytes: 1 << 20,
		ErrorLog:       slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}
}

// Serve owns the listener and drains active requests on cancellation. Request
// contexts are not canceled just because draining starts.
func Serve(ctx context.Context, server *http.Server, listener net.Listener, timeout time.Duration, drain func(), logger *slog.Logger) error {
	defer listener.Close()
	result := make(chan error, 1)
	go func() { result <- server.Serve(listener) }()
	select {
	case err := <-result:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("serve HTTP: %w", err)
	case <-ctx.Done():
		drain()
		logger.Info("HTTP shutdown started")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			_ = server.Close()
			<-result
			return fmt.Errorf("drain HTTP: %w", err)
		}
		err := <-result
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve HTTP: %w", err)
		}
		logger.Info("HTTP shutdown complete")
		return nil
	}
}
