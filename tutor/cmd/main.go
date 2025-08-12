package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"tutor/internal/delivery/http"
	"tutor/internal/repository"
	"tutor/internal/usecase"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbURL := mustEnv("DB_URL")
	port := env("PORT", "8081")
	jwtSecret := mustEnv("JWT_SECRET")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pgxCfg, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		log.Fatalf("parse db url: %v", err)
	}
	pgxCfg.MaxConns = 20
	db, err := pgxpool.NewWithConfig(ctx, pgxCfg)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer db.Close()

	// wiring (classes)
	tutorRepo := repository.NewTutorRepository(db)
	tokenUC := usecase.NewTokenUseCase(jwtSecret)
	tutorUC := usecase.NewTutorUseCase(tutorRepo)

	r := mux.NewRouter()
	api := httpapi.NewTutorHandler(tutorUC, tokenUC)
	api.RegisterRoutes(r)

	addr := ":" + port
	log.Printf("tutor service listening on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}

func mustEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("missing env %s", k)
	}
	return v
}
func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
