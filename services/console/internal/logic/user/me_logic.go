// Code scaffolded by goctl. Safe to edit.

package user

import (
	"context"

	"github.com/jieyuc/inkforge/services/console/internal/pkg/apperr"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/authctx"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/phone"
	"github.com/jieyuc/inkforge/services/console/internal/svc"
	"github.com/jieyuc/inkforge/services/console/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type MeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MeLogic {
	return &MeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MeLogic) Me() (resp *types.MeResp, err error) {
	uid, ok := authctx.StringClaim(l.ctx, authctx.KeysUserID)
	if !ok || uid == "" {
		return nil, apperr.Unauthorized("UNAUTHORIZED", "Invalid or missing access token")
	}

	u, err := l.svcCtx.Users.FindByID(l.ctx, uid)
	if err != nil {
		l.Errorf("me lookup: %v", err)
		return nil, apperr.Unauthorized("UNAUTHORIZED", "Invalid or missing access token")
	}
	if u == nil {
		return nil, apperr.Unauthorized("UNAUTHORIZED", "Invalid or missing access token")
	}

	return &types.MeResp{
		UserID: u.ID,
		Phone:  phone.MaskDisplay(u.Phone),
	}, nil
}
