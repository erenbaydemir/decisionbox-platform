package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/decisionbox-io/decisionbox/libs/go-common/health"
	"github.com/decisionbox-io/decisionbox/libs/go-common/metrics"
)

type Config struct {
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	MaxHeaderBytes  int

	// TLS configuration. If both are set, the server starts with TLS.
	TLSCertFile string
	TLSKeyFile  string

	// OnShutdown is called after the HTTP server stops but before Run returns.
	// Use it to clean up resources (connection pools, background workers, etc.).
	OnShutdown func()
}

func DefaultConfig() Config {
	return Config{
		Port:            8080,
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		IdleTimeout:     120 * time.Second,
		ShutdownTimeout: 30 * time.Second,
	}
}

// Run starts the HTTP server with graceful shutdown on SIGTERM/SIGINT.
// It automatically registers /metrics and /health endpoints.
func Run(cfg Config, handler http.Handler, healthHandler *health.Handler) error {
	mux := http.NewServeMux()

	// Health endpoints
	if healthHandler != nil {
		mux.HandleFunc("/health", healthHandler.LivenessHandler())
		mux.HandleFunc("/health/ready", healthHandler.ReadinessHandler())
	}

	// Prometheus metrics
	mux.Handle("/metrics", metrics.Handler())

	// Application routes (wrapped with metrics middleware)
	mux.Handle("/", metrics.HTTPMiddleware(handler))

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      mux,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
	if cfg.MaxHeaderBytes > 0 {
		srv.MaxHeaderBytes = cfg.MaxHeaderBytes
	}

	errChan := make(chan error, 1)
	go func() {
		var err error
		if cfg.TLSCertFile != "" && cfg.TLSKeyFile != "" {
			err = srv.ListenAndServeTLS(cfg.TLSCertFile, cfg.TLSKeyFile)
		} else {
			err = srv.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	var shutdownErr error
	select {
	case err := <-errChan:
		shutdownErr = err
	case <-sigChan:
		ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		shutdownErr = srv.Shutdown(ctx)
	}

	if cfg.OnShutdown != nil {
		cfg.OnShutdown()
	}

	return shutdownErr
}
