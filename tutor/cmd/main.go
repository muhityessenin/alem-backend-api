package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	httpapi "tutor/internal/delivery/http"
	"tutor/internal/repository"
	"tutor/internal/usecase"

	"tutor/internal/config"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// 1) Load config
	cfg := config.MustLoad()

	// 2) DB pool
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pgxCfg, err := pgxpool.ParseConfig(cfg.DB.URL)
	if err != nil {
		log.Fatalf("db parse config: %v", err)
	}
	pgxCfg.MaxConns = cfg.DB.MaxConns

	db, err := pgxpool.NewWithConfig(ctx, pgxCfg)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer db.Close()

	// 3) UC wiring
	tutorRepo := repository.NewTutorRepository(db)
	tokenUC := usecase.NewTokenUseCase(strings.TrimSpace(cfg.JWT.AccessSecret))
	tutorUC := usecase.NewTutorUseCase(tutorRepo)

	// 4) Router + handlers
	r := mux.NewRouter()
	httpapi.NewTutorHandler(tutorUC, tokenUC).RegisterRoutes(r)

	// 5) CORS (из конфигов)
	cors := handlers.CORS(
		handlers.AllowedOrigins(cfg.CORS.AllowedOrigins),
		handlers.AllowedMethods(cfg.CORS.AllowedMethods),
		handlers.AllowedHeaders(cfg.CORS.AllowedHeaders),
	)

	// 6) HTTP server with timeouts
	srv := &http.Server{
		Addr:         cfg.App.Host + ":" + itoa(cfg.App.Port),
		Handler:      cors(r),
		ReadTimeout:  cfg.App.ReadTimeout,
		WriteTimeout: cfg.App.WriteTimeout,
		IdleTimeout:  cfg.App.IdleTimeout,
	}

	// Optional: fingerprint secret to убедиться, что секреты совпадают между сервисами
	fp := func(s string) string {
		sum := sha256.Sum256([]byte(strings.TrimSpace(s)))
		return hex.EncodeToString(sum[:])[:12]
	}
	log.Printf("[BOOT] env=%s port=%d db_max_conns=%d jwt_fp=%s",
		cfg.App.Env, cfg.App.Port, cfg.DB.MaxConns, fp(cfg.JWT.AccessSecret))

	// 7) Graceful shutdown
	errCh := make(chan error, 1)
	go func() {
		log.Printf("tutor service listening on http://%s:%d", cfg.App.Host, cfg.App.Port)
		errCh <- srv.ListenAndServe()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		log.Printf("shutdown signal: %s", sig)
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}

	shCtx, shCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shCancel()
	if err := srv.Shutdown(shCtx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}
	log.Printf("server stopped")
}

// маленький helper (без strconv импортов)
func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}
