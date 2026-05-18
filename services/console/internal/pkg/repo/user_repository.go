package repo

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type User struct {
	ID                  string       `db:"id"`
	Phone               string       `db:"phone"`
	PasswordHash        string       `db:"password_hash"`
	FailedLoginAttempts int          `db:"failed_login_attempts"`
	LockedUntil         sql.NullTime `db:"locked_until"`
}

type Users struct {
	conn sqlx.SqlConn
}

func NewUsers(conn sqlx.SqlConn) *Users {
	return &Users{conn: conn}
}

func (u *Users) Insert(ctx context.Context, id, phone, passwordHash string) error {
	const q = `INSERT INTO console_users (id, phone, password_hash) VALUES (?, ?, ?)`
	_, err := u.conn.ExecCtx(ctx, q, id, phone, passwordHash)
	return err
}

func (u *Users) FindByPhone(ctx context.Context, phone string) (*User, error) {
	var row User
	const q = `SELECT id, phone, password_hash, failed_login_attempts, locked_until FROM console_users WHERE phone = ? LIMIT 1`
	switch err := u.conn.QueryRowCtx(ctx, &row, q, phone); {
	case errors.Is(err, sql.ErrNoRows):
		return nil, nil
	case err != nil:
		return nil, err
	default:
		return &row, nil
	}
}

func (u *Users) FindByID(ctx context.Context, id string) (*User, error) {
	var row User
	const q = `SELECT id, phone, password_hash, failed_login_attempts, locked_until FROM console_users WHERE id = ? LIMIT 1`
	switch err := u.conn.QueryRowCtx(ctx, &row, q, id); {
	case errors.Is(err, sql.ErrNoRows):
		return nil, nil
	case err != nil:
		return nil, err
	default:
		return &row, nil
	}
}

func (u *Users) ResetLoginFails(ctx context.Context, id string) error {
	const q = `UPDATE console_users SET failed_login_attempts = 0, locked_until = NULL WHERE id = ?`
	_, err := u.conn.ExecCtx(ctx, q, id)
	return err
}

func (u *Users) RecordLoginFail(ctx context.Context, id string, fails int, lockUntil *time.Time) error {
	if lockUntil == nil {
		const q = `UPDATE console_users SET failed_login_attempts = ? WHERE id = ?`
		_, err := u.conn.ExecCtx(ctx, q, fails, id)
		return err
	}
	const q = `UPDATE console_users SET failed_login_attempts = ?, locked_until = ? WHERE id = ?`
	_, err := u.conn.ExecCtx(ctx, q, fails, lockUntil.UTC(), id)
	return err
}
