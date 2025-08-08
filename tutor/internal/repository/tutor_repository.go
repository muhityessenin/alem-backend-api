package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"tutor/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TutorRepository interface {
	FindAll(ctx context.Context, filters map[string]string, page, limit int) ([]domain.Tutor, int, error)
	FindByID(ctx context.Context, id string) (*domain.Tutor, error)
	UpdateProfile(ctx context.Context, tutor *domain.Tutor) error
	UpdateStatus(ctx context.Context, tutorID string, status string, reason string) error
	FindPending(ctx context.Context, page, limit int) ([]domain.Tutor, int, error)
}

type tutorPostgresRepo struct {
	db *pgxpool.Pool
}

func NewTutorPostgresRepo(db *pgxpool.Pool) TutorRepository {
	return &tutorPostgresRepo{db: db}
}
func (r *tutorPostgresRepo) FindPending(ctx context.Context, page, limit int) ([]domain.Tutor, int, error) {
	// Base query to select only pending tutors
	baseQuery := `SELECT id, name, avatar, subjects, languages, rating, review_count, price, description, country, is_online FROM tutors WHERE status = 'pending'`
	countQuery := `SELECT COUNT(*) FROM tutors WHERE status = 'pending'`

	// Get total count for pagination
	var total int
	if err := r.db.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []domain.Tutor{}, 0, nil
	}

	// Add pagination
	offset := (page - 1) * limit
	baseQuery += fmt.Sprintf(" ORDER BY created_at ASC LIMIT %d OFFSET %d", limit, offset)

	rows, err := r.db.Query(ctx, baseQuery)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	// This logic is identical to your FindAll method
	var tutors []domain.Tutor
	for rows.Next() {
		var t domain.Tutor
		var avatar, description, country sql.NullString
		var subjects, languages []string

		if err := rows.Scan(
			&t.ID, &t.Name, &avatar, &subjects, &languages,
			&t.Rating, &t.ReviewCount, &t.Price, &description, &country, &t.IsOnline,
		); err != nil {
			return nil, 0, err
		}

		if avatar.Valid {
			t.Avatar = avatar.String
		}
		if description.Valid {
			t.Description = description.String
		}
		if country.Valid {
			t.Country = country.String
		}
		t.Subjects = subjects
		t.Languages = languages

		tutors = append(tutors, t)
	}

	return tutors, total, nil
}
func (r *tutorPostgresRepo) FindAll(ctx context.Context, filters map[string]string, page, limit int) ([]domain.Tutor, int, error) {
	baseQuery := `SELECT id, name, avatar, subjects, languages, rating, review_count, price, description, country, is_online FROM tutors WHERE status = 'approved'`
	countQuery := `SELECT COUNT(*) FROM tutors WHERE status = 'approved'`

	var whereClauses []string
	var args []interface{}
	argId := 1

	for key, value := range filters {
		if value == "" {
			continue
		}
		if key == "search" {
			whereClauses = append(whereClauses, fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d)", argId, argId+1))
			args = append(args, "%"+value+"%", "%"+value+"%")
			argId += 2
		}
	}

	if len(whereClauses) > 0 {
		whereStr := " AND " + strings.Join(whereClauses, " AND ")
		baseQuery += whereStr
		countQuery += whereStr
	}

	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []domain.Tutor{}, 0, nil
	}

	offset := (page - 1) * limit
	baseQuery += fmt.Sprintf(" ORDER BY rating DESC LIMIT $%d OFFSET $%d", argId, argId+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tutors []domain.Tutor
	for rows.Next() {
		var t domain.Tutor
		var avatar, description, country sql.NullString

		// FIX: Scan directly into &t.Subjects. pgx handles NULL arrays automatically.
		if err := rows.Scan(
			&t.ID, &t.Name, &avatar, &t.Subjects, &t.Languages,
			&t.Rating, &t.ReviewCount, &t.Price, &description, &country, &t.IsOnline,
		); err != nil {
			return nil, 0, err
		}

		if avatar.Valid {
			t.Avatar = avatar.String
		}
		if description.Valid {
			t.Description = description.String
		}
		if country.Valid {
			t.Country = country.String
		}

		tutors = append(tutors, t)
	}
	return tutors, total, nil
}

func (r *tutorPostgresRepo) FindByID(ctx context.Context, id string) (*domain.Tutor, error) {
	query := `SELECT
       id, name, avatar, subjects, languages, rating, review_count, price, description, country, is_online,
       video_intro, teaching_style, education, certifications, availability
       FROM tutors WHERE id = $1 AND status = 'approved'`

	t := &domain.Tutor{}
	var avatar, description, country, videoIntro, teachingStyle sql.NullString
	var educationBytes, certificationsBytes, availabilityBytes []byte

	err := r.db.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.Name, &avatar, &t.Subjects, &t.Languages,
		&t.Rating, &t.ReviewCount, &t.Price, &description, &country, &t.IsOnline,
		&videoIntro, &teachingStyle, &educationBytes, &certificationsBytes, &availabilityBytes,
	)
	if err != nil {
		return nil, err
	}

	if avatar.Valid {
		t.Avatar = avatar.String
	}
	if description.Valid {
		t.Description = description.String
	}
	if country.Valid {
		t.Country = country.String
	}
	if videoIntro.Valid {
		t.VideoIntro = videoIntro.String
	}
	if teachingStyle.Valid {
		t.TeachingStyle = teachingStyle.String
	}

	if len(educationBytes) > 0 {
		json.Unmarshal(educationBytes, &t.Education)
	}
	if len(certificationsBytes) > 0 {
		json.Unmarshal(certificationsBytes, &t.Certifications)
	}
	if len(availabilityBytes) > 0 {
		json.Unmarshal(availabilityBytes, &t.Availability)
	}

	return t, nil
}

func (r *tutorPostgresRepo) UpdateProfile(ctx context.Context, tutor *domain.Tutor) error {
	educationBytes, _ := json.Marshal(tutor.Education)
	certificationsBytes, _ := json.Marshal(tutor.Certifications)
	availabilityBytes, _ := json.Marshal(tutor.Availability)

	query := `
       INSERT INTO tutors (
          id, user_id, name, price, description, subjects, languages,
          video_intro, teaching_style, education, certifications, availability, status
       ) VALUES (
          $1, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 'pending'
       )
       ON CONFLICT (id) DO UPDATE SET
          name = EXCLUDED.name,
          price = EXCLUDED.price,
          description = EXCLUDED.description,
          subjects = EXCLUDED.subjects,
          languages = EXCLUDED.languages,
          video_intro = EXCLUDED.video_intro,
          teaching_style = EXCLUDED.teaching_style,
          education = EXCLUDED.education,
          certifications = EXCLUDED.certifications,
          availability = EXCLUDED.availability,
          status = 'pending', 
          updated_at = NOW();
    `
	_, err := r.db.Exec(ctx, query,
		tutor.ID, tutor.Name, tutor.Price, tutor.Description, tutor.Subjects, tutor.Languages,
		tutor.VideoIntro, tutor.TeachingStyle, educationBytes, certificationsBytes, availabilityBytes,
	)
	return err
}

func (r *tutorPostgresRepo) UpdateStatus(ctx context.Context, tutorID string, status string, reason string) error {
	query := `UPDATE tutors SET status = $1, rejection_reason = $2, updated_at = NOW() WHERE id = $3`
	_, err := r.db.Exec(ctx, query, status, reason, tutorID)
	return err
}
