package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"0xfer/internal/config"
	"0xfer/internal/handlers"
	"0xfer/internal/repositories"
	"0xfer/internal/services"
	"0xfer/pkg/shortcut"
)

func main() {
	cfg := config.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:       slog.LevelInfo,
		AddSource:   false,
		ReplaceAttr: nil,
	}))

	shortcut.FatalIfErr(os.MkdirAll(cfg.DataDir, 0755))

	db, err := repositories.NewDBRepository(cfg.DBPath)
	shortcut.FatalIfErr(err)
	defer func() { _ = db.Close() }()

	fs, err := repositories.NewFileRepository(cfg.DataDir)
	shortcut.FatalIfErr(err)

	fileService := services.NewFileService(db, fs, cfg.TTL)
	cleanupService := services.NewCleanupService(db, fs, time.Minute, logger)

	baseURL := cfg.BaseURL
	if baseURL == "" {
		host := cfg.Addr
		if host[0] == ':' {
			host = "localhost" + host
		}
		baseURL = "http://" + host
	}

	mux := http.NewServeMux()
	mux.Handle("POST /", handlers.NewUploadHandler(fileService, cfg.MaxSize, baseURL))
	mux.Handle("PUT /", handlers.NewUploadHandler(fileService, cfg.MaxSize, baseURL))
	mux.Handle("GET /d/{id}", handlers.NewDownloadHandler(fileService))
	mux.Handle("DELETE /d/{id}/{secret}", handlers.NewDeleteHandler(fileService))
	mux.Handle("GET /health", handlers.NewHealthHandler())

	logger.Info("starting server", "addr", cfg.Addr, "data-dir", cfg.DataDir, "ttl", cfg.TTL)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go cleanupService.Start(ctx)

	srv := &http.Server{ //nolint:exhaustruct
		Addr:    cfg.Addr,
		Handler: mux,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "err", err)
		}
	}()

	<-ctx.Done()

	logger.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown error", "err", err)
	}
}
