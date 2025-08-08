package usecase

import (
	"auth/internal/domain"
	"auth/internal/repository"
	"context"
	"errors"
	"fmt"
	_ "fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"time"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid email or password")
)

type Claims struct {
	UserID string      `json:"user_id"`
	Role   domain.Role `json:"role"`
	jwt.RegisteredClaims
}

type AuthUseCase interface {
	Register(ctx context.Context, name, email, password string, role domain.Role) (*domain.User, string, error)
	Login(ctx context.Context, email, password string) (accessToken, refreshToken string, err error)
	ParseToken(ctx context.Context, token string) (*Claims, error)
	RefreshTokens(ctx context.Context, refreshToken string) (string, string, error)
}

type authUseCase struct {
	userRepo        repository.UserRepository
	jwtSecret       string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewAuthUseCase(userRepo repository.UserRepository, secret string, accessTTL, refreshTTL time.Duration) AuthUseCase {
	return &authUseCase{
		userRepo:        userRepo,
		jwtSecret:       secret,
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
	}
}
func (uc *authUseCase) Login(ctx context.Context, email, password string) (string, string, error) {
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Здесь можно проверить на pgx.ErrNoRows и вернуть нашу кастомную ошибку
		return "", "", ErrUserNotFound
	}

	// Сравниваем пароль из запроса с хешем в базе
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", "", ErrInvalidCredentials
	}

	// Генерируем Access Token
	accessToken, err := uc.generateToken(user.ID, user.Role, uc.accessTokenTTL)
	if err != nil {
		return "", "", err
	}

	// Генерируем Refresh Token
	refreshToken, err := uc.generateToken(user.ID, user.Role, uc.refreshTokenTTL)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// Вспомогательный метод для генерации токена
func (uc *authUseCase) generateToken(userID string, role domain.Role, ttl time.Duration) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(uc.jwtSecret))
}
func (uc *authUseCase) Register(ctx context.Context, name, email, password string, role domain.Role) (*domain.User, string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	user := &domain.User{
		ID:         uuid.NewString(),
		Name:       name,
		Email:      email,
		Password:   string(hashedPassword),
		Role:       role,
		IsVerified: false,
		CreatedAt:  time.Now(),
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, "", err
	}

	accessToken, err := uc.generateToken(user.ID, user.Role, uc.accessTokenTTL)
	if err != nil {
		return nil, "", err
	}

	return user, accessToken, nil
}

func (uc *authUseCase) ParseToken(ctx context.Context, accessToken string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(accessToken, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(uc.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func (uc *authUseCase) RefreshTokens(ctx context.Context, refreshToken string) (string, string, error) {
	claims, err := uc.ParseToken(ctx, refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("invalid refresh token: %w", err)
	}

	newAccessToken, err := uc.generateToken(claims.UserID, claims.Role, uc.accessTokenTTL)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err := uc.generateToken(claims.UserID, claims.Role, uc.refreshTokenTTL)
	if err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}
