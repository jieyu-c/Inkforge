package namespace

import (
	"context"
	"database/sql"
	"time"

	"github.com/jieyuc/inkforge/services/console/internal/pkg/accountscope"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/nsaudit"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/nsguard"
	"github.com/jieyuc/inkforge/services/console/internal/svc"
	"github.com/jieyuc/inkforge/services/console/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ArchiveNamespaceLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewArchiveNamespaceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ArchiveNamespaceLogic {
	return &ArchiveNamespaceLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ArchiveNamespaceLogic) ArchiveNamespace(nsSlug string) (*types.NamespaceDetail, error) {
	tid, err := accountscope.TenantID(l.ctx)
	if err != nil {
		return nil, err
	}
	row, err := findTenantNsOr404(l.ctx, l.svcCtx, tid, nsSlug)
	if err != nil {
		return nil, err
	}
	if row.Status == nsguard.StatusArchived {
		return detailFromRow(l.svcCtx, row)
	}
	now := time.Now().UTC()
	row.Status = nsguard.StatusArchived
	row.ArchivedAt = sql.NullTime{Time: now, Valid: true}
	if err := l.svcCtx.ConsoleNamespacesModel.Update(l.ctx, row); err != nil {
		l.Errorf("namespace archive: %v", err)
		return nil, err
	}
	nsaudit.Archived(tid, nsSlug)
	rec, ferr := findTenantNsOr404(l.ctx, l.svcCtx, tid, nsSlug)
	if ferr != nil {
		return nil, ferr
	}
	return detailFromRow(l.svcCtx, rec)
}
