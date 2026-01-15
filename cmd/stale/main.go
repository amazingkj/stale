package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jiin/stale/internal/api"
	"github.com/jiin/stale/internal/config"
	"github.com/jiin/stale/internal/database"
	"github.com/jiin/stale/internal/repository"
	"github.com/jiin/stale/internal/service/scanner"
	"github.com/jiin/stale/internal/service/scheduler"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var Version = "0.1.0"

func main() {
	// Load configuration
	cfg := config.Load()

	// Setup logging
	setupLogging(cfg.LogLevel)

	log.Info().
		Str("version", Version).
		Str("port", cfg.Port).
		Str("db_path", cfg.DatabasePath).
		Msg("starting stale")

	// Initialize database
	db, err := database.New(cfg.DatabasePath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer db.Close()

	// Run migrations
	if err := database.Migrate(db); err != nil {
		log.Fatal().Err(err).Msg("failed to run migrations")
	}

	log.Info().Msg("database initialized")

	// Initialize repositories
	sourceRepo := repository.NewSourceRepository(db)
	repoRepo := repository.NewRepoRepository(db)
	depRepo := repository.NewDependencyRepository(db)
	scanRepo := repository.NewScanRepository(db)

	// Initialize services
	scannerService := scanner.New(sourceRepo, repoRepo, depRepo, scanRepo)
	schedulerService := scheduler.New(scannerService, scanRepo, cfg.ScanIntervalHours)

	// Start background scheduler
	go schedulerService.Start()

	// Initialize router
	router := api.NewRouter(db, schedulerService)

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Info().Str("addr", srv.Addr).Msg("server started")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down server...")

	// Stop scheduler
	schedulerService.Stop()

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("server shutdown error")
	}

	log.Info().Msg("server stopped")
}

func setupLogging(level string) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Pretty console output for development
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	switch level {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}
