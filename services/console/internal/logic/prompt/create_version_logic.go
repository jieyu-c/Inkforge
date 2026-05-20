package prompt

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/jieyuc/inkforge/services/console/internal/model"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/apperr"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/authctx"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/nsaudit"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/nsguard"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/promptversion"
	"github.com/jieyuc/inkforge/services/console/internal/svc"
	"github.com/jieyuc/inkforge/services/console/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateVersionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateVersionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateVersionLogic {
	return &CreateVersionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateVersionLogic) CreateVersion(req *types.CreateVersionReq) (*types.PromptVersionItem, error) {
	ns, err := resolveTenantNs(l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}
	if err := nsguard.RejectArchivedWrite(ns, "creating a version snapshot is not allowed"); err != nil {
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
	maxCap := l.svcCtx.Config.Prompt.MaxVersionsPerPrompt
	if maxCap > 0 {
		n, err := l.svcCtx.ConsolePromptVersionsModel.CountByPromptScoped(l.ctx, ns.TenantId, ns.Id, pRow.Id, "")
		if err != nil {
			return nil, err
		}
		if n >= maxCap {
			return nil, apperr.Conflict("PROMPT_VERSIONS_QUOTA_EXCEEDED", "Too many snapshots for this prompt")
		}
	}
	uid, ok := authctx.StringClaim(l.ctx, authctx.KeysUserID)
	if !ok || uid == "" {
		return nil, apperr.Unauthorized("SESSION_INVALID", "User id missing from session")
	}
	existing, err := l.svcCtx.ConsolePromptVersionsModel.ListVersionLabelsByPrompt(l.ctx, ns.TenantId, ns.Id, pRow.Id)
	if err != nil {
		return nil, err
	}
	versionLabel, err := promptversion.ResolveLabel(req.Version, existing)
	if err != nil {
		return nil, err
	}
	maxNum, err := l.svcCtx.ConsolePromptVersionsModel.MaxVersionNum(l.ctx, ns.TenantId, ns.Id, pRow.Id)
	if err != nil {
		return nil, err
	}
	versionNum := maxNum + 1
	v := model.ConsolePromptVersions{
		Id:              uuid.NewString(),
		TenantId:        ns.TenantId,
		NsId:            ns.Id,
		PromptId:        pRow.Id,
		VersionNum:      versionNum,
		VersionLabel:    versionLabel,
		Body:            pRow.DraftBody,
		SchemaJson:      pRow.DraftSchema,
		ChangeNote:      sql.NullString{},
		CreatedByUserId: uid,
	}
	if note := strings.TrimSpace(req.ChangeNote); note != "" {
		v.ChangeNote = sql.NullString{String: note, Valid: true}
	}
	if _, err := l.svcCtx.ConsolePromptVersionsModel.Insert(l.ctx, &v); err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return nil, apperr.Conflict("VERSION_LABEL_TAKEN", "version already exists for this prompt")
		}
		return nil, err
	}
	nsaudit.PromptVersionCreated(ns.TenantId, ns.NsSlug, key, versionLabel)
	return versionToItem(&v), nil
}

func versionToItem(v *model.ConsolePromptVersions) *types.PromptVersionItem {
	item := &types.PromptVersionItem{
		Id:              v.Id,
		Version:         v.VersionLabel,
		CreatedByUserId: v.CreatedByUserId,
		CreatedAt:       v.CreatedAt.UTC().Format(time.RFC3339),
	}
	if v.ChangeNote.Valid {
		item.ChangeNote = v.ChangeNote.String
	}
	return item
}
