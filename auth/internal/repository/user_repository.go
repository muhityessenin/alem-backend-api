package repository

import (
	"auth/internal/domain"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
}

type userPostgresRepo struct {
	db *pgxpool.Pool
}

func NewUserPostgresRepo(db *pgxpool.Pool) UserRepository {
	return &userPostgresRepo{db: db}
}

func (r *userPostgresRepo) Create(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (id, email, password_hash, name, role, created_at) 
              VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.Exec(ctx, query, user.ID, user.Email, user.Password, user.Name, user.Role, user.CreatedAt)
	if err != nil {
		// Log the specific database error
		log.Printf("ERROR: failed to insert user into db: %v", err)
		// Wrap the error to provide context to the use case layer
		return fmt.Errorf("repository.Create: %w", err)
	}
	return nil
}
func (r *userPostgresRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, email, password_hash, name, role, is_verified, created_at FROM users WHERE email = $1`

	user := &domain.User{}

	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Name,
		&user.Role,
		&user.IsVerified,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}
