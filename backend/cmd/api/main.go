package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ffk00/iyte-hci-vespin/backend/internal/auth"
	"github.com/ffk00/iyte-hci-vespin/backend/internal/config"
	"github.com/ffk00/iyte-hci-vespin/backend/internal/db"
	"github.com/ffk00/iyte-hci-vespin/backend/internal/devices"
	"github.com/ffk00/iyte-hci-vespin/backend/internal/eqprofiles"
	"github.com/ffk00/iyte-hci-vespin/backend/internal/firmware"
	"github.com/ffk00/iyte-hci-vespin/backend/internal/partysessions"
	"github.com/ffk00/iyte-hci-vespin/backend/internal/server"
	"github.com/ffk00/iyte-hci-vespin/backend/internal/users"
	"github.com/google/uuid"
)

func main() {
	if err := run(); err != nil {
		slog.Error("api stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: parseLogLevel(cfg.LogLevel),
	}))
	slog.SetDefault(logger)

	ctx := context.Background()
	pool, err := db.OpenPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	queries := db.New(pool)
	tokens := auth.NewTokens(cfg.JWTSecret)
	authMW := auth.Middleware(tokens)

	defaultEQ, err := queries.GetDefaultEQProfile(ctx)
	if err != nil {
		return fmt.Errorf("load default eq profile: %w", err)
	}
	defaultEQID, err := uuid.FromBytes(defaultEQ.ID.Bytes[:])
	if err != nil {
		return fmt.Errorf("decode default eq profile id: %w", err)
	}

	r := server.NewRouter(server.Deps{
		AuthMW:          authMW,
		AuthHandler:     auth.NewHandler(queries, pool, tokens),
		UserHandler:     users.NewHandler(queries),
		DeviceHandler:   devices.NewHandler(queries, defaultEQID),
		EQHandler:       eqprofiles.NewHandler(queries, pool),
		PartyHandler:    partysessions.NewHandler(queries, pool),
		FirmwareHandler: firmware.NewHandler(queries),
	})

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("server starting", "addr", cfg.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case <-signalCtx.Done():
		slog.Info("server shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown server: %w", err)
		}
		return <-errCh
	case err := <-errCh:
		return err
	}
}

func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
