package prompt

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/jieyuc/inkforge/services/console/internal/model"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/apperr"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/nsaudit"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/nsguard"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/nsquota"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/nstags"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/promptkey"
	"github.com/jieyuc/inkforge/services/console/internal/svc"
	"github.com/jieyuc/inkforge/services/console/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type CreatePromptLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreatePromptLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreatePromptLogic {
	return &CreatePromptLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreatePromptLogic) CreatePrompt(req *types.CreatePromptReq) (*types.PromptDetail, error) {
	ns, err := resolveTenantNs(l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}
	if err := nsguard.RejectArchivedWrite(ns, "creating a prompt is not allowed"); err != nil {
		return nil, err
	}
	if err := nsquota.PromptCreateRejected(
		l.svcCtx.Config.Namespace.DefaultQuotaPromptsPerNs,
		l.svcCtx.Config.Namespace.PlatformQuotaPromptsCap,
		ns,
	); err != nil {
		nsaudit.PromptQuotaExceeded(ns.TenantId, ns.NsSlug)
		return nil, err
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, apperr.BadRequest("INVALID_PAYLOAD", "title is required")
	}
	key, err := promptkey.AllocateUnique(title, func(candidate string) bool {
		_, ferr := l.svcCtx.ConsolePromptsModel.FindOneByNsIdPromptKey(l.ctx, ns.Id, candidate)
		return ferr == nil
	})
	if err != nil {
		return nil, err
	}
	tagNull, err := nstags.ToNull(req.Tags)
	if err != nil {
		return nil, err
	}
	row := model.ConsolePrompts{
		Id:          uuid.NewString(),
		TenantId:    ns.TenantId,
		NsId:        ns.Id,
		PromptKey:   key,
		Title:       sql.NullString{},
		Tags:        tagNull,
		OwnerUserId: sql.NullString{},
		DraftBody:   "",
		DraftSchema: sql.NullString{},
	}
	row.Title = sql.NullString{String: title, Valid: true}
	if uid := strings.TrimSpace(req.OwnerUserId); uid != "" {
		row.OwnerUserId = sql.NullString{String: uid, Valid: true}
	}

	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		pm := model.NewConsolePromptsModel(sqlx.NewSqlConnFromSession(session))
		nm := model.NewConsoleNamespacesModel(sqlx.NewSqlConnFromSession(session))
		if _, ierr := pm.Insert(ctx, &row); ierr != nil {
			var mysqlErr *mysql.MySQLError
			if errors.As(ierr, &mysqlErr) && mysqlErr.Number == 1062 {
				return apperr.Conflict("PROMPT_KEY_TAKEN", "prompt_key already exists in this namespace")
			}
			return ierr
		}
		return nm.AddPromptCount(ctx, ns.TenantId, ns.Id, 1)
	})
	if err != nil {
		return nil, err
	}
	nsaudit.PromptCreated(ns.TenantId, ns.NsSlug, key)
	rec, err := l.svcCtx.ConsolePromptsModel.FindOneByNsIdPromptKey(l.ctx, ns.Id, key)
	if err != nil {
		return nil, err
	}
	return rowToDetail(rec)
}
