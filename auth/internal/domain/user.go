package domain

import (
	"time"
)

// Role определяет роль пользователя в системе.
type Role string

const (
	StudentRole Role = "student"
	TutorRole   Role = "tutor"
)

// User представляет модель пользователя в системе.
type User struct {
	ID         string    `json:"id"`
	Email      string    `json:"email"`
	Password   string    `json:"-"` // Пароль не должен возвращаться в JSON
	Name       string    `json:"name"`
	Role       Role      `json:"role"`
	IsVerified bool      `json:"isVerified"`
	CreatedAt  time.Time `json:"createdAt"`
}
