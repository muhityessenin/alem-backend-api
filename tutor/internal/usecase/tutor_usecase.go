package usecase

import (
	"context"
	"math"

	"tutor/internal/domain"
	"tutor/internal/repository"
)

type TutorUseCase interface {
	// Wizard steps
	UpsertAbout(ctx context.Context, userID, first, last, phone, gender, avatar string, langs []domain.TutorLanguage) error
	ReplaceAvailability(ctx context.Context, userID, timezone string, days []domain.AvailabilityDay) error
	UpsertEducation(ctx context.Context, userID string, bio string, education []domain.Education, certs []domain.Certification) error
	ReplaceSubjects(ctx context.Context, userID string, items []domain.TutorSubjectDTO, regular map[string]int64, trial map[string]int64) error
	SetVideo(ctx context.Context, userID, videoURL string) error
	Complete(ctx context.Context, userID string) error

	// Queries
	List(ctx context.Context, filters map[string]string, page, limit int) ([]domain.TutorProfile, *domain.Pagination, error)
	GetDetails(ctx context.Context, id string) (*domain.TutorProfile, error)
}

type tutorUseCase struct {
	repo repository.TutorRepository
}

func NewTutorUseCase(repo repository.TutorRepository) TutorUseCase {
	return &tutorUseCase{repo: repo}
}

// Steps
func (uc *tutorUseCase) UpsertAbout(ctx context.Context, userID, first, last, phone, gender, avatar string, langs []domain.TutorLanguage) error {
	return uc.repo.UpsertAbout(ctx, userID, first, last, phone, gender, avatar, langs)
}
func (uc *tutorUseCase) ReplaceAvailability(ctx context.Context, userID, tz string, days []domain.AvailabilityDay) error {
	return uc.repo.ReplaceWeeklyAvailability(ctx, userID, tz, days)
}
func (uc *tutorUseCase) UpsertEducation(ctx context.Context, userID string, bio string, education []domain.Education, certs []domain.Certification) error {
	return uc.repo.UpsertEducation(ctx, userID, bio, education, certs)
}
func (uc *tutorUseCase) ReplaceSubjects(ctx context.Context, userID string, items []domain.TutorSubjectDTO, regular map[string]int64, trial map[string]int64) error {
	return uc.repo.ReplaceSubjects(ctx, userID, items, regular, trial)
}
func (uc *tutorUseCase) SetVideo(ctx context.Context, userID, videoURL string) error {
	return uc.repo.SetVideo(ctx, userID, videoURL)
}
func (uc *tutorUseCase) Complete(ctx context.Context, userID string) error {
	return uc.repo.MarkCompleted(ctx, userID)
}

// Queries
func (uc *tutorUseCase) List(ctx context.Context, filters map[string]string, page, limit int) ([]domain.TutorProfile, *domain.Pagination, error) {
	list, total, err := uc.repo.FindTutorCardList(ctx, filters, page, limit)
	if err != nil {
		return nil, nil, err
	}
	p := &domain.Pagination{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: int(math.Ceil(float64(total) / float64(limit))),
	}
	return list, p, nil
}
func (uc *tutorUseCase) GetDetails(ctx context.Context, id string) (*domain.TutorProfile, error) {
	return uc.repo.FindTutorDetails(ctx, id)
}
