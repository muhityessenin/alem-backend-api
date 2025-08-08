package usecase

import (
	"context"
	"math"
	"user/internal/domain"
	"user/internal/repository"
)

type UserUseCase interface {
	GetProfile(ctx context.Context, id string) (*domain.User, error)
	UpdateProfile(ctx context.Context, user *domain.User) (*domain.User, error)
	UpdateAvatar(ctx context.Context, userID, avatarURL string) error
	ListStudents(ctx context.Context, page, limit int) ([]domain.User, *domain.Pagination, error)
}

type userUseCase struct {
	userRepo repository.UserRepository
}

func NewUserUseCase(repo repository.UserRepository) UserUseCase {
	return &userUseCase{userRepo: repo}
}

func (uc *userUseCase) GetProfile(ctx context.Context, id string) (*domain.User, error) {
	return uc.userRepo.GetProfileByID(ctx, id)
}

func (uc *userUseCase) UpdateProfile(ctx context.Context, user *domain.User) (*domain.User, error) {
	// Сначала нужно получить текущий профиль, чтобы не затереть поля,
	// которые не пришли в запросе.
	currentUser, err := uc.userRepo.GetProfileByID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	// Обновляем поля, если они были переданы в запросе.
	// Это предотвращает сброс полей на пустые значения.
	if user.Name != "" {
		currentUser.Name = user.Name
	}
	if user.Age != 0 {
		currentUser.Age = user.Age
	}
	if user.Avatar != "" {
		currentUser.Avatar = user.Avatar
	}
	if len(user.LearningGoals) > 0 {
		currentUser.LearningGoals = user.LearningGoals
	}
	if user.Description != "" {
		currentUser.Description = user.Description
	}
	// Для Notifications можно просто присвоить, т.к. фронтенд обычно присылает весь объект.
	currentUser.Notifications = user.Notifications

	// Вызываем репозиторий для сохранения обновленных данных.
	if err := uc.userRepo.UpdateProfile(ctx, currentUser); err != nil {
		return nil, err
	}

	// Возвращаем полностью обновленный профиль.
	return currentUser, nil
}

func (uc *userUseCase) UpdateAvatar(ctx context.Context, userID, avatarURL string) error {
	return uc.userRepo.UpdateAvatarURL(ctx, userID, avatarURL)
}

func (uc *userUseCase) ListStudents(ctx context.Context, page, limit int) ([]domain.User, *domain.Pagination, error) {
	students, total, err := uc.userRepo.FindStudents(ctx, page, limit)
	if err != nil {
		return nil, nil, err
	}
	pagination := &domain.Pagination{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: int(math.Ceil(float64(total) / float64(limit))),
	}
	return students, pagination, nil
}
