// Code scaffolded by goctl. Safe to edit.

package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/apperr"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/iputil"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/jwtissue"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/refreshcookie"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/requestscope"
	"github.com/jieyuc/inkforge/services/console/internal/svc"
	"github.com/jieyuc/inkforge/services/console/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type RefreshLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRefreshLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshLogic {
	return &RefreshLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RefreshLogic) Refresh() (resp *types.RefreshResp, err error) {
	hdr, ok := requestscope.Peek(l.ctx)
	if !ok {
		return nil, apperr.BadRequest("INTERNAL_ERROR", "Request context unavailable")
	}
	w, r := hdr.W, hdr.R

	ip := httpx.GetRemoteAddr(r)

	name := l.svcCtx.Config.Auth.RefreshCookieName
	ck, err := r.Cookie(name)
	if err != nil || ck.Value == "" {
		return nil, apperr.Unauthorized("INVALID_REFRESH", "Missing refresh session")
	}

	raw, err := jwtissue.DecodeRefreshToken(ck.Value)
	if err != nil {
		return nil, apperr.Unauthorized("INVALID_REFRESH", "Missing refresh session")
	}
	rh := jwtissue.RefreshHash(raw)

	sess, err := l.svcCtx.Sessions.FindByRefreshHash(l.ctx, rh)
	if err != nil {
		l.Errorf("refresh lookup: %v", err)
		return nil, apperr.Unauthorized("INVALID_REFRESH", "Missing refresh session")
	}
	if sess == nil {
		return nil, apperr.Unauthorized("INVALID_REFRESH", "Missing refresh session")
	}

	if sess.RevokedAt.Valid {
		_ = l.svcCtx.Sessions.RevokeFamilyAll(l.ctx, sess.FamilyID)
		l.Infof("refresh reuse detected family=%s sid=%s", sess.FamilyID, sess.ID)
		return nil, apperr.Unauthorized("REFRESH_TOKEN_REUSE", "Session is no longer valid")
	}

	now := time.Now().UTC()
	if !sess.ExpiresAt.IsZero() && now.After(sess.ExpiresAt.UTC()) {
		return nil, apperr.Unauthorized("SESSION_EXPIRED", "Please sign in again")
	}

	newSID := uuid.NewString()
	rawNew, err := jwtissue.RandBytes(32)
	if err != nil {
		l.Errorf("rand refresh: %v", err)
		return nil, apperr.Unauthorized("INVALID_REFRESH", "Missing refresh session")
	}
	rhNew := jwtissue.RefreshHash(rawNew)

	ttl := time.Duration(l.svcCtx.Config.Auth.RefreshTTLSeconds) * time.Second
	if ttl <= 0 {
		ttl = 7 * 24 * time.Hour
	}
	newExp := now.Add(ttl)
	uaH := jwtissue.UAHash(r.UserAgent())
	ipB := iputil.PackIP(ip)

	const revokeQ = `UPDATE console_sessions SET revoked_at = UTC_TIMESTAMP(3), replaced_by = ?
WHERE id = ? AND revoked_at IS NULL AND expires_at > UTC_TIMESTAMP(3)`
	const insQ = `INSERT INTO console_sessions (id, user_id, family_id, refresh_hash, expires_at, ua_hash, last_ip)
VALUES (?, ?, ?, ?, ?, ?, ?)`

	err = l.svcCtx.DB.TransactCtx(l.ctx, func(c context.Context, se sqlx.Session) error {
		res, e := se.ExecCtx(c, revokeQ, newSID, sess.ID)
		if e != nil {
			return e
		}
		n, e := res.RowsAffected()
		if e != nil {
			return e
		}
		if n != 1 {
			return apperr.Unauthorized("INVALID_REFRESH", "Missing refresh session")
		}
		_, e = se.ExecCtx(c, insQ, newSID, sess.UserID, sess.FamilyID, rhNew, newExp.UTC(), uaH, ipB)
		return e
	})
	if err != nil {
		var he *apperr.HTTP
		if errors.As(err, &he) {
			return nil, err
		}
		l.Errorf("refresh rotate: %v", err)
		return nil, apperr.Unauthorized("INVALID_REFRESH", "Missing refresh session")
	}

	accessTTL := time.Duration(l.svcCtx.Config.JwtAuth.AccessExpire) * time.Second
	if accessTTL <= 0 {
		accessTTL = 900 * time.Second
	}
	secret := l.svcCtx.Config.JwtAuth.AccessSecret
	access, expiresIn, err := jwtissue.IssueAccess(secret, sess.UserID, newSID, accessTTL)
	if err != nil {
		l.Errorf("jwt refresh: %v", err)
		return nil, apperr.Unauthorized("INVALID_REFRESH", "Missing refresh session")
	}

	maxAge := int(l.svcCtx.Config.Auth.RefreshTTLSeconds)
	if maxAge <= 0 {
		maxAge = int(ttl.Seconds())
	}
	refreshcookie.Set(w, l.svcCtx.Config, jwtissue.EncodeRefreshToken(rawNew), maxAge)

	return &types.RefreshResp{
		AccessToken: access,
		ExpiresIn:   expiresIn,
	}, nil
}
