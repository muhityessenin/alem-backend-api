-- 11.1 Направления и поднаправления
CREATE TABLE IF NOT EXISTS directions (
                                          id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    slug text UNIQUE NOT NULL,
    name jsonb NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
    );

CREATE TABLE IF NOT EXISTS subdirections (
                                             id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    direction_id uuid NOT NULL REFERENCES directions(id) ON DELETE CASCADE,
    slug text UNIQUE NOT NULL,
    name jsonb NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
    );

CREATE TABLE IF NOT EXISTS tutor_subdirections (
                                                   tutor_id uuid NOT NULL REFERENCES tutor_profiles(user_id),
    subdirection_id uuid NOT NULL REFERENCES subdirections(id),
    level text,
    price_minor bigint,
    currency char(3) DEFAULT 'KZT',
    PRIMARY KEY (tutor_id, subdirection_id)
    );

ALTER TABLE IF EXISTS bookings ADD COLUMN IF NOT EXISTS subdirection_id uuid REFERENCES subdirections(id);
ALTER TABLE IF EXISTS lessons  ADD COLUMN IF NOT EXISTS subdirection_id uuid REFERENCES subdirections(id);
CREATE INDEX IF NOT EXISTS bookings_subdir_idx ON bookings (subdirection_id);
CREATE INDEX IF NOT EXISTS lessons_subdir_idx  ON lessons (subdirection_id);

-- 11.2 Запросы на чат
ALTER TABLE conversations
    ADD COLUMN IF NOT EXISTS status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','active','blocked')),
    ADD COLUMN IF NOT EXISTS initiator_id uuid REFERENCES users(id),
    ADD COLUMN IF NOT EXISTS approved_at timestamptz;
CREATE INDEX IF NOT EXISTS conversations_pending_idx ON conversations (tutor_id) WHERE status='pending';

-- 11.3 Универсальные выплаты
ALTER TABLE payouts
    ADD COLUMN IF NOT EXISTS owner_type text CHECK (owner_type IN ('tutor','student','platform')),
    ADD COLUMN IF NOT EXISTS owner_id uuid,
    ADD COLUMN IF NOT EXISTS legacy_tutor_id uuid;
UPDATE payouts SET owner_type='tutor', owner_id=tutor_id WHERE owner_id IS NULL AND tutor_id IS NOT NULL;

ALTER TABLE payouts
    ADD CONSTRAINT payouts_owner_ck CHECK (
        (owner_type='tutor'    AND owner_id IS NOT NULL) OR
        (owner_type='student'  AND owner_id IS NOT NULL) OR
        (owner_type='platform' AND owner_id IS NOT NULL)
        );

-- 11.4 Google Calendar
CREATE TABLE IF NOT EXISTS oauth_tokens (
                                            id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id),
    provider text NOT NULL CHECK (provider IN ('google')),
    access_token text NOT NULL,
    refresh_token text,
    expires_at timestamptz,
    scope text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
    );
CREATE UNIQUE INDEX IF NOT EXISTS oauth_tokens_user_provider_uniq ON oauth_tokens(user_id, provider);

CREATE TABLE IF NOT EXISTS calendar_accounts (
                                                 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id),
    provider text NOT NULL CHECK (provider IN ('google')),
    email citext,
    calendar_id text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
    );

CREATE TABLE IF NOT EXISTS calendar_events (
                                               id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    lesson_id uuid NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    account_id uuid NOT NULL REFERENCES calendar_accounts(id) ON DELETE CASCADE,
    provider text NOT NULL CHECK (provider IN ('google')),
    event_id text NOT NULL,
    html_link text,
    synced_at timestamptz,
    status text NOT NULL DEFAULT 'synced' CHECK (status IN ('synced','to_create','to_update','to_delete','error')),
    error_msg text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
    );
CREATE UNIQUE INDEX IF NOT EXISTS calendar_events_lesson_uniq ON calendar_events(lesson_id);