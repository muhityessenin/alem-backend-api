package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	delivery "auth/internal/delivery/http"
	"auth/internal/repository"
	"auth/internal/usecase"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// TODO: Загружать из .env или флагов
	dbConnStr := "postgresql://root:V9ENYoHKjvjJ5m0pdWxZ6cxm7sQG9X1y@dpg-d2abklndiees738s14k0-a.oregon-postgres.render.com/alem_db?sslmode=disable"
	jwtSecret := "EACBXCIVYXWYFJKMWNEJWUIHQVISPJZWQATTIXDTJPWSNOAIOOJHLLQFMGDXGWNO"
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

	port := 8081
	log.Printf("Auth service starting on port %d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), router); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
