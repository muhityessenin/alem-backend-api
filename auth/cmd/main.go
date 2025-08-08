package main

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"time"

	delivery "auth/internal/delivery/http"
	"auth/internal/repository"
	"auth/internal/usecase"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
	dbConnStr := os.Getenv("DATABASE_URL")
	jwtSecret := os.Getenv("JWT_SECRET")
	accessTokenTTL := 15 * time.Minute
	refreshTokenTTL := 24 * time.Hour * 30

	dbpool, err := pgxpool.New(context.Background(), dbConnStr)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer dbpool.Close()

	userRepo := repository.NewUserPostgresRepo(dbpool)

	// Обновляем создание usecase, передавая новые параметры
	authUseCase := usecase.NewAuthUseCase(userRepo, jwtSecret, accessTokenTTL, refreshTokenTTL)

	authHandler := delivery.NewAuthHandler(authUseCase)

	router := mux.NewRouter()
	authHandler.RegisterRoutes(router)

	port := os.Getenv("PORT")
	log.Printf("Auth service starting on port %s", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), router); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
