ALTER TABLE tutors
    ADD COLUMN video_intro TEXT,
ADD COLUMN teaching_style TEXT,
ADD COLUMN education JSONB,
ADD COLUMN certifications JSONB,
ADD COLUMN availability JSONB;