package namespace

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/jieyuc/inkforge/services/console/internal/model"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/accountscope"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/apperr"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/nsaudit"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/nsguard"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/nslug"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/nstags"
	"github.com/jieyuc/inkforge/services/console/internal/svc"
	"github.com/jieyuc/inkforge/services/console/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateNamespaceLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateNamespaceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateNamespaceLogic {
	return &CreateNamespaceLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateNamespaceLogic) CreateNamespace(req *types.CreateNamespaceReq) (resp *types.NamespaceDetail, err error) {
	tenantID, err := accountscope.TenantID(l.ctx)
	if err != nil {
		return nil, err
	}
	displayName := strings.TrimSpace(req.DisplayName)
	if displayName == "" {
		return nil, apperr.BadRequest("INVALID_PAYLOAD", "display_name is required")
	}
	slug, err := nslug.AllocateUnique(displayName, func(candidate string) bool {
		_, ferr := l.svcCtx.ConsoleNamespacesModel.FindOneByTenantIdNsSlug(l.ctx, tenantID, candidate)
		return ferr == nil
	})
	if err != nil {
		return nil, err
	}
	tagNull, err := nstags.ToNull(req.Tags)
	if err != nil {
		return nil, err
	}
	var quota sql.NullInt64
	switch {
	case req.QuotaPromptsMax > 0:
		quota = sql.NullInt64{Valid: true, Int64: req.QuotaPromptsMax}
	default:
		quota = sql.NullInt64{}
	}
	row := model.ConsoleNamespaces{
		Id:                 uuid.NewString(),
		TenantId:           tenantID,
		NsSlug:             slug,
		DisplayName:        displayName,
		Description:        sql.NullString{},
		Tags:               tagNull,
		Status:             nsguard.StatusActive,
		DefaultChannelSlug: sql.NullString{String: nslug.DefaultChannelSlug, Valid: true},
		ArchivedAt:         sql.NullTime{},
		QuotaPromptsMax:    quota,
		PromptCount:        0,
	}
	if req.Description != "" {
		row.Description = sql.NullString{String: strings.TrimSpace(req.Description), Valid: true}
	}
	if _, ierr := l.svcCtx.ConsoleNamespacesModel.Insert(l.ctx, &row); ierr != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(ierr, &mysqlErr) && mysqlErr.Number == 1062 {
			return nil, apperr.Conflict("NS_SLUG_TAKEN", "Namespace slug already exists for this account")
		}
		l.Errorf("namespace insert: %v", ierr)
		return nil, ierr
	}
	nsaudit.Created(tenantID, slug)
	rec, ferr := l.svcCtx.ConsoleNamespacesModel.FindOneByTenantIdNsSlug(l.ctx, tenantID, slug)
	if ferr != nil {
		l.Errorf("namespace reload after create: %v", ferr)
		return nil, ferr
	}
	return detailFromRow(l.svcCtx, rec)
}
