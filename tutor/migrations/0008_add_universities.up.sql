-- extensions на случай чистой БД
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Университеты (как таксономия: slug + локализованное имя)
CREATE TABLE IF NOT EXISTS universities (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    slug         text UNIQUE NOT NULL,
    name         jsonb NOT NULL,
    country_code text,
    city         text,
    created_at   timestamptz NOT NULL DEFAULT now()
);

-- Полезные индексы
CREATE INDEX IF NOT EXISTS universities_country_idx ON universities (lower(country_code));
CREATE INDEX IF NOT EXISTS universities_city_idx    ON universities (lower(city));
