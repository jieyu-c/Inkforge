package prompt

import (
	"context"
	"database/sql"
	"strings"

	"github.com/jieyuc/inkforge/services/console/internal/pkg/nsguard"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/nstags"
	"github.com/jieyuc/inkforge/services/console/internal/svc"
	"github.com/jieyuc/inkforge/services/console/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PatchPromptLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPatchPromptLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PatchPromptLogic {
	return &PatchPromptLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PatchPromptLogic) PatchPrompt(req *types.PatchPromptReq) (*types.PromptDetail, error) {
	ns, err := resolveTenantNs(l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}
	if err := nsguard.RejectArchivedWrite(ns, "updating prompt metadata is not allowed"); err != nil {
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
	changed := false
	if s := strings.TrimSpace(req.Title); s != "" {
		if !row.Title.Valid || row.Title.String != s {
			row.Title = sql.NullString{String: s, Valid: true}
			changed = true
		}
	}
	if req.Tags != nil {
		tagNull, err := nstags.ToNull(req.Tags)
		if err != nil {
			return nil, err
		}
		row.Tags = tagNull
		changed = true
	}
	if s := strings.TrimSpace(req.OwnerUserId); s != "" {
		if !row.OwnerUserId.Valid || row.OwnerUserId.String != s {
			row.OwnerUserId = sql.NullString{String: s, Valid: true}
			changed = true
		}
	}
	if changed {
		if err := l.svcCtx.ConsolePromptsModel.Update(l.ctx, row); err != nil {
			return nil, err
		}
	}
	rec, err := l.svcCtx.ConsolePromptsModel.FindOne(l.ctx, row.Id)
	if err != nil {
		return nil, err
	}
	return rowToDetail(rec)
}
