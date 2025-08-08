package domain

import "time"

type Notifications struct {
	Lessons   bool `json:"lessons"`
	Messages  bool `json:"messages"`
	Reminders bool `json:"reminders"`
}

type User struct {
	ID             string        `json:"id"`
	Email          string        `json:"email"`
	Name           string        `json:"name"`
	Role           string        `json:"role"` // 'student' or 'tutor'
	Avatar         string        `json:"avatar,omitempty"`
	Age            int           `json:"age,omitempty"`
	LearningGoals  []string      `json:"learningGoals,omitempty"`
	Description    string        `json:"description,omitempty"`
	Notifications  Notifications `json:"notifications"`
	CreatedAt      time.Time     `json:"createdAt"`
	Level          string        `json:"level,omitempty"`
	NativeLanguage string        `json:"nativeLanguage,omitempty"`
	Budget         float64       `json:"budget,omitempty"`
	Availability   []string      `json:"availability,omitempty"`
}

type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}
