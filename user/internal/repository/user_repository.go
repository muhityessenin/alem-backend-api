package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
	"user/internal/domain"
)

type UserRepository interface {
	GetProfileByID(ctx context.Context, id string) (*domain.User, error)
	UpdateProfile(ctx context.Context, user *domain.User) error
	UpdateAvatarURL(ctx context.Context, userID, avatarURL string) error
	FindStudents(ctx context.Context, page, limit int) ([]domain.User, int, error) // New method
}

type userPostgresRepo struct {
	db *pgxpool.Pool
}

func NewUserPostgresRepo(db *pgxpool.Pool) UserRepository {
	return &userPostgresRepo{db: db}
}
func (r *userPostgresRepo) GetProfileByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT 
			id, email, name, role, avatar, age, learning_goals, description,
			notifications_lessons, notifications_messages, notifications_reminders, created_at 
		FROM users WHERE id = $1`

	var avatar sql.NullString
	var age sql.NullInt64
	var description sql.NullString

	user := &domain.User{}

	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Name, &user.Role,
		&avatar,
		&age,
		pq.Array(&user.LearningGoals),
		&description,
		&user.Notifications.Lessons, &user.Notifications.Messages, &user.Notifications.Reminders,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	if avatar.Valid {
		user.Avatar = avatar.String
	}
	if age.Valid {
		user.Age = int(age.Int64)
	}
	if description.Valid {
		user.Description = description.String
	}

	return user, nil
}

func (r *userPostgresRepo) UpdateProfile(ctx context.Context, user *domain.User) error {
	// Мы будем обновлять только те поля, которые пользователь может менять.
	// email, role, id и т.д. не меняются.
	query := `
		UPDATE users
		SET 
			name = $1,
			age = $2,
			avatar = $3,
			learning_goals = $4,
			description = $5,
			notifications_lessons = $6,
			notifications_messages = $7,
			notifications_reminders = $8,
			updated_at = NOW()
		WHERE id = $9
	`
	_, err := r.db.Exec(ctx, query,
		user.Name,
		user.Age,
		user.Avatar,
		pq.Array(user.LearningGoals),
		user.Description,
		user.Notifications.Lessons,
		user.Notifications.Messages,
		user.Notifications.Reminders,
		user.ID, // ID пользователя для условия WHERE
	)

	return err
}

func (r *userPostgresRepo) UpdateAvatarURL(ctx context.Context, userID, avatarURL string) error {
	query := `UPDATE users SET avatar = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, avatarURL, userID)
	return err
}

func (r *userPostgresRepo) FindStudents(ctx context.Context, page, limit int) ([]domain.User, int, error) {
	// Base query to select only students
	baseQuery := `SELECT id, name, avatar, learning_goals, native_language, level, budget, availability, description, age, country
                 FROM users WHERE role = 'student'`
	countQuery := `SELECT COUNT(*) FROM users WHERE role = 'student'`

	// TODO: Add dynamic WHERE clauses for filters (level, budget, etc.)

	var total int
	if err := r.db.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Add pagination
	offset := (page - 1) * limit
	baseQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT %d OFFSET %d", limit, offset)

	rows, err := r.db.Query(ctx, baseQuery)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var students []domain.User
	for rows.Next() {
		var s domain.User
		// Use nullable types for fields that can be NULL
		var avatar, description, country, level, nativeLanguage sql.NullString
		var age sql.NullInt64
		var budget sql.NullFloat64

		if err := rows.Scan(
			&s.ID, &s.Name, &avatar, pq.Array(&s.LearningGoals), &nativeLanguage, &level,
			&budget, pq.Array(&s.Availability), &description, &age, &country,
		); err != nil {
			return nil, 0, err
		}

		// Assign values from nullable types
		if avatar.Valid {
			s.Avatar = avatar.String
		}
		// ... assign other nullable fields ...

		students = append(students, s)
	}

	return students, total, nil
}
