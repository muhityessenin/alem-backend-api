-- Подключитесь к вашей базе данных и выполните этот запрос
CREATE EXTENSION IF NOT EXISTS "uuid-ossp"; -- Может понадобиться для генерации UUID на уровне БД

CREATE TABLE IF NOT EXISTS users (
                                     id UUID PRIMARY KEY,
                                     email TEXT NOT NULL UNIQUE,
                                     password_hash TEXT NOT NULL,
                                     name TEXT NOT NULL,
                                     role TEXT NOT NULL, -- 'student' или 'tutor'
                                     avatar TEXT,
                                     age INT,
                                     learning_goals TEXT[], -- Массив строк
                                     description TEXT,
                                     notifications_lessons BOOLEAN DEFAULT TRUE,
                                     notifications_messages BOOLEAN DEFAULT TRUE,
                                     notifications_reminders BOOLEAN DEFAULT TRUE,
                                     is_verified BOOLEAN DEFAULT FALSE,
                                     created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ
    );