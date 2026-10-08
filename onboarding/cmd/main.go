package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/controllers"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/domains"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/repositories"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/server"
	"github.com/ksendel-coder/onboarding_croquant/platform/config"
	"github.com/ksendel-coder/onboarding_croquant/platform/logger"
	"github.com/ksendel-coder/onboarding_croquant/platform/pg"
)

// @title Cluer Onboarding
// @version 0.1.0
// @description Interactive onboarding platform.

// @host localhost:8080
// @BasePath /v1
func main() {
	cfg := config.Load()

	defaultLogger := logger.New(cfg.Logger.Default.Level)
	defaultLogger.Info().Msg("Logger setup successfully")

	if cfg.PostgresConfig == nil {
		defaultLogger.Fatal().Msg("postgres config is required for the onboarding service")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	dsn := cfg.PostgresConfig.GetDSN()

	pool, err := pg.Connect(ctx, dsn, defaultLogger)
	if err != nil {
		defaultLogger.Fatal().
			Err(err).
			Str("dsn", cfg.PostgresConfig.SafeDSN()).
			Msg("Database connection failed")
	}
	defer pool.Close()

	runtimeRepo := repositories.NewRuntimeRepository(pool, defaultLogger)

	runtimeDomain := domains.NewRuntimeDomain(runtimeRepo)

	srvLogger := logger.New(cfg.GetLoggerConfig("server").Level)

	runtimeController := controllers.NewRuntimeController(runtimeDomain)

	srv := server.NewServer(cfg.ServerConfig, &server.CreateStruct{
		Logger:            srvLogger,
		RuntimeController: runtimeController,
	})

	defaultLogger.Debug().Msg("Server created successfully")

	go func() {
		if err := srv.Start(); err != nil {
			defaultLogger.Fatal().Err(err).Msg("Application failed to run")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	shutdownCtx, shutdownCancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		defaultLogger.Error().Err(err).Msg("Server shutdown")
	}

	defaultLogger.Info().Msg("Server exiting")
}
