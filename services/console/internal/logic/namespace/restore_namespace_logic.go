package namespace

import (
	"context"
	"database/sql"

	"github.com/jieyuc/inkforge/services/console/internal/pkg/accountscope"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/nsaudit"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/nsguard"
	"github.com/jieyuc/inkforge/services/console/internal/svc"
	"github.com/jieyuc/inkforge/services/console/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RestoreNamespaceLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRestoreNamespaceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RestoreNamespaceLogic {
	return &RestoreNamespaceLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RestoreNamespaceLogic) RestoreNamespace(nsSlug string) (*types.NamespaceDetail, error) {
	tid, err := accountscope.TenantID(l.ctx)
	if err != nil {
		return nil, err
	}
	row, err := FindTenantNsOr404(l.ctx, l.svcCtx, tid, nsSlug)
	if err != nil {
		return nil, err
	}
	if row.Status == nsguard.StatusActive {
		return detailFromRow(l.svcCtx, row)
	}
	row.Status = nsguard.StatusActive
	row.ArchivedAt = sql.NullTime{}
	if err := l.svcCtx.ConsoleNamespacesModel.Update(l.ctx, row); err != nil {
		l.Errorf("namespace restore: %v", err)
		return nil, err
	}
	nsaudit.Restored(tid, nsSlug)
	rec, ferr := FindTenantNsOr404(l.ctx, l.svcCtx, tid, nsSlug)
	if ferr != nil {
		return nil, ferr
	}
	return detailFromRow(l.svcCtx, rec)
}
