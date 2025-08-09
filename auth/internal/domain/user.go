package domain

import "time"

type Role string

const (
	StudentRole Role = "student"
	TutorRole   Role = "tutor"
	AdminRole   Role = "admin"
)

type User struct {
	ID              string     `json:"id"`
	Email           string     `json:"email"`
	Phone           string     `json:"phone"`
	FirstName       string     `json:"firstName"`
	LastName        string     `json:"lastName"`
	Role            Role       `json:"role"`
	CreatedAt       time.Time  `json:"createdAt"`
	EmailVerifiedAt *time.Time `json:"-"`
	PhoneVerifiedAt *time.Time `json:"-"`
}
