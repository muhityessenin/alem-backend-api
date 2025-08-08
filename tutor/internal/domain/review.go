package domain

import "time"

// Review represents a single review left by a student.
type Review struct {
	ID            string    `json:"id"`
	TutorId       string    `json:"tutorId,omitempty"` // <-- Add this field
	BookingID     string    `json:"bookingId,omitempty"`
	StudentId     string    `json:"studentId"`
	StudentName   string    `json:"studentName"`
	StudentAvatar string    `json:"studentAvatar"`
	Rating        int       `json:"rating"`
	Comment       string    `json:"comment"`
	Subject       string    `json:"subject"`
	CreatedAt     time.Time `json:"createdAt"`
}

// ReviewStats holds aggregated data about all reviews for a tutor.
type ReviewStats struct {
	AverageRating      float64        `json:"averageRating"`
	RatingDistribution map[string]int `json:"ratingDistribution"` // e.g., "5": 10, "4": 2
}
