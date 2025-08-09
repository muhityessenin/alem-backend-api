package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionsRepository interface {
	Create(ctx context.Context, id string, userID string, tokenHash string, expiresAt time.Time, ua, ip string) error
	GetByHash(ctx context.Context, tokenHash string) (id string, userID string, expiresAt time.Time, revokedAt *time.Time, err error)
	Revoke(ctx context.Context, id string) error
	RevokeAllByUser(ctx context.Context, userID string) error
}

type sessionsRepo struct{ db *pgxpool.Pool }

func NewSessionsRepo(db *pgxpool.Pool) SessionsRepository { return &sessionsRepo{db: db} }

func (r *sessionsRepo) Create(ctx context.Context, id, userID, tokenHash string, expiresAt time.Time, ua, ip string) error {
	_, err := r.db.Exec(ctx, `
INSERT INTO auth_sessions (id, user_id, refresh_token_hash, expires_at, user_agent, ip, created_at)
VALUES ($1,$2,$3,$4,$5,$6, now())`,
		id, userID, tokenHash, expiresAt.UTC(), ua, ip)
	return err
}

func (r *sessionsRepo) GetByHash(ctx context.Context, tokenHash string) (string, string, time.Time, *time.Time, error) {
	var id, userID string
	var exp time.Time
	var revokedAt *time.Time
	err := r.db.QueryRow(ctx, `
SELECT id, user_id, expires_at, revoked_at
FROM auth_sessions
WHERE refresh_token_hash = $1
`, tokenHash).Scan(&id, &userID, &exp, &revokedAt)
	return id, userID, exp, revokedAt, err
}

func (r *sessionsRepo) Revoke(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `UPDATE auth_sessions SET revoked_at = now() WHERE id = $1 AND revoked_at IS NULL`, id)
	return err
}

func (r *sessionsRepo) RevokeAllByUser(ctx context.Context, userID string) error {
	_, err := r.db.Exec(ctx, `UPDATE auth_sessions SET revoked_at = now() WHERE user_id = $1 AND revoked_at IS NULL`, userID)
	return err
}
