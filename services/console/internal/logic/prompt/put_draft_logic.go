package prompt

import (
	"context"
	"strings"

	"github.com/jieyuc/inkforge/services/console/internal/pkg/nsaudit"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/nsguard"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/promptdraft"
	"github.com/jieyuc/inkforge/services/console/internal/svc"
	"github.com/jieyuc/inkforge/services/console/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PutDraftLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPutDraftLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PutDraftLogic {
	return &PutDraftLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PutDraftLogic) PutDraft(req *types.PutDraftReq) (*types.DraftResp, error) {
	ns, err := resolveTenantNs(l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}
	if err := nsguard.RejectArchivedWrite(ns, "saving a draft is not allowed"); err != nil {
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
	mode := strings.ToLower(strings.TrimSpace(l.svcCtx.Config.Prompt.DraftValidation))
	if mode == "" {
		mode = promptdraft.ModeStrict
	}
	if mode != promptdraft.ModeStrict && mode != promptdraft.ModeWarn {
		mode = promptdraft.ModeStrict
	}
	val, err := promptdraft.Validate(req.Body, req.Schema, mode)
	if err != nil {
		return nil, err
	}
	row.DraftBody = req.Body
	row.DraftSchema = draftSchemaNull(req.Schema)
	if err := l.svcCtx.ConsolePromptsModel.Update(l.ctx, row); err != nil {
		return nil, err
	}
	nsaudit.PromptDraftSaved(ns.TenantId, ns.NsSlug, key)
	rec, err := l.svcCtx.ConsolePromptsModel.FindOne(l.ctx, row.Id)
	if err != nil {
		return nil, err
	}
	var warnings []string
	if val != nil {
		warnings = val.Warnings
	}
	return rowToDraft(rec, warnings)
}
