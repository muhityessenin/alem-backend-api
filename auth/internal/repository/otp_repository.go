package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OTPPurpose string

const (
	OTPPurposeLogin  OTPPurpose = "login"
	OTPPurposeVerify OTPPurpose = "verify"
)

type OTPRepository interface {
	Upsert(ctx context.Context, phone, codeHash string, purpose OTPPurpose, expiresAt time.Time) error
	GetActive(ctx context.Context, phone string, purpose OTPPurpose) (codeHash string, expiresAt time.Time, attempts int, err error)
	Consume(ctx context.Context, phone string, purpose OTPPurpose) error
	IncAttempt(ctx context.Context, phone string, purpose OTPPurpose) error
}

type otpRepo struct{ db *pgxpool.Pool }

func NewOTPRepo(db *pgxpool.Pool) OTPRepository { return &otpRepo{db: db} }

func (r *otpRepo) Upsert(ctx context.Context, phone, codeHash string, purpose OTPPurpose, expiresAt time.Time) error {
	_, err := r.db.Exec(ctx, `
INSERT INTO otp_codes (phone_e164, code_hash, purpose, expires_at, attempts_used, created_at)
VALUES ($1,$2,$3,$4,0,now())
ON CONFLICT (phone_e164, purpose)
DO UPDATE SET code_hash=EXCLUDED.code_hash, expires_at=EXCLUDED.expires_at, attempts_used=0, created_at=now(), consumed_at=NULL
`, phone, codeHash, string(purpose), expiresAt.UTC())
	return err
}

func (r *otpRepo) GetActive(ctx context.Context, phone string, purpose OTPPurpose) (string, time.Time, int, error) {
	var h string
	var exp time.Time
	var attempts int
	err := r.db.QueryRow(ctx, `
SELECT code_hash, expires_at, attempts_used
FROM otp_codes
WHERE phone_e164=$1 AND purpose=$2 AND consumed_at IS NULL AND expires_at > now()
`, phone, string(purpose)).Scan(&h, &exp, &attempts)
	return h, exp, attempts, err
}

func (r *otpRepo) Consume(ctx context.Context, phone string, purpose OTPPurpose) error {
	_, err := r.db.Exec(ctx, `
UPDATE otp_codes SET consumed_at = now()
WHERE phone_e164=$1 AND purpose=$2 AND consumed_at IS NULL
`, phone, string(purpose))
	return err
}

func (r *otpRepo) IncAttempt(ctx context.Context, phone string, purpose OTPPurpose) error {
	_, err := r.db.Exec(ctx, `
UPDATE otp_codes SET attempts_used = attempts_used + 1
WHERE phone_e164=$1 AND purpose=$2
`, phone, string(purpose))
	return err
}
