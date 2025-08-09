package repository

import (
	"auth/internal/domain"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrDuplicateEmail = errors.New("email already exists")
var ErrDuplicatePhone = errors.New("phone already exists")
var ErrNoRows = pgx.ErrNoRows

type UserRepository interface {
	Create(ctx context.Context, u *domain.User, passwordHash string) error
	GetByEmail(ctx context.Context, email string) (*domain.User, string, error)
	GetByPhone(ctx context.Context, phone string) (*domain.User, error)
}

type userPostgresRepo struct {
	db *pgxpool.Pool
}

func NewUserPostgresRepo(db *pgxpool.Pool) UserRepository {
	return &userPostgresRepo{db: db}
}

// Create: вставляем в users + создаём пустой student_profile (как дефолт после регистрации)
func (r *userPostgresRepo) Create(ctx context.Context, u *domain.User, passwordHash string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	q := `
INSERT INTO users (id, email, phone_e164, password_hash, role, locale, status, created_at, updated_at, country_code, first_name, last_name)
VALUES ($1, LOWER($2), $3, $4, $5, 'ru', 'active', $6, $6, NULL, $7, $8)`
	// Примечание: колонки first_name/last_name должны быть добавлены вашей миграцией 0007.
	// Если их нет, замени на display_name в student_profiles.

	now := time.Now().UTC()
	_, err = tx.Exec(ctx, q,
		u.ID, u.Email, u.Phone, passwordHash, u.Role, now, u.FirstName, u.LastName,
	)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "users_email_key") {
			return ErrDuplicateEmail
		}
		if strings.Contains(strings.ToLower(err.Error()), "users_phone_e164_key") {
			return ErrDuplicatePhone
		}
		return fmt.Errorf("insert users: %w", err)
	}

	// создаём пустой профиль студента (требование бизнес-логики)
	_, err = tx.Exec(ctx, `INSERT INTO student_profiles (user_id, created_at, updated_at, prefs) VALUES ($1, $2, $2, '{}'::jsonb)`, u.ID, now)
	if err != nil {
		return fmt.Errorf("insert student_profiles: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}
	u.CreatedAt = now
	return nil
}

func (r *userPostgresRepo) GetByEmail(ctx context.Context, email string) (*domain.User, string, error) {
	row := r.db.QueryRow(ctx, `
SELECT id, email, COALESCE(phone_e164,''), COALESCE(first_name,''), COALESCE(last_name,''), role, created_at, password_hash
FROM users WHERE email = LOWER($1) AND deleted_at IS NULL
`, email)

	var u domain.User
	var passHash string
	if err := row.Scan(&u.ID, &u.Email, &u.Phone, &u.FirstName, &u.LastName, &u.Role, &u.CreatedAt, &passHash); err != nil {
		return nil, "", err
	}
	return &u, passHash, nil
}

func (r *userPostgresRepo) GetByPhone(ctx context.Context, phone string) (*domain.User, error) {
	row := r.db.QueryRow(ctx, `
SELECT id, email, COALESCE(phone_e164,''), COALESCE(first_name,''), COALESCE(last_name,''), role, created_at
FROM users WHERE phone_e164 = $1 AND deleted_at IS NULL
`, phone)

	var u domain.User
	if err := row.Scan(&u.ID, &u.Email, &u.Phone, &u.FirstName, &u.LastName, &u.Role, &u.CreatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}
