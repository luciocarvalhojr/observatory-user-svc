// @title           Observatory User Service
// @version         1.0
// @description     User management service for the Observatory platform.
// @contact.name    Lucio Carvalho
// @host            localhost:8082
// @BasePath        /

// Package main is the entry point for the user-svc application.
package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/luciocarvalhojr/observatory-user-svc/internal/config"
	"github.com/luciocarvalhojr/observatory-user-svc/internal/handler"
	"github.com/luciocarvalhojr/observatory-user-svc/internal/middleware"
	"github.com/luciocarvalhojr/observatory-user-svc/internal/publisher"
	"github.com/luciocarvalhojr/observatory-user-svc/internal/repository"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// ── Structured logging ────────────────────────────────────────────
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.With().Str("service", "user-svc").Logger()

	// ── Config ────────────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	if cfg.Env == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// ── PostgreSQL ────────────────────────────────────────────────────
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	repo, err := repository.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("postgres unreachable")
	}
	defer repository.Close(repo)
	log.Info().Msg("postgres connected")

	// ── NATS ──────────────────────────────────────────────────────────
	pub, err := publisher.New(cfg.NATSUrl)
	if err != nil {
		log.Fatal().Err(err).Msg("nats unreachable")
	}
	defer pub.Close()
	log.Info().Str("url", cfg.NATSUrl).Msg("nats connected")

	// ── Router ────────────────────────────────────────────────────────
	r := gin.New()
	r.Use(middleware.Logger(), middleware.Recovery())

	health := handler.NewHealth(repo)
	health.Register(&r.RouterGroup)

	users := handler.NewUser(repo, pub)
	users.Register(r.Group("/users"))

	// ── Server ────────────────────────────────────────────────────────
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info().Str("port", cfg.Port).Msg("user-svc starting")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	// ── Graceful shutdown ─────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down gracefully")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("shutdown error")
	}
	log.Info().Msg("user-svc stopped")
}
