DROP TABLE otp_codes;

DROP TABLE auth_sessions;

ALTER TABLE users DROP COLUMN phone_verified_at;

ALTER TABLE users DROP COLUMN email_verified_at;

ALTER TABLE users DROP COLUMN last_name;

ALTER TABLE users DROP COLUMN first_name;