-- extensions
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS citext;

-- users & profiles
-- (смотри раздел 3.1)
-- Вставлено целиком для автономности миграции
CREATE TABLE users (
                       id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
                       email citext UNIQUE NOT NULL,
                       phone_e164 text UNIQUE,
                       password_hash text,
                       role text NOT NULL CHECK (role IN ('student','tutor','admin')),
                       locale text NOT NULL DEFAULT 'ru',
                       country_code text,
                       status text NOT NULL DEFAULT 'active' CHECK (status IN ('active','blocked','pending')),
                       created_at timestamptz NOT NULL DEFAULT now(),
                       updated_at timestamptz NOT NULL DEFAULT now(),
                       deleted_at timestamptz
);

CREATE TABLE student_profiles (
                                  user_id uuid PRIMARY KEY REFERENCES users(id),
                                  display_name text,
                                  birthdate date,
                                  prefs jsonb NOT NULL DEFAULT '{}'::jsonb,
                                  created_at timestamptz NOT NULL DEFAULT now(),
                                  updated_at timestamptz NOT NULL DEFAULT now(),
                                  deleted_at timestamptz
);

CREATE TABLE tutor_profiles (
                                user_id uuid PRIMARY KEY REFERENCES users(id),
                                headline text,
                                bio text,
                                hourly_rate_minor bigint NOT NULL,
                                currency char(3) NOT NULL DEFAULT 'KZT',
                                video_url text,
                                years_experience int,
                                verification text NOT NULL DEFAULT 'pending' CHECK (verification IN ('pending','verified','rejected')),
                                rating_avg numeric(3,2) NOT NULL DEFAULT 0,
                                rating_count int NOT NULL DEFAULT 0,
                                props jsonb NOT NULL DEFAULT '{}'::jsonb,
                                created_at timestamptz NOT NULL DEFAULT now(),
                                updated_at timestamptz NOT NULL DEFAULT now(),
                                deleted_at timestamptz
);

-- taxonomy
CREATE TABLE subjects (
                          id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
                          slug text UNIQUE NOT NULL,
                          name jsonb NOT NULL,
                          created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE languages (
                           code text PRIMARY KEY,
                           name text NOT NULL
);

CREATE TABLE tutor_subjects (
                                tutor_id uuid NOT NULL REFERENCES tutor_profiles(user_id),
                                subject_id uuid NOT NULL REFERENCES subjects(id),
                                level text,
                                price_minor bigint,
                                currency char(3) DEFAULT 'KZT',
                                PRIMARY KEY (tutor_id, subject_id)
);

CREATE TABLE tutor_languages (
                                 tutor_id uuid NOT NULL REFERENCES tutor_profiles(user_id),
                                 lang_code text NOT NULL REFERENCES languages(code),
                                 proficiency text NOT NULL CHECK (proficiency IN ('A1','A2','B1','B2','C1','C2','native')),
                                 PRIMARY KEY (tutor_id, lang_code)
);