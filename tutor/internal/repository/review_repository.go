package repository

import (
	"context"
	"github.com/google/uuid"
	"strconv"
	"time"
	"tutor/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ReviewRepository interface {
	FindByTutorID(ctx context.Context, tutorID string, page, limit int) ([]domain.Review, int, error)
	GetStatsByTutorID(ctx context.Context, tutorID string) (*domain.ReviewStats, error)
	Create(ctx context.Context, review *domain.Review) error
}

type reviewPostgresRepo struct {
	db *pgxpool.Pool
}

func NewReviewPostgresRepo(db *pgxpool.Pool) ReviewRepository {
	return &reviewPostgresRepo{db: db}
}

func (r *reviewPostgresRepo) FindByTutorID(ctx context.Context, tutorID string, page, limit int) ([]domain.Review, int, error) {
	// For this query, we'd ideally JOIN with the users table to get student name/avatar.
	// For simplicity, we'll return placeholder data for now.
	query := `SELECT id, student_id, rating, comment, subject, created_at FROM reviews
            WHERE tutor_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	offset := (page - 1) * limit
	rows, err := r.db.Query(ctx, query, tutorID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var reviews []domain.Review
	for rows.Next() {
		var rev domain.Review
		if err := rows.Scan(&rev.ID, &rev.StudentId, &rev.Rating, &rev.Comment, &rev.Subject, &rev.CreatedAt); err != nil {
			return nil, 0, err
		}
		// TODO: In a real app, fetch student details from the User service.
		rev.StudentName = "Alex Johnson"
		rev.StudentAvatar = "https://example.com/avatar.jpg"
		reviews = append(reviews, rev)
	}

	// Get total count
	var total int
	countQuery := "SELECT COUNT(*) FROM reviews WHERE tutor_id = $1"
	if err := r.db.QueryRow(ctx, countQuery, tutorID).Scan(&total); err != nil {
		return nil, 0, err
	}

	return reviews, total, nil
}

func (r *reviewPostgresRepo) GetStatsByTutorID(ctx context.Context, tutorID string) (*domain.ReviewStats, error) {
	stats := &domain.ReviewStats{
		RatingDistribution: make(map[string]int),
	}

	// Get average rating
	avgQuery := `SELECT COALESCE(AVG(rating), 0.0) FROM reviews WHERE tutor_id = $1`
	if err := r.db.QueryRow(ctx, avgQuery, tutorID).Scan(&stats.AverageRating); err != nil {
		return nil, err
	}

	// Get rating distribution
	distQuery := `SELECT rating, COUNT(*) FROM reviews WHERE tutor_id = $1 GROUP BY rating`
	rows, err := r.db.Query(ctx, distQuery, tutorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var rating, count int
		if err := rows.Scan(&rating, &count); err != nil {
			return nil, err
		}
		stats.RatingDistribution[strconv.Itoa(rating)] = count
	}

	return stats, nil
}

func (r *reviewPostgresRepo) Create(ctx context.Context, review *domain.Review) error {
	query := `INSERT INTO reviews (id, tutor_id, student_id, booking_id, rating, comment, subject, created_at)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	// For the review ID, we'll let the DB generate it with gen_random_uuid()
	// but for the response, we need to generate it here first.
	review.ID = uuid.NewString()  // You'll need to import "github.com/google/uuid"
	review.CreatedAt = time.Now() // You'll need "time"

	_, err := r.db.Exec(ctx, query,
		review.ID,
		review.TutorId, // We'll need to add this field to our domain struct
		review.StudentId,
		review.BookingID, // And this one too
		review.Rating,
		review.Comment,
		review.Subject,
		review.CreatedAt,
	)
	return err
}
