package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type TokenUseCase interface {
	ParseToken(ctx context.Context, token string) (*Claims, error)
}

type tokenUseCase struct{ secret string }

func NewTokenUseCase(secret string) TokenUseCase { return &tokenUseCase{secret: secret} }

// Оборачиваем ошибки в свои, чтобы middleware мог отличать причины.
var (
	ErrExpired          = fmt.Errorf("token expired: %w", jwt.ErrTokenExpired)
	ErrSignatureInvalid = fmt.Errorf("token signature invalid: %w", jwt.ErrTokenSignatureInvalid)
	ErrMalformed        = fmt.Errorf("token malformed: %w", jwt.ErrTokenMalformed)
)

func IsTokenExpired(err error) bool {
	return errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, ErrExpired)
}
func IsTokenSignatureInvalid(err error) bool {
	return errors.Is(err, jwt.ErrTokenSignatureInvalid) || errors.Is(err, ErrSignatureInvalid)
}
func IsTokenMalformed(err error) bool {
	return errors.Is(err, jwt.ErrTokenMalformed) || errors.Is(err, ErrMalformed)
}

func (uc *tokenUseCase) ParseToken(ctx context.Context, t string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(t, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Поймаем подмену алгоритма (alg)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method=%v: %w", token.Header["alg"], jwt.ErrTokenUnverifiable)
		}
		return []byte(uc.secret), nil
	})
	if err != nil {
		// Классифицируем
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpired
		}
		if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
			return nil, ErrSignatureInvalid
		}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, ErrMalformed
		}
		// Прочие проблемы верификации
		return nil, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token: claims not valid")
	}
	return claims, nil
}
