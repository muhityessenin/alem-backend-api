CREATE TABLE availability_slots (
                                    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
                                    tutor_id uuid NOT NULL REFERENCES tutor_profiles(user_id),
                                    starts_at timestamptz NOT NULL,
                                    ends_at timestamptz NOT NULL,
                                    recurrence text,
                                    is_recurring boolean NOT NULL DEFAULT false,
                                    created_at timestamptz NOT NULL DEFAULT now(),
                                    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS availability_slots_tutor_starts_idx ON availability_slots (tutor_id, starts_at);

CREATE TABLE bookings (
                          id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
                          student_id uuid NOT NULL REFERENCES student_profiles(user_id),
                          tutor_id uuid NOT NULL REFERENCES tutor_profiles(user_id),
                          subject_id uuid REFERENCES subjects(id),
                          starts_at timestamptz NOT NULL,
                          ends_at timestamptz NOT NULL,
                          status text NOT NULL CHECK (status IN ('pending','awaiting_payment','confirmed','cancelled','expired')),
                          idempotency_key text UNIQUE,
                          metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
                          created_at timestamptz NOT NULL DEFAULT now(),
                          updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS bookings_tutor_starts_idx ON bookings (tutor_id, starts_at);

CREATE TABLE lessons (
                         id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
                         booking_id uuid UNIQUE REFERENCES bookings(id),
                         student_id uuid NOT NULL REFERENCES student_profiles(user_id),
                         tutor_id uuid NOT NULL REFERENCES tutor_profiles(user_id),
                         subject_id uuid REFERENCES subjects(id),
                         status text NOT NULL CHECK (status IN ('scheduled','in_progress','completed','no_show','cancelled')),
                         is_trial boolean NOT NULL DEFAULT false,
                         room_id text,
                         started_at timestamptz,
                         ended_at timestamptz,
                         duration_seconds int,
                         created_at timestamptz NOT NULL DEFAULT now(),
                         updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS lessons_tutor_status_idx ON lessons (tutor_id, status);