// Code scaffolded by goctl. Safe to edit.

package auth

import (
	"context"
	"errors"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/apperr"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/password"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/phone"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/requestscope"
	"github.com/jieyuc/inkforge/services/console/internal/svc"
	"github.com/jieyuc/inkforge/services/console/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func bcryptCost(v int) int {
	if v < bcrypt.MinCost {
		return bcrypt.DefaultCost
	}
	if v > bcrypt.MaxCost {
		return bcrypt.MaxCost
	}
	return v
}

func (l *RegisterLogic) Register(req *types.RegisterReq) (resp *types.RegisterResp, err error) {
	if _, ok := requestscope.Peek(l.ctx); !ok {
		return nil, apperr.BadRequest("INTERNAL_ERROR", "Request context unavailable")
	}
	phoneCanon := phone.Canonical(req.Phone)
	if phoneCanon == "" {
		return nil, apperr.BadRequest("INVALID_PHONE", "Invalid phone number format")
	}

	minLen := l.svcCtx.Config.Auth.MinPasswordLength
	if minLen <= 0 {
		minLen = 8
	}
	if len(req.Password) < minLen {
		return nil, apperr.BadRequest("WEAK_PASSWORD", "Password does not meet policy")
	}

	ph, err := password.Hash(req.Password, bcryptCost(l.svcCtx.Config.Auth.BcryptCost))
	if err != nil {
		l.Errorf("hash password: %v", err)
		return nil, apperr.BadRequest("BAD_REQUEST", "Could not complete registration")
	}

	id := uuid.NewString()
	err = l.svcCtx.Users.Insert(l.ctx, id, phoneCanon, ph)
	if err != nil {
		var me *mysql.MySQLError
		if errors.As(err, &me) && me.Number == 1062 {
			return nil, apperr.Conflict("PHONE_ALREADY_REGISTERED", "Phone number already registered")
		}
		l.Errorf("insert user: %v", err)
		return nil, apperr.BadRequest("BAD_REQUEST", "Could not complete registration")
	}

	return &types.RegisterResp{Message: "ok"}, nil
}
