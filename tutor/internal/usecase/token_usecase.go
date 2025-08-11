package usecase

import (
	"context"
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

func (uc *tokenUseCase) ParseToken(ctx context.Context, t string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(t, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(uc.secret), nil
	})
	if err != nil {
		return nil, err
	}
	if c, ok := token.Claims.(*Claims); ok && token.Valid {
		return c, nil
	}
	return nil, fmt.Errorf("invalid token")
}
