// Code scaffolded by goctl. Safe to edit.

package auth

import (
	"context"

	"github.com/jieyuc/inkforge/services/console/internal/pkg/apperr"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/jwtissue"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/refreshcookie"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/requestscope"
	"github.com/jieyuc/inkforge/services/console/internal/svc"
	"github.com/jieyuc/inkforge/services/console/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type LogoutLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogoutLogic {
	return &LogoutLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LogoutLogic) Logout() (resp *types.Empty, err error) {
	hdr, ok := requestscope.Peek(l.ctx)
	if !ok {
		return nil, apperr.BadRequest("INTERNAL_ERROR", "Request context unavailable")
	}
	w, r := hdr.W, hdr.R

	name := l.svcCtx.Config.Auth.RefreshCookieName
	ck, err := r.Cookie(name)
	if err == nil && ck.Value != "" {
		if raw, derr := jwtissue.DecodeRefreshToken(ck.Value); derr == nil {
			rh := jwtissue.RefreshHash(raw)
			sess, qerr := l.svcCtx.Sessions.FindByRefreshHash(l.ctx, rh)
			if qerr != nil {
				l.Errorf("logout lookup: %v", qerr)
			} else if sess != nil && !sess.RevokedAt.Valid {
				if rerr := l.svcCtx.Sessions.RevokeByID(l.ctx, sess.ID); rerr != nil {
					l.Errorf("revoke session: %v", rerr)
				}
			}
		}
	}

	refreshcookie.Clear(w, l.svcCtx.Config)
	return &types.Empty{}, nil
}
