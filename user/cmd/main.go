package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"log"
	"net/http"

	delivery "user/internal/delivery/http"
	"user/internal/repository"
	"user/internal/usecase"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)

func main() {
	dbConnStr := "postgresql://root:V9ENYoHKjvjJ5m0pdWxZ6cxm7sQG9X1y@dpg-d2abklndiees738s14k0-a.oregon-postgres.render.com/alem_db?sslmode=disable"
	jwtSecret := "EACBXCIVYXWYFJKMWNEJWUIHQVISPJZWQATTIXDTJPWSNOAIOOJHLLQFMGDXGWNO"
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

	port := 8082
	log.Printf("User service starting on port %d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), router); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

// Add this function to the bottom of main.go
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
