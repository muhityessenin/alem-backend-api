package usecase

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
)

// Claims для JWT токена (должны быть идентичны в обоих сервисах)
type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type TokenUseCase interface {
	ParseToken(ctx context.Context, token string) (*Claims, error)
}

type tokenUseCase struct {
	jwtSecret string
}

func NewTokenUseCase(secret string) TokenUseCase {
	return &tokenUseCase{jwtSecret: secret}
}

func (uc *tokenUseCase) ParseToken(ctx context.Context, accessToken string) (*Claims, error) {
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
