-- Удаление индексов (если требуется явно)
DROP INDEX IF EXISTS lessons_scheduled_starts_idx;
DROP INDEX IF EXISTS outbox_status_created_idx;
DROP INDEX IF EXISTS messages_conv_created_idx;
DROP INDEX IF EXISTS conversations_pair_uniq;
DROP INDEX IF EXISTS wallet_entries_wallet_created_idx;
DROP INDEX IF EXISTS payments_student_created_idx;
DROP INDEX IF EXISTS lessons_tutor_status_idx;
DROP INDEX IF EXISTS bookings_tutor_starts_idx;
DROP INDEX IF EXISTS availability_slots_tutor_starts_idx;

-- Очистка сидов
DELETE FROM subjects WHERE slug IN ('english','math','kazakh');
DELETE FROM languages WHERE code IN ('ru','en','kk');

-- Триггеры
DROP TRIGGER IF EXISTS trg_users_updated ON users;
DROP TRIGGER IF EXISTS trg_student_profiles_updated ON student_profiles;
DROP TRIGGER IF EXISTS trg_tutor_profiles_updated ON tutor_profiles;
DROP TRIGGER IF EXISTS trg_availability_slots_updated ON availability_slots;
DROP TRIGGER IF EXISTS trg_bookings_updated ON bookings;
DROP TRIGGER IF EXISTS trg_lessons_updated ON lessons;
DROP TRIGGER IF EXISTS trg_payments_updated ON payments;
DROP TRIGGER IF EXISTS trg_wallet_entries_updated ON wallet_entries;
DROP TRIGGER IF EXISTS trg_payouts_updated ON payouts;

-- Функция
DROP FUNCTION IF EXISTS set_updated_at();