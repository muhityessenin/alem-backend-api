package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	delivery "tutor/internal/delivery/http"
	"tutor/internal/repository"
	"tutor/internal/usecase"
)

func main() {

	dbConnStr := "postgresql://root:V9ENYoHKjvjJ5m0pdWxZ6cxm7sQG9X1y@dpg-d2abklndiees738s14k0-a.oregon-postgres.render.com/alem_db?sslmode=disable"
	port := 8083
	jwtSecret := "EACBXCIVYXWYFJKMWNEJWUIHQVISPJZWQATTIXDTJPWSNOAIOOJHLLQFMGDXGWNO"

	// log.Println("Running database migrations...")
	// err := runMigrations(dbConnStr)
	// if err != nil {
	//	log.Fatalf("Could not run database migrations: %v", err)
	// }
	// log.Println("Database migrations completed successfully.")

	dbpool, err := pgxpool.New(context.Background(), dbConnStr)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer dbpool.Close()

	tutorRepo := repository.NewTutorPostgresRepo(dbpool)
	reviewRepo := repository.NewReviewPostgresRepo(dbpool)
	tutorUseCase := usecase.NewTutorUseCase(tutorRepo)
	reviewUseCase := usecase.NewReviewUseCase(reviewRepo)
	tokenUseCase := usecase.NewTokenUseCase(jwtSecret) // <-- Add this

	tutorHandler := delivery.NewTutorHandler(tutorUseCase, reviewUseCase, tokenUseCase)
	router := mux.NewRouter()
	tutorHandler.RegisterRoutes(router)

	log.Printf("Tutor service starting on port %d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), router); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
func runMigrations(dbURL string) error {
	// Open a standard database connection for the migration.
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return fmt.Errorf("could not open db connection for migration: %w", err)
	}
	defer db.Close()

	// Create a new driver instance with the connection.
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create postgres driver for migration: %w", err)
	}

	// Create a new migrate instance.
	m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	if err != nil {
		return fmt.Errorf("could not create migrate instance: %w", err)
	}

	// Run the migrations.
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("could not apply migrations: %w", err)
	}

	return nil
}
