ALTER TABLE tutors
ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'pending',
ADD COLUMN rejection_reason TEXT;