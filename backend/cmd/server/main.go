package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"unesarent/backend/internal/auth"
	"unesarent/backend/internal/config"
	"unesarent/backend/internal/httpapi"
	"unesarent/backend/internal/user"
)

func main() {
	if err := config.LoadDotEnv(".env"); err != nil {
		slog.Warn("gagal memuat .env", "error", err)
	}

	cfg, err := config.Load()
	if err != nil {
		slog.Error("konfigurasi tidak valid", "error", err)
		os.Exit(1)
	}

	store, err := user.NewMemoryStore(cfg.StorePath)
	if err != nil {
		slog.Error("gagal menyiapkan storage", "error", err)
		os.Exit(1)
	}

	tokens := auth.NewTokenIssuer(cfg.JWTSecret, cfg.TokenTTL)
	svc := user.NewService(store, tokens)
	router := httpapi.NewRouter(cfg, httpapi.NewAuthHandler(svc), tokens)

	server := &http.Server{
		Addr:              cfg.Addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErr := make(chan error, 1)
	go func() {
		slog.Info("server berjalan", "addr", cfg.Addr, "token_ttl", cfg.TokenTTL.String())
		serverErr <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server berhenti", "error", err)
			os.Exit(1)
		}
	case <-ctx.Done():
		slog.Info("mematikan server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("gagal shutdown", "error", err)
			os.Exit(1)
		}
	}
}
