package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example.com/market/internal/config"
	_ "example.com/market/docs"
	_ "example.com/market/internal/domain"
	"example.com/market/internal/logger"
	"example.com/market/internal/services/ads"
	"example.com/market/internal/services/auth"
	"example.com/market/internal/storage/postgres"
	httptransport "example.com/market/internal/transport/http"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title Marketplace API
// @version 1.0
// @description This is a simple marketplace API.
// @host localhost:8080
// @BasePath /v1
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	// 1. Load config
	cfg := config.MustLoad()

	// 2. Setup logger
	log := logger.SetupLogger(cfg.LogLevel)
	log.Info("starting application", slog.String("log_level", cfg.LogLevel))

	// 3. Init storage (db)
	db, err := postgres.New(context.Background(), cfg.DB.DSN, log)
	if err != nil {
		log.Error("failed to init storage", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()
	log.Info("database connection established")

	// 4. Init services
	authService := auth.New(db, cfg.Auth.JWTSecret, cfg.Auth.TokenTTL)
	adsService := ads.New(db)

	// 5. Init transport (router, handlers)
	authHandler := httptransport.NewAuthHandler(authService, log)
	adsHandler := httptransport.NewAdsHandler(adsService, log)

	// Init router
	router := httptransport.NewRouter(log, authHandler, adsHandler, authService)
	router.Get("/swagger/*", httpSwagger.WrapHandler)

	// 6. Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	srv := &http.Server{
		Addr:         cfg.HTTP.Addr,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		log.Info("server started", slog.String("addr", cfg.HTTP.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server failed to start", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	<-done
	log.Info("stopping server")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("server shutdown failed", slog.String("error", err.Error()))
	}

	log.Info("server stopped gracefully")
}
