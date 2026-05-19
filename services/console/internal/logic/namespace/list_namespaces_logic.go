// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package namespace

import (
	"context"

	"github.com/jieyuc/inkforge/services/console/internal/svc"
	"github.com/jieyuc/inkforge/services/console/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListNamespacesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListNamespacesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListNamespacesLogic {
	return &ListNamespacesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListNamespacesLogic) ListNamespaces() (*types.NamespaceListResp, error) {
	tenantID, err := scopeTenant(l.ctx)
	if err != nil {
		return nil, err
	}
	rows, err := l.svcCtx.ConsoleNamespacesModel.ListByTenantId(l.ctx, tenantID)
	if err != nil {
		l.Errorf("namespace list: %v", err)
		return nil, err
	}
	items := make([]types.NamespaceDetail, 0, len(rows))
	for _, row := range rows {
		d, derr := detailFromRow(l.svcCtx, row)
		if derr != nil {
			l.Errorf("namespace map detail: %v", derr)
			return nil, derr
		}
		items = append(items, *d)
	}
	return &types.NamespaceListResp{Namespaces: items}, nil
}
