package namespace

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/apperr"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/nsaudit"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/nsguard"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/nstags"
	"github.com/jieyuc/inkforge/services/console/internal/svc"
	"github.com/jieyuc/inkforge/services/console/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PatchNamespaceLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPatchNamespaceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PatchNamespaceLogic {
	return &PatchNamespaceLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PatchNamespaceLogic) PatchNamespace(nsSlug string, req *types.PatchNamespaceReq) (*types.NamespaceDetail,
	error,
) {
	tid, err := scopeTenant(l.ctx)
	if err != nil {
		return nil, err
	}
	row, err := FindTenantNsOr404(l.ctx, l.svcCtx, tid, nsSlug)
	if err != nil {
		return nil, err
	}
	if nsguard.IsArchived(row.Status) {
		return nil, apperr.Conflict("NS_ARCHIVED_READ_ONLY",
			"This namespace is archived; update settings via restore first")
	}

	changed := false
	if s := strings.TrimSpace(req.DisplayName); s != "" && s != row.DisplayName {
		row.DisplayName = s
		changed = true
	}

	if req.Tags != nil {
		tagNull, terr := nstags.ToNull(req.Tags)
		if terr != nil {
			return nil, terr
		}
		row.Tags = tagNull
		changed = true
	}

	if s := strings.TrimSpace(req.DefaultChannelSlug); s != "" {
		if !row.DefaultChannelSlug.Valid || row.DefaultChannelSlug.String != s {
			row.DefaultChannelSlug = sql.NullString{String: s, Valid: true}
			changed = true
		}
	}

	// MVP: description update only when explicitly present and non-whitespace JSON string.
	desc := strings.TrimSpace(req.Description)
	if req.Description != "" && desc != "" {
		if !row.Description.Valid || row.Description.String != desc {
			row.Description = sql.NullString{String: desc, Valid: true}
			changed = true
		}
	}

	if req.QuotaPromptsMax > 0 {
		if !row.QuotaPromptsMax.Valid || row.QuotaPromptsMax.Int64 != req.QuotaPromptsMax {
			row.QuotaPromptsMax = sql.NullInt64{Valid: true, Int64: req.QuotaPromptsMax}
			changed = true
		}
	}

	if !changed {
		return detailFromRow(l.svcCtx, row)
	}

	if err := l.svcCtx.ConsoleNamespacesModel.Update(l.ctx, row); err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return nil, apperr.Conflict("NS_SLUG_TAKEN", "Namespace slug collision")
		}
		l.Errorf("namespace patch: %v", err)
		return nil, err
	}

	rec, ferr := FindTenantNsOr404(l.ctx, l.svcCtx, tid, nsSlug)
	if ferr != nil {
		return nil, ferr
	}
	nsaudit.SettingsUpdated(tid, nsSlug)
	return detailFromRow(l.svcCtx, rec)
}
