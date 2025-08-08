ALTER TABLE users
ADD COLUMN level TEXT,             -- e.g., 'Beginner', 'Intermediate'
ADD COLUMN native_language TEXT,
ADD COLUMN budget NUMERIC(10, 2),
ADD COLUMN availability TEXT[];