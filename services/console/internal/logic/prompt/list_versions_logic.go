package prompt

import (
	"context"

	"github.com/jieyuc/inkforge/services/console/internal/svc"
	"github.com/jieyuc/inkforge/services/console/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListVersionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListVersionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListVersionsLogic {
	return &ListVersionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListVersionsLogic) ListVersions(req *types.ListVersionsReq) (*types.PromptVersionListResp, error) {
	ns, err := resolveTenantNs(l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}
	key, err := pathPromptKey(l.ctx)
	if err != nil {
		return nil, err
	}
	pRow, err := loadPromptScoped(l.ctx, l.svcCtx, ns, key)
	if err != nil {
		return nil, err
	}
	page, ps := pageClamp(req.Page, req.PageSize)
	offset := int((page - 1) * ps)
	versionQ := req.Q
	total, err := l.svcCtx.ConsolePromptVersionsModel.CountByPromptScoped(l.ctx, ns.TenantId, ns.Id, pRow.Id, versionQ)
	if err != nil {
		return nil, err
	}
	rows, err := l.svcCtx.ConsolePromptVersionsModel.ListByPromptScoped(l.ctx, ns.TenantId, ns.Id, pRow.Id, versionQ, offset, int(ps))
	if err != nil {
		return nil, err
	}
	items := make([]types.PromptVersionItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, *versionToItem(r))
	}
	return &types.PromptVersionListResp{Items: items, Total: total}, nil
}
