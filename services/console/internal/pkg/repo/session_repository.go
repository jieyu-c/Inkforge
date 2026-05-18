package repo

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type Session struct {
	ID         string         `db:"id"`
	UserID     string         `db:"user_id"`
	FamilyID   string         `db:"family_id"`
	RevokedAt  sql.NullTime   `db:"revoked_at"`
	ReplacedBy sql.NullString `db:"replaced_by"`
	ExpiresAt  time.Time      `db:"expires_at"`
}

type Sessions struct {
	conn sqlx.SqlConn
}

func NewSessions(conn sqlx.SqlConn) *Sessions {
	return &Sessions{conn: conn}
}

func (s *Sessions) Create(ctx context.Context, id, userID, familyID string, refreshHash []byte,
	expiresAt time.Time,
) error {
	const q = `INSERT INTO console_sessions (id, user_id, family_id, refresh_hash, expires_at)
VALUES (?, ?, ?, ?, ?)`
	_, err := s.conn.ExecCtx(ctx, q, id, userID, familyID, refreshHash, expiresAt.UTC())
	return err
}

func (s *Sessions) CreateWithUA(ctx context.Context, id, userID, familyID string, refreshHash []byte,
	expiresAt time.Time, uaHash []byte, lastIP []byte,
) error {
	const q = `INSERT INTO console_sessions (id, user_id, family_id, refresh_hash, expires_at, ua_hash, last_ip)
VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := s.conn.ExecCtx(ctx, q, id, userID, familyID, refreshHash, expiresAt.UTC(), uaHash, lastIP)
	return err
}

func (s *Sessions) FindByRefreshHash(ctx context.Context, hash []byte) (*Session, error) {
	var row Session
	const q = `SELECT id, user_id, family_id, revoked_at, replaced_by, expires_at
FROM console_sessions WHERE refresh_hash = ? LIMIT 1`
	switch err := s.conn.QueryRowCtx(ctx, &row, q, hash); {
	case errors.Is(err, sql.ErrNoRows):
		return nil, nil
	case err != nil:
		return nil, err
	default:
		return &row, nil
	}
}

func (s *Sessions) RevokeByIDReplace(ctx context.Context, id string, replacedBy string) error {
	const q = `UPDATE console_sessions SET revoked_at = UTC_TIMESTAMP(3), replaced_by = ?
WHERE id = ? AND revoked_at IS NULL`
	_, err := s.conn.ExecCtx(ctx, q, replacedBy, id)
	return err
}

func (s *Sessions) RevokeByID(ctx context.Context, id string) error {
	const q = `UPDATE console_sessions SET revoked_at = UTC_TIMESTAMP(3)
WHERE id = ? AND revoked_at IS NULL`
	_, err := s.conn.ExecCtx(ctx, q, id)
	return err
}

func (s *Sessions) RevokeFamilyExcept(ctx context.Context, familyID string, exceptID string) error {
	const q = `UPDATE console_sessions SET revoked_at = UTC_TIMESTAMP(3)
WHERE family_id = ? AND id <> ? AND revoked_at IS NULL`
	_, err := s.conn.ExecCtx(ctx, q, familyID, exceptID)
	return err
}

func (s *Sessions) RevokeFamilyAll(ctx context.Context, familyID string) error {
	const q = `UPDATE console_sessions SET revoked_at = UTC_TIMESTAMP(3)
WHERE family_id = ? AND revoked_at IS NULL`
	_, err := s.conn.ExecCtx(ctx, q, familyID)
	return err
}
