CREATE EXTENSION IF NOT EXISTS "uuid-ossp"; -- Может понадобиться для генерации UUID на уровне БД


CREATE TABLE IF NOT EXISTS tutors (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    name TEXT NOT NULL,
    avatar TEXT,
    country TEXT,
    price NUMERIC(10, 2) DEFAULT 0.00,
    rating NUMERIC(3, 2) DEFAULT 0.00,
    review_count INT DEFAULT 0,
    description TEXT,
    subjects TEXT[],
    languages TEXT[],
    is_online BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ
);