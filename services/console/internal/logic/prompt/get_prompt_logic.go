package prompt

import (
	"context"

	"github.com/jieyuc/inkforge/services/console/internal/svc"
	"github.com/jieyuc/inkforge/services/console/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPromptLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPromptLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPromptLogic {
	return &GetPromptLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPromptLogic) GetPrompt() (*types.PromptDetail, error) {
	ns, err := resolveTenantNs(l.ctx, l.svcCtx)
	if err != nil {
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
	return rowToDetail(row)
}
