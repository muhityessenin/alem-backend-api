-- 11.4 Google Calendar
DROP TABLE IF EXISTS calendar_events CASCADE;
DROP TABLE IF EXISTS calendar_accounts CASCADE;
DROP INDEX IF EXISTS oauth_tokens_user_provider_uniq;
DROP TABLE IF EXISTS oauth_tokens CASCADE;

-- 11.3 Универсальные выплаты (откат к старой схеме)
ALTER TABLE payouts DROP CONSTRAINT IF EXISTS payouts_owner_ck;
-- Данные в owner_type/owner_id останутся; при желании можно их очистить
ALTER TABLE payouts DROP COLUMN IF EXISTS owner_id;
ALTER TABLE payouts DROP COLUMN IF EXISTS owner_type;
ALTER TABLE payouts DROP COLUMN IF EXISTS legacy_tutor_id;

-- 11.2 Запросы на чат
DROP INDEX IF EXISTS conversations_pending_idx;
ALTER TABLE conversations DROP COLUMN IF EXISTS approved_at;
ALTER TABLE conversations DROP COLUMN IF EXISTS initiator_id;
ALTER TABLE conversations DROP COLUMN IF EXISTS status;

-- 11.1 Таксономия
DROP INDEX IF EXISTS lessons_subdir_idx;
DROP INDEX IF EXISTS bookings_subdir_idx;
ALTER TABLE lessons  DROP COLUMN IF EXISTS subdirection_id;
ALTER TABLE bookings DROP COLUMN IF EXISTS subdirection_id;
DROP TABLE IF EXISTS tutor_subdirections CASCADE;
DROP TABLE IF EXISTS subdirections CASCADE;
DROP TABLE IF EXISTS directions CASCADE;