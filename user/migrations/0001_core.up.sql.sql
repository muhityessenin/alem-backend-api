CREATE EXTENSION IF NOT EXISTS pgcrypto; -- gen_random_uuid
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE users (
                       id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
                       email             citext UNIQUE NOT NULL,
                       phone_e164        text UNIQUE,
                       password_hash     text,
                       role              text NOT NULL CHECK (role IN ('student','tutor','admin')),
                       locale            text NOT NULL DEFAULT 'ru',
                       country_code      text,
                       status            text NOT NULL DEFAULT 'active' CHECK (status IN ('active','blocked','pending')),
                       created_at        timestamptz NOT NULL DEFAULT now(),
                       updated_at        timestamptz NOT NULL DEFAULT now(),
                       deleted_at        timestamptz
);

CREATE TABLE student_profiles (
                                  user_id           uuid PRIMARY KEY REFERENCES users(id),
                                  display_name      text,
                                  birthdate         date,
                                  prefs             jsonb NOT NULL DEFAULT '{}'::jsonb,
                                  created_at        timestamptz NOT NULL DEFAULT now(),
                                  updated_at        timestamptz NOT NULL DEFAULT now(),
                                  deleted_at        timestamptz
);

CREATE TABLE tutor_profiles (
                                user_id           uuid PRIMARY KEY REFERENCES users(id),
                                headline          text,
                                bio               text,
                                hourly_rate_minor bigint NOT NULL,
                                currency          char(3) NOT NULL DEFAULT 'KZT',
                                video_url         text,
                                years_experience  int,
                                verification      text NOT NULL DEFAULT 'pending' CHECK (verification IN ('pending','verified','rejected')),
                                rating_avg        numeric(3,2) NOT NULL DEFAULT 0,
                                rating_count      int NOT NULL DEFAULT 0,
                                props             jsonb NOT NULL DEFAULT '{}'::jsonb,
                                created_at        timestamptz NOT NULL DEFAULT now(),
                                updated_at        timestamptz NOT NULL DEFAULT now(),
                                deleted_at        timestamptz
);