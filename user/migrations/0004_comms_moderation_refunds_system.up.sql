CREATE TABLE payments (
                          id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
                          booking_id uuid REFERENCES bookings(id),
                          lesson_id uuid REFERENCES lessons(id),
                          student_id uuid NOT NULL REFERENCES student_profiles(user_id),
                          amount_minor bigint NOT NULL,
                          currency char(3) NOT NULL,
                          provider text NOT NULL,
                          provider_intent_id text,
                          provider_payload jsonb NOT NULL DEFAULT '{}'::jsonb,
                          status text NOT NULL CHECK (status IN ('requires_action','authorized','captured','failed','refunded','voided')),
                          idempotency_key text UNIQUE,
                          created_at timestamptz NOT NULL DEFAULT now(),
                          updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS payments_student_created_idx ON payments (student_id, created_at DESC);

CREATE TABLE commissions (
                             id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
                             lesson_id uuid UNIQUE REFERENCES lessons(id),
                             commission_minor bigint NOT NULL,
                             tutor_earn_minor bigint NOT NULL,
                             currency char(3) NOT NULL,
                             rule_version text NOT NULL,
                             details jsonb NOT NULL DEFAULT '{}'::jsonb,
                             created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE wallets (
                         id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
                         owner_type text NOT NULL CHECK (owner_type IN ('tutor','student','platform')),
                         owner_id uuid NOT NULL,
                         currency char(3) NOT NULL,
                         UNIQUE(owner_type, owner_id, currency)
);

CREATE TABLE wallet_entries (
                                id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
                                wallet_id uuid NOT NULL REFERENCES wallets(id),
                                lesson_id uuid REFERENCES lessons(id),
                                payment_id uuid REFERENCES payments(id),
                                type text NOT NULL CHECK (type IN ('credit','debit')),
                                reason text NOT NULL,
                                amount_minor bigint NOT NULL,
                                balance_after bigint NOT NULL,
                                created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS wallet_entries_wallet_created_idx ON wallet_entries (wallet_id, created_at);

CREATE TABLE payouts (
                         id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
                         tutor_id uuid NOT NULL REFERENCES tutor_profiles(user_id),
                         amount_minor bigint NOT NULL,
                         currency char(3) NOT NULL,
                         method text NOT NULL,
                         destination jsonb NOT NULL,
                         status text NOT NULL CHECK (status IN ('requested','approved','processing','paid','failed','cancelled')),
                         requested_by uuid NOT NULL REFERENCES users(id),
                         approved_by uuid REFERENCES users(id),
                         provider_payload jsonb NOT NULL DEFAULT '{}'::jsonb,
                         idempotency_key text UNIQUE,
                         created_at timestamptz NOT NULL DEFAULT now(),
                         updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS payouts_tutor_created_idx ON payouts (tutor_id, created_at DESC);