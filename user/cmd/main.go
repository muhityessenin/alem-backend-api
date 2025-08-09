package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"

	delivery "user/internal/delivery/http"
	"user/internal/repository"
	"user/internal/usecase"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
	dbConnStr := os.Getenv("DB_URL")
	jwtSecret := os.Getenv("JWT_SECRET")
	// log.Println("Running database migrations for user service...")
	// if err := runMigrations(dbConnStr); err != nil {
	//	log.Fatalf("Could not run database migrations: %v", err)
	// }
	// log.Println("Database migrations completed successfully.")
	dbpool, err := pgxpool.New(context.Background(), dbConnStr)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer dbpool.Close()

	userRepo := repository.NewUserPostgresRepo(dbpool)
	userUseCase := usecase.NewUserUseCase(userRepo)
	tokenUseCase := usecase.NewTokenUseCase(jwtSecret)

	userHandler := delivery.NewUserHandler(userUseCase, tokenUseCase)

	router := mux.NewRouter()

	router.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	userHandler.RegisterRoutes(router)

	port := os.Getenv("PORT")
	log.Printf("User service starting on port %s", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), router); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func runMigrations(dbURL string) error {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return fmt.Errorf("could not open db connection for migration: %w", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create postgres driver for migration: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	if err != nil {
		return fmt.Errorf("could not create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("could not apply migrations: %w", err)
	}

	return nil
}
