package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"auth/internal/config"
	delivery "auth/internal/delivery/http"
	"auth/internal/repository"
	"auth/internal/usecase"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.MustLoad()

	// DB pool
	pgxCfg, err := pgxpool.ParseConfig(cfg.DB.URL)
	if err != nil {
		log.Fatalf("parse db url: %v", err)
	}
	pgxCfg.MaxConns = cfg.DB.MaxConns
	pgxCfg.MinConns = cfg.DB.MinConns
	pgxCfg.MaxConnLifetime = cfg.DB.MaxConnLifetime
	pgxCfg.MaxConnIdleTime = cfg.DB.MaxConnIdleTime

	dbctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	dbpool, err := pgxpool.NewWithConfig(dbctx, pgxCfg)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer dbpool.Close()

	// ...
	userRepo := repository.NewUserPostgresRepo(dbpool)
	sessRepo := repository.NewSessionsRepo(dbpool)
	otpRepo := repository.NewOTPRepo(dbpool)

	authUC := usecase.NewAuthUseCase(
		userRepo,
		sessRepo,
		otpRepo,
		usecase.Config{
			AccessSecret:   cfg.JWT.AccessSecret,
			RefreshSecret:  cfg.JWT.RefreshSecret,
			Issuer:         cfg.JWT.Issuer,
			Audience:       cfg.JWT.Audience,
			AccessTTL:      cfg.JWT.AccessTTL,
			RefreshTTL:     cfg.JWT.RefreshTTL,
			OTPEnabled:     cfg.OTP.Enabled,
			OTPTTL:         cfg.OTP.TTL,
			OTPLength:      cfg.OTP.Length,
			MaxOTPAttempts: 5,
		},
	)

	// http
	router := mux.NewRouter()
	authHandler := delivery.NewAuthHandler(authUC)
	authHandler.RegisterRoutes(router)

	cors := handlers.CORS(
		handlers.AllowedOrigins(cfg.CORS.AllowedOrigins),
		handlers.AllowedMethods(cfg.CORS.AllowedMethods),
		handlers.AllowedHeaders(cfg.CORS.AllowedHeaders),
	)

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.App.Host, cfg.App.Port),
		Handler:      cors(router),
		ReadTimeout:  cfg.App.ReadTimeout,
		WriteTimeout: cfg.App.WriteTimeout,
		IdleTimeout:  cfg.App.IdleTimeout,
	}

	// graceful shutdown
	errCh := make(chan error, 1)
	go func() {
		log.Printf("%s starting on %s", cfg.App.Name, srv.Addr)
		errCh <- srv.ListenAndServe()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		log.Printf("shutdown signal: %s", sig.String())
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("server shutdown error: %v", err)
		}
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}
}
