package prompt

import (
	"context"
	"strings"

	"github.com/jieyuc/inkforge/services/console/internal/svc"
	"github.com/jieyuc/inkforge/services/console/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListPromptsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListPromptsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPromptsLogic {
	return &ListPromptsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListPromptsLogic) ListPrompts(req *types.ListPromptsReq) (*types.PromptListResp, error) {
	ns, err := resolveTenantNs(l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}
	tid := ns.TenantId
	page, ps := pageClamp(req.Page, req.PageSize)
	q := strings.ToLower(strings.TrimSpace(req.Q))
	offset := int((page - 1) * ps)
	total, err := l.svcCtx.ConsolePromptsModel.CountByTenantNs(l.ctx, tid, ns.Id, q)
	if err != nil {
		return nil, err
	}
	rows, err := l.svcCtx.ConsolePromptsModel.ListByTenantNs(l.ctx, tid, ns.Id, q, offset, int(ps))
	if err != nil {
		return nil, err
	}
	items := make([]types.PromptSummary, 0, len(rows))
	for _, r := range rows {
		sum, err := rowToSummary(r)
		if err != nil {
			return nil, err
		}
		items = append(items, *sum)
	}
	return &types.PromptListResp{Items: items, Total: total}, nil
}
