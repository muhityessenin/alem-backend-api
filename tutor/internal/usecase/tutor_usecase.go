package usecase

import (
	"context"
	"math"
	"tutor/internal/domain"
	"tutor/internal/repository"
)

type TutorUseCase interface {
	List(ctx context.Context, filters map[string]string, page, limit int) ([]domain.Tutor, *domain.Pagination, error)
	GetDetails(ctx context.Context, id string) (*domain.Tutor, error)
	UpdateProfile(ctx context.Context, tutor *domain.Tutor) (*domain.Tutor, error) // <-- Add method
	ModerateProfile(ctx context.Context, tutorID, status, reason string) error
	ListPending(ctx context.Context, page, limit int) ([]domain.Tutor, *domain.Pagination, error)
}

type tutorUseCase struct {
	repo repository.TutorRepository
}

func (uc *tutorUseCase) ListPending(ctx context.Context, page, limit int) ([]domain.Tutor, *domain.Pagination, error) {
	tutors, total, err := uc.repo.FindPending(ctx, page, limit)
	if err != nil {
		return nil, nil, err
	}

	pagination := &domain.Pagination{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: int(math.Ceil(float64(total) / float64(limit))),
	}

	return tutors, pagination, nil
}
func NewTutorUseCase(repo repository.TutorRepository) TutorUseCase {
	return &tutorUseCase{repo: repo}
}

func (uc *tutorUseCase) ModerateProfile(ctx context.Context, tutorID, status, reason string) error {
	return uc.repo.UpdateStatus(ctx, tutorID, status, reason)
}

func (uc *tutorUseCase) List(ctx context.Context, filters map[string]string, page, limit int) ([]domain.Tutor, *domain.Pagination, error) {
	tutors, total, err := uc.repo.FindAll(ctx, filters, page, limit)
	if err != nil {
		return nil, nil, err
	}

	pagination := &domain.Pagination{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: int(math.Ceil(float64(total) / float64(limit))),
	}

	return tutors, pagination, nil
}

func (uc *tutorUseCase) GetDetails(ctx context.Context, id string) (*domain.Tutor, error) {
	return uc.repo.FindByID(ctx, id)
}

func (uc *tutorUseCase) UpdateProfile(ctx context.Context, tutor *domain.Tutor) (*domain.Tutor, error) {
	// In a real app, you might merge fields instead of overwriting,
	// but for this API, overwriting is acceptable.
	if err := uc.repo.UpdateProfile(ctx, tutor); err != nil {
		return nil, err
	}
	// Return the updated profile by fetching it again
	return uc.repo.FindByID(ctx, tutor.ID)
}
