package prompt

import (
	"context"

	"github.com/jieyuc/inkforge/services/console/internal/model"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/nsaudit"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/nsguard"
	"github.com/jieyuc/inkforge/services/console/internal/svc"
	"github.com/jieyuc/inkforge/services/console/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type DeletePromptLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeletePromptLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeletePromptLogic {
	return &DeletePromptLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeletePromptLogic) DeletePrompt() (*types.Empty, error) {
	ns, err := resolveTenantNs(l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}
	if err := nsguard.RejectArchivedWrite(ns, "deleting a prompt is not allowed"); err != nil {
		return nil, err
	}
	key, err := pathPromptKey(l.ctx)
	if err != nil {
		return nil, err
	}
	row, err := loadPromptScoped(l.ctx, l.svcCtx, ns, key)
	if err != nil {
		return nil, err
	}
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		pm := model.NewConsolePromptsModel(sqlx.NewSqlConnFromSession(session))
		nm := model.NewConsoleNamespacesModel(sqlx.NewSqlConnFromSession(session))
		if err := pm.Delete(ctx, row.Id); err != nil {
			return err
		}
		return nm.AddPromptCount(ctx, ns.TenantId, ns.Id, -1)
	})
	if err != nil {
		return nil, err
	}
	nsaudit.PromptDeleted(ns.TenantId, ns.NsSlug, key)
	return &types.Empty{}, nil
}
