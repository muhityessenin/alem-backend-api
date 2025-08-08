package usecase

import (
	"context"
	"math"
	"tutor/internal/domain"
	"tutor/internal/repository"
)

type ReviewUseCase interface {
	ListByTutor(ctx context.Context, tutorID string, page, limit int) ([]domain.Review, *domain.Pagination, *domain.ReviewStats, error)
	Create(ctx context.Context, review *domain.Review) (*domain.Review, error) // <-- New method
}

type reviewUseCase struct {
	repo repository.ReviewRepository
}

func NewReviewUseCase(repo repository.ReviewRepository) ReviewUseCase {
	return &reviewUseCase{repo: repo}
}

func (uc *reviewUseCase) ListByTutor(ctx context.Context, tutorID string, page, limit int) ([]domain.Review, *domain.Pagination, *domain.ReviewStats, error) {
	reviews, total, err := uc.repo.FindByTutorID(ctx, tutorID, page, limit)
	if err != nil {
		return nil, nil, nil, err
	}

	stats, err := uc.repo.GetStatsByTutorID(ctx, tutorID)
	if err != nil {
		return nil, nil, nil, err
	}

	pagination := &domain.Pagination{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: int(math.Ceil(float64(total) / float64(limit))),
	}

	return reviews, pagination, stats, nil
}

func (uc *reviewUseCase) Create(ctx context.Context, review *domain.Review) (*domain.Review, error) {
	// Business logic checks can go here. For example:
	// 1. Check if the booking ID is valid.
	// 2. Check if the lesson was actually completed.
	// 3. Check if the user already reviewed this booking.

	if err := uc.repo.Create(ctx, review); err != nil {
		return nil, err
	}

	// The review object was populated with ID and CreatedAt in the repository
	return review, nil
}
