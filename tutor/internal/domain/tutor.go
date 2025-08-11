package domain

import "time"

// “Классы” домена
type TutorProfile struct {
	UserID       string            `json:"userId"`
	FirstName    string            `json:"firstName"`
	LastName     string            `json:"lastName"`
	PhoneE164    string            `json:"phone"`
	Gender       string            `json:"gender,omitempty"`
	AvatarURL    string            `json:"avatar,omitempty"`
	Languages    []TutorLanguage   `json:"languages"`
	Bio          string            `json:"bio,omitempty"`
	VideoURL     string            `json:"videoUrl,omitempty"`
	Timezone     string            `json:"timezone,omitempty"`
	Prices       map[string]int64  `json:"prices,omitempty"` // subdirection_slug -> regular price (minor)
	TrialPrices  map[string]int64  `json:"trialPrices,omitempty"`
	Education    []Education       `json:"education,omitempty"`
	Certificates []Certification   `json:"certificates,omitempty"`
	Subjects     []TutorSubjectDTO `json:"subjects,omitempty"`
	RatingAvg    float32           `json:"ratingAvg"`
	RatingCount  int               `json:"ratingCount"`
	Verification string            `json:"verification"` // pending/verified/rejected
	CreatedAt    time.Time         `json:"createdAt"`
	UpdatedAt    time.Time         `json:"updatedAt"`
}

type TutorLanguage struct {
	Code        string `json:"code"`        // ru,en,kk
	Proficiency string `json:"proficiency"` // A1..C2/native
	DisplayName string `json:"displayName"` // опционально
}

type TutorSubjectDTO struct {
	SubjectID   string `json:"subjectId"`
	SubjectSlug string `json:"subjectSlug"`
	Level       string `json:"level,omitempty"`
	PriceMinor  int64  `json:"priceMinor"`
	Currency    string `json:"currency"` // KZT
}

type AvailabilityDay struct {
	Day   string   `json:"day"`   // Понедельник/...
	Slots []string `json:"slots"` // ["18:00","18:30",...]
}

type Education struct {
	Degree      string `json:"degree"`
	Institution string `json:"institution"`
	Year        string `json:"year"`
	Field       string `json:"field"`
}

type Certification struct {
	Name   string `json:"name"`
	Issuer string `json:"issuer"`
	Year   string `json:"year"`
}

type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}
