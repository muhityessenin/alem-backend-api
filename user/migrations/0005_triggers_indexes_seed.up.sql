-- функция обновления updated_at
CREATE OR REPLACE FUNCTION set_updated_at() RETURNS trigger AS $$
BEGIN
  NEW.updated_at := now();
RETURN NEW;
END;$$ LANGUAGE plpgsql;

-- пример подключения триггеров
CREATE TRIGGER trg_users_updated BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_student_profiles_updated BEFORE UPDATE ON student_profiles FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_tutor_profiles_updated BEFORE UPDATE ON tutor_profiles FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_availability_slots_updated BEFORE UPDATE ON availability_slots FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_bookings_updated BEFORE UPDATE ON bookings FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_lessons_updated BEFORE UPDATE ON lessons FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_payments_updated BEFORE UPDATE ON payments FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_wallet_entries_updated BEFORE UPDATE ON wallet_entries FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_payouts_updated BEFORE UPDATE ON payouts FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- начальные справочники
INSERT INTO languages(code,name) VALUES
    ('ru','Русский') ON CONFLICT DO NOTHING,
  ('en','English') ON CONFLICT DO NOTHING,
    ('kk','Қазақ тілі') ON CONFLICT DO NOTHING;

INSERT INTO subjects(slug,name) VALUES
                                    ('english', '{"ru":"Английский","en":"English"}'),
                                    ('math',    '{"ru":"Математика","en":"Mathematics"}'),
                                    ('kazakh',  '{"ru":"Казахский","en":"Kazakh"}')
    ON CONFLICT DO NOTHING;

-- полезные индексы
CREATE INDEX IF NOT EXISTS lessons_scheduled_starts_idx ON lessons (started_at) WHERE status='scheduled';