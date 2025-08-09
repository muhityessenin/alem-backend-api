CREATE TABLE conversations (
                               id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
                               student_id uuid NOT NULL REFERENCES student_profiles(user_id),
                               tutor_id uuid NOT NULL REFERENCES tutor_profiles(user_id),
                               created_at timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS conversations_pair_uniq ON conversations (student_id, tutor_id);

CREATE TABLE messages (
                          id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
                          conversation_id uuid NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
                          sender_id uuid NOT NULL REFERENCES users(id),
                          body text,
                          body_redacted text,
                          meta jsonb NOT NULL DEFAULT '{}'::jsonb,
                          created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS messages_conv_created_idx ON messages (conversation_id, created_at);

CREATE TABLE reviews (
                         id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
                         lesson_id uuid UNIQUE REFERENCES lessons(id),
                         student_id uuid NOT NULL REFERENCES student_profiles(user_id),
                         tutor_id uuid NOT NULL REFERENCES tutor_profiles(user_id),
                         rating int NOT NULL CHECK (rating BETWEEN 1 AND 5),
                         comment text,
                         created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE reports (
                         id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
                         reporter_id uuid NOT NULL REFERENCES users(id),
                         target_type text NOT NULL CHECK (target_type IN ('user','lesson','message')),
                         target_id uuid NOT NULL,
                         reason text NOT NULL,
                         status text NOT NULL DEFAULT 'open' CHECK (status IN ('open','in_review','resolved','rejected')),
                         created_at timestamptz NOT NULL DEFAULT now(),
                         updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE cancellation_policies (
                                       id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
                                       name text NOT NULL,
                                       rules jsonb NOT NULL
);

CREATE TABLE cancellations (
                               id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
                               lesson_id uuid UNIQUE REFERENCES lessons(id),
                               requested_by uuid NOT NULL REFERENCES users(id),
                               reason text,
                               policy_id uuid REFERENCES cancellation_policies(id),
                               outcome jsonb,
                               created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE refunds (
                         id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
                         payment_id uuid NOT NULL REFERENCES payments(id),
                         amount_minor bigint NOT NULL,
                         currency char(3) NOT NULL,
                         status text NOT NULL CHECK (status IN ('pending','succeeded','failed')),
                         provider_payload jsonb NOT NULL DEFAULT '{}'::jsonb,
                         created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE notifications (
                               id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
                               user_id uuid NOT NULL REFERENCES users(id),
                               channel text NOT NULL CHECK (channel IN ('email','sms','push','inapp')),
                               template_key text NOT NULL,
                               payload jsonb NOT NULL,
                               status text NOT NULL DEFAULT 'queued' CHECK (status IN ('queued','sent','failed')),
                               created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE feature_flags (
                               key text PRIMARY KEY,
                               enabled boolean NOT NULL DEFAULT false,
                               rules jsonb NOT NULL DEFAULT '{}'::jsonb,
                               updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE outbox_events (
                               id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
                               aggregate text NOT NULL,
                               aggregate_id uuid NOT NULL,
                               event_type text NOT NULL,
                               payload jsonb NOT NULL,
                               status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','sent','failed')),
                               created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS outbox_status_created_idx ON outbox_events (status, created_at);

CREATE TABLE audit_log (
                           id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
                           actor_id uuid REFERENCES users(id),
                           entity_type text NOT NULL,
                           entity_id uuid NOT NULL,
                           action text NOT NULL,
                           before jsonb,
                           after jsonb,
                           created_at timestamptz NOT NULL DEFAULT now()
);