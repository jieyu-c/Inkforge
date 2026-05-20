package prompt

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jieyuc/inkforge/services/console/internal/logic/namespace"
	"github.com/jieyuc/inkforge/services/console/internal/model"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/accountscope"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/apperr"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/consolepath"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/nstags"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/promptkey"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/promptversion"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/requestscope"
	"github.com/jieyuc/inkforge/services/console/internal/svc"
	"github.com/jieyuc/inkforge/services/console/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func httpReq(ctx context.Context) (*http.Request, error) {
	h, ok := requestscope.Peek(ctx)
	if !ok {
		return nil, apperr.Internal("REQUEST_SCOPE_MISSING", "request scope not available")
	}
	return h.R, nil
}

func pathNsSlug(ctx context.Context) (string, error) {
	r, err := httpReq(ctx)
	if err != nil {
		return "", err
	}
	return consolepath.TrimmedNsSlug(r)
}

func pathPromptKey(ctx context.Context) (string, error) {
	r, err := httpReq(ctx)
	if err != nil {
		return "", err
	}
	var p struct {
		PromptKey string `path:"promptKey"`
	}
	if err := httpx.ParsePath(r, &p); err != nil {
		return "", apperr.BadRequest("BAD_REQUEST", "Invalid path parameters")
	}
	return promptkey.MustValidate(p.PromptKey)
}

func pathChannel(ctx context.Context) (string, error) {
	r, err := httpReq(ctx)
	if err != nil {
		return "", err
	}
	var p struct {
		Channel string `path:"channel"`
	}
	if err := httpx.ParsePath(r, &p); err != nil {
		return "", apperr.BadRequest("BAD_REQUEST", "Invalid path parameters")
	}
	s := strings.ToLower(strings.TrimSpace(p.Channel))
	if s == "" || len(s) > 128 {
		return "", apperr.BadRequest("INVALID_CHANNEL", "channel slug is required (max 128 chars)")
	}
	return s, nil
}

func resolveVersionLabelPair(versionA, versionB string) (string, string, error) {
	a := promptversion.NormalizeLabel(versionA)
	b := promptversion.NormalizeLabel(versionB)
	if a == "" || b == "" {
		return "", "", apperr.BadRequest("INVALID_VERSION", "version_a and version_b are required")
	}
	return a, b, nil
}

func resolveTenantNs(ctx context.Context, svc *svc.ServiceContext) (*model.ConsoleNamespaces, error) {
	tid, err := accountscope.TenantID(ctx)
	if err != nil {
		return nil, err
	}
	nsSlug, err := pathNsSlug(ctx)
	if err != nil {
		return nil, err
	}
	return namespace.FindTenantNsOr404(ctx, svc, tid, nsSlug)
}

func loadPromptScoped(ctx context.Context, svc *svc.ServiceContext, ns *model.ConsoleNamespaces, key string,
) (*model.ConsolePrompts, error) {
	row, err := svc.ConsolePromptsModel.FindOneByNsIdPromptKey(ctx, ns.Id, key)
	switch {
	case errors.Is(err, model.ErrNotFound):
		return nil, apperr.NotFound("PROMPT_NOT_FOUND", "Prompt does not exist in this namespace")
	case err != nil:
		return nil, err
	case row.TenantId != ns.TenantId || row.NsId != ns.Id:
		return nil, apperr.Forbidden("PROMPT_SCOPE", "Prompt is not in this namespace")
	default:
		return row, nil
	}
}

func pageClamp(page, pageSize int64) (int64, int64) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func rowToSummary(row *model.ConsolePrompts) (*types.PromptSummary, error) {
	tags, err := nstags.FromNull(row.Tags)
	if err != nil {
		tags = nil
	}
	s := &types.PromptSummary{
		PromptKey: row.PromptKey,
		Tags:      tags,
		UpdatedAt: row.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if row.Title.Valid {
		s.Title = row.Title.String
	}
	if row.OwnerUserId.Valid {
		s.OwnerUserId = row.OwnerUserId.String
	}
	return s, nil
}

func rowToDetail(row *model.ConsolePrompts) (*types.PromptDetail, error) {
	tags, err := nstags.FromNull(row.Tags)
	if err != nil {
		tags = nil
	}
	d := &types.PromptDetail{
		PromptKey: row.PromptKey,
		Tags:      tags,
		DraftBody: row.DraftBody,
		UpdatedAt: row.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if row.Title.Valid {
		d.Title = row.Title.String
	}
	if row.OwnerUserId.Valid {
		d.OwnerUserId = row.OwnerUserId.String
	}
	if row.DraftSchema.Valid && strings.TrimSpace(row.DraftSchema.String) != "" {
		d.DraftSchema = row.DraftSchema.String
	}
	return d, nil
}

func rowToDraft(row *model.ConsolePrompts, warnings []string) (*types.DraftResp, error) {
	d := &types.DraftResp{
		Body:      row.DraftBody,
		UpdatedAt: row.UpdatedAt.UTC().Format(time.RFC3339),
		Warnings:  warnings,
	}
	if row.DraftSchema.Valid && strings.TrimSpace(row.DraftSchema.String) != "" {
		d.Schema = row.DraftSchema.String
	}
	return d, nil
}

func draftSchemaNull(schema string) sql.NullString {
	s := strings.TrimSpace(schema)
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
