package domain

type Education struct {
	Degree      string `json:"degree"`
	Institution string `json:"institution"`
	Year        string `json:"year"`
	Field       string `json:"field"`
}

// Certification represents a single certificate.
type Certification struct {
	Name   string `json:"name"`
	Issuer string `json:"issuer"`
	Year   string `json:"year"`
}

// Tutor represents the full, detailed information for a single tutor.
type Tutor struct {
	ID             string          `json:"id"`
	Name           string          `json:"name"`
	Avatar         string          `json:"avatar"`
	Subjects       []string        `json:"subjects"`
	Languages      []string        `json:"languages"`
	Rating         float64         `json:"rating"`
	ReviewCount    int             `json:"reviewCount"`
	Price          float64         `json:"price"`
	Description    string          `json:"description"`
	Country        string          `json:"country"`
	IsOnline       bool            `json:"isOnline"`
	VideoIntro     string          `json:"videoIntro,omitempty"`
	TeachingStyle  string          `json:"teachingStyle,omitempty"`
	Education      []Education     `json:"education,omitempty"`
	Certifications []Certification `json:"certifications,omitempty"`
	// Availability can be a map[string]interface{} or a more specific struct
	Availability map[string]interface{} `json:"availability,omitempty"`
}

// Pagination хранит информацию для постраничного вывода.
type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}
