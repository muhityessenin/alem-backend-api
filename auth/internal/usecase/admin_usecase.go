package usecase

import (
	"auth/internal/repository"
	"context"
)

type AdminUseCase interface {
	CreateLanguage(ctx context.Context, code, name string) error
	ListLanguages(ctx context.Context) ([]repository.Language, error)

	CreateSubject(ctx context.Context, slug string, name map[string]any) error
	ListSubjects(ctx context.Context) ([]repository.Subject, error)

	CreateDirection(ctx context.Context, slug string, name map[string]any) error
	ListDirections(ctx context.Context) ([]repository.Direction, error)

	CreateSubdirection(ctx context.Context, directionID, directionSlug, slug string, name map[string]any) error
	ListSubdirections(ctx context.Context, directionSlug string) ([]repository.Subdirection, error)

	UpsertTutorSubject(ctx context.Context, tutorID, subjectID, subjectSlug, level string, price int64, currency string) error
	ListTutorSubjects(ctx context.Context, tutorID string) ([]repository.TutorSubjectView, error)

	UpsertTutorLanguage(ctx context.Context, tutorID, code, proficiency string) error
	ListTutorLanguages(ctx context.Context, tutorID string) ([]repository.TutorLanguageView, error)

	UpsertTutorSubdirection(ctx context.Context, tutorID, subdirID, subdirSlug, level string, price int64, currency string) error
	ListTutorSubdirections(ctx context.Context, tutorID string) ([]repository.TutorSubdirectionView, error)
}

type adminUseCase struct{ repo repository.AdminRepository }

func NewAdminUseCase(r repository.AdminRepository) AdminUseCase { return &adminUseCase{repo: r} }

func (uc *adminUseCase) CreateLanguage(ctx context.Context, code, name string) error {
	return uc.repo.CreateLanguage(ctx, code, name)
}
func (uc *adminUseCase) ListLanguages(ctx context.Context) ([]repository.Language, error) {
	return uc.repo.ListLanguages(ctx)
}

func (uc *adminUseCase) CreateSubject(ctx context.Context, slug string, name map[string]any) error {
	return uc.repo.CreateSubject(ctx, slug, name)
}
func (uc *adminUseCase) ListSubjects(ctx context.Context) ([]repository.Subject, error) {
	return uc.repo.ListSubjects(ctx)
}

func (uc *adminUseCase) CreateDirection(ctx context.Context, slug string, name map[string]any) error {
	return uc.repo.CreateDirection(ctx, slug, name)
}
func (uc *adminUseCase) ListDirections(ctx context.Context) ([]repository.Direction, error) {
	return uc.repo.ListDirections(ctx)
}

func (uc *adminUseCase) CreateSubdirection(ctx context.Context, directionID, directionSlug, slug string, name map[string]any) error {
	return uc.repo.CreateSubdirection(ctx, directionID, directionSlug, slug, name)
}
func (uc *adminUseCase) ListSubdirections(ctx context.Context, directionSlug string) ([]repository.Subdirection, error) {
	return uc.repo.ListSubdirections(ctx, directionSlug)
}

func (uc *adminUseCase) UpsertTutorSubject(ctx context.Context, tutorID, subjectID, subjectSlug, level string, price int64, currency string) error {
	return uc.repo.UpsertTutorSubject(ctx, tutorID, subjectID, subjectSlug, level, price, currency)
}
func (uc *adminUseCase) ListTutorSubjects(ctx context.Context, tutorID string) ([]repository.TutorSubjectView, error) {
	return uc.repo.ListTutorSubjects(ctx, tutorID)
}

func (uc *adminUseCase) UpsertTutorLanguage(ctx context.Context, tutorID, code, proficiency string) error {
	return uc.repo.UpsertTutorLanguage(ctx, tutorID, code, proficiency)
}
func (uc *adminUseCase) ListTutorLanguages(ctx context.Context, tutorID string) ([]repository.TutorLanguageView, error) {
	return uc.repo.ListTutorLanguages(ctx, tutorID)
}

func (uc *adminUseCase) UpsertTutorSubdirection(ctx context.Context, tutorID, subdirID, subdirSlug, level string, price int64, currency string) error {
	return uc.repo.UpsertTutorSubdirection(ctx, tutorID, subdirID, subdirSlug, level, price, currency)
}
func (uc *adminUseCase) ListTutorSubdirections(ctx context.Context, tutorID string) ([]repository.TutorSubdirectionView, error) {
	return uc.repo.ListTutorSubdirections(ctx, tutorID)
}
