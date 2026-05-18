// Code scaffolded by goctl. Safe to edit.

package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/apperr"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/iputil"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/jwtissue"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/password"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/phone"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/refreshcookie"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/requestscope"
	"github.com/jieyuc/inkforge/services/console/internal/svc"
	"github.com/jieyuc/inkforge/services/console/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func invalidLoginCreds() error {
	return apperr.Unauthorized("INVALID_CREDENTIALS", "Incorrect phone or password")
}

func (l *LoginLogic) Login(req *types.LoginReq) (resp *types.LoginResp, err error) {
	hdr, ok := requestscope.Peek(l.ctx)
	if !ok {
		return nil, apperr.BadRequest("INTERNAL_ERROR", "Request context unavailable")
	}
	w, r := hdr.W, hdr.R

	ip := httpx.GetRemoteAddr(r)

	phoneCanon := phone.Canonical(req.Phone)
	if phoneCanon == "" {
		return nil, invalidLoginCreds()
	}

	u, err := l.svcCtx.Users.FindByPhone(l.ctx, phoneCanon)
	if err != nil {
		l.Errorf("lookup user: %v", err)
		return nil, invalidLoginCreds()
	}
	if u == nil {
		return nil, invalidLoginCreds()
	}

	now := time.Now().UTC()
	if u.LockedUntil.Valid && u.LockedUntil.Time.After(now) {
		return nil, invalidLoginCreds()
	}

	if err := password.Verify(u.PasswordHash, req.Password); err != nil {
		fails := u.FailedLoginAttempts + 1
		maxFails := l.svcCtx.Config.Auth.LockFailures
		if maxFails <= 0 {
			maxFails = 8
		}
		lockMin := l.svcCtx.Config.Auth.LockDurationMinutes
		if lockMin <= 0 {
			lockMin = 15
		}
		var lock *time.Time
		if fails >= maxFails {
			t := now.Add(time.Duration(lockMin) * time.Minute)
			lock = &t
			fails = 0
		}
		if uerr := l.svcCtx.Users.RecordLoginFail(l.ctx, u.ID, fails, lock); uerr != nil {
			l.Errorf("record login fail: %v", uerr)
		}
		return nil, invalidLoginCreds()
	}

	if err := l.svcCtx.Users.ResetLoginFails(l.ctx, u.ID); err != nil {
		l.Errorf("reset fails: %v", err)
		return nil, invalidLoginCreds()
	}

	sid := uuid.NewString()
	family := uuid.NewString()
	raw, err := jwtissue.RandBytes(32)
	if err != nil {
		l.Errorf("rand: %v", err)
		return nil, invalidLoginCreds()
	}
	rh := jwtissue.RefreshHash(raw)
	ttl := time.Duration(l.svcCtx.Config.Auth.RefreshTTLSeconds) * time.Second
	exp := now.Add(ttl)
	uaH := jwtissue.UAHash(r.UserAgent())
	ipB := iputil.PackIP(ip)
	if err := l.svcCtx.Sessions.CreateWithUA(l.ctx, sid, u.ID, family, rh, exp, uaH, ipB); err != nil {
		l.Errorf("session: %v", err)
		return nil, invalidLoginCreds()
	}

	accessTTL := time.Duration(l.svcCtx.Config.JwtAuth.AccessExpire) * time.Second
	if accessTTL <= 0 {
		accessTTL = 900 * time.Second
	}
	secret := l.svcCtx.Config.JwtAuth.AccessSecret
	access, expiresIn, err := jwtissue.IssueAccess(secret, u.ID, sid, accessTTL)
	if err != nil {
		l.Errorf("jwt: %v", err)
		return nil, invalidLoginCreds()
	}

	maxAge := int(l.svcCtx.Config.Auth.RefreshTTLSeconds)
	if maxAge <= 0 {
		maxAge = int(ttl.Seconds())
	}
	refreshcookie.Set(w, l.svcCtx.Config, jwtissue.EncodeRefreshToken(raw), maxAge)

	return &types.LoginResp{
		AccessToken: access,
		ExpiresIn:   expiresIn,
	}, nil
}
