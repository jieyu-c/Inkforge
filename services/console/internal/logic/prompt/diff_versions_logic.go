package prompt

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/jieyuc/inkforge/services/console/internal/model"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/apperr"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/promptdiff"
	"github.com/jieyuc/inkforge/services/console/internal/svc"
	"github.com/jieyuc/inkforge/services/console/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DiffVersionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDiffVersionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DiffVersionsLogic {
	return &DiffVersionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DiffVersionsLogic) DiffVersions(req *types.DiffVersionsReq) (*types.VersionDiffResp, error) {
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
	labelA, labelB, err := resolveVersionLabelPair(req.VersionA, req.VersionB)
	if err != nil {
		return nil, err
	}
	a, err := loadVersionByLabel(l.ctx, l.svcCtx, ns, pRow, labelA)
	if err != nil {
		return nil, err
	}
	b, err := loadVersionByLabel(l.ctx, l.svcCtx, ns, pRow, labelB)
	if err != nil {
		return nil, err
	}
	if a.TenantId != ns.TenantId || b.TenantId != ns.TenantId {
		return nil, apperr.Forbidden("VERSION_SCOPE", "Version scope mismatch")
	}
	tagA := versionDiffTag(a)
	tagB := versionDiffTag(b)
	sa := schemaStr(a.SchemaJson)
	sb := schemaStr(b.SchemaJson)
	canonA, _ := canonicalJSON(sa)
	canonB, _ := canonicalJSON(sb)
	schemaDiff := ""
	if canonA != canonB {
		schemaDiff = promptdiff.Unified("schema "+tagA, "schema "+tagB, sa, sb)
	}
	bodyDiff := promptdiff.Unified("body "+tagA, "body "+tagB, a.Body, b.Body)
	return &types.VersionDiffResp{
		BodyDiff:   strings.TrimSpace(bodyDiff),
		SchemaDiff: strings.TrimSpace(schemaDiff),
	}, nil
}

func loadVersionByLabel(
	ctx context.Context,
	svc *svc.ServiceContext,
	ns *model.ConsoleNamespaces,
	pRow *model.ConsolePrompts,
	label string,
) (*model.ConsolePromptVersions, error) {
	row, err := svc.ConsolePromptVersionsModel.FindOneByPromptVersionLabel(ctx, ns.TenantId, ns.Id, pRow.Id, label)
	if err == nil {
		return row, nil
	}
	if !errors.Is(err, model.ErrNotFound) {
		return nil, err
	}
	if n, parseErr := strconv.ParseUint(label, 10, 64); parseErr == nil {
		row, err = svc.ConsolePromptVersionsModel.FindOneByPromptIdVersionNum(ctx, pRow.Id, n)
		if errors.Is(err, model.ErrNotFound) {
			return nil, apperr.NotFound("VERSION_NOT_FOUND", "Version not found for this prompt")
		}
		if err != nil {
			return nil, err
		}
		return row, nil
	}
	return nil, apperr.NotFound("VERSION_NOT_FOUND", "Version not found for this prompt")
}

func versionDiffTag(v *model.ConsolePromptVersions) string {
	if lbl := strings.TrimSpace(v.VersionLabel); lbl != "" {
		return "v" + lbl
	}
	return "v" + strconv.FormatUint(v.VersionNum, 10)
}

func schemaStr(ns sql.NullString) string {
	if !ns.Valid {
		return ""
	}
	return strings.TrimSpace(ns.String)
}

func canonicalJSON(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	var v any
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return raw, err
	}
	b, err := json.Marshal(v)
	if err != nil {
		return raw, err
	}
	return string(b), nil
}
