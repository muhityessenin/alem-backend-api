CREATE TABLE IF NOT EXISTS reviews (
                                       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tutor_id UUID NOT NULL REFERENCES tutors(id) ON DELETE CASCADE,
    student_id UUID NOT NULL, -- We assume student IDs come from the user service
    booking_id UUID NOT NULL UNIQUE, -- Each booking can only be reviewed once
    rating INT NOT NULL CHECK (rating >= 1 AND rating <= 5),
    comment TEXT,
    subject TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

CREATE INDEX idx_reviews_tutor_id ON reviews(tutor_id);