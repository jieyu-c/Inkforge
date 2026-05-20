package namespace

import (
	"context"

	"github.com/jieyuc/inkforge/services/console/internal/svc"
	"github.com/jieyuc/inkforge/services/console/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetNamespaceLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetNamespaceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetNamespaceLogic {
	return &GetNamespaceLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetNamespaceLogic) GetNamespace(nsSlug string) (*types.NamespaceDetail, error) {
	tid, err := scopeTenant(l.ctx)
	if err != nil {
		return nil, err
	}
	row, err := FindTenantNsOr404(l.ctx, l.svcCtx, tid, nsSlug)
	if err != nil {
		return nil, err
	}
	return detailFromRow(l.svcCtx, row)
}
