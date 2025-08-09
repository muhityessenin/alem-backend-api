-- оставляем как у тебя
ALTER TABLE users ADD COLUMN first_name text;
ALTER TABLE users ADD COLUMN last_name text;
ALTER TABLE users ADD COLUMN email_verified_at timestamptz;
ALTER TABLE users ADD COLUMN phone_verified_at timestamptz;

-- ============================================
-- auth_sessions — для refresh-сессий с ревокацией
-- ============================================
CREATE TABLE auth_sessions (
                               id                 uuid PRIMARY KEY DEFAULT gen_random_uuid(),
                               user_id            uuid NOT NULL REFERENCES users(id),
                               refresh_token_hash text NOT NULL,
                               expires_at         timestamptz NOT NULL,
                               user_agent         text,
                               ip                 text,
                               created_at         timestamptz NOT NULL DEFAULT now(),
                               revoked_at         timestamptz
);

-- быстрый поиск/проверка refresh по хешу и пользователю
CREATE UNIQUE INDEX auth_sessions_refresh_hash_uniq ON auth_sessions (refresh_token_hash);
CREATE INDEX auth_sessions_user_id_idx ON auth_sessions (user_id);
CREATE INDEX auth_sessions_expires_idx ON auth_sessions (expires_at);

-- ============================================
-- otp_codes — для входа/верификации по телефону
-- ============================================
CREATE TABLE otp_codes (
                           phone_e164    text NOT NULL,
                           code_hash     text NOT NULL,
                           purpose       text NOT NULL CHECK (purpose IN ('login','verify')),
                           expires_at    timestamptz NOT NULL,
                           attempts_used int NOT NULL DEFAULT 0,
                           created_at    timestamptz NOT NULL DEFAULT now(),
                           consumed_at   timestamptz,

    -- один активный код на телефон+назначение (для UPSERT по (phone_e164, purpose))
                           CONSTRAINT otp_phone_purpose_uniq UNIQUE (phone_e164, purpose)
);

CREATE INDEX otp_expires_idx ON otp_codes (expires_at);
