package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alfredtm/yeeted/internal/db"
	"github.com/alfredtm/yeeted/internal/handler"
	"github.com/alfredtm/yeeted/internal/telemetry"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdownTracing, err := telemetry.Init(ctx, "yeeted")
	if err != nil {
		log.Printf("telemetry init failed: %v (continuing)", err)
	}
	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") == "" {
		log.Printf("tracing disabled (no OTEL_EXPORTER_OTLP_ENDPOINT)")
	}

	pool, err := db.Connect(ctx)
	if err != nil {
		log.Printf("db connect failed: %v (continuing without persistence)", err)
	}
	if pool != nil {
		log.Printf("connected to postgres")
	} else {
		log.Printf("no DATABASE_URL set, running without persistence")
	}

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           handler.NewRouter(pool),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		log.Fatalf("server error: %v", err)
	case sig := <-stop:
		log.Printf("shutting down (signal: %s)", sig)
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}

	if shutdownTracing != nil {
		if err := shutdownTracing(shutdownCtx); err != nil {
			log.Printf("tracing shutdown error: %v", err)
		}
	}

	if pool != nil {
		pool.Close()
	}
}
