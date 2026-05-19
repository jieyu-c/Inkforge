package namespace

import (
	"context"
	"errors"

	"github.com/jieyuc/inkforge/services/console/internal/model"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/accountscope"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/apperr"
	"github.com/jieyuc/inkforge/services/console/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func findTenantNsOr404(ctx context.Context, svcCtx *svc.ServiceContext, tenantID, slug string,
) (*model.ConsoleNamespaces, error) {
	row, err := svcCtx.ConsoleNamespacesModel.FindOneByTenantIdNsSlug(ctx, tenantID, slug)
	switch {
	case errors.Is(err, sqlx.ErrNotFound):
		return nil, apperr.NotFound("NS_NOT_FOUND", "Namespace does not exist")
	case err != nil:
		logx.WithContext(ctx).Errorf("namespace lookup: %v", err)
		return nil, err
	default:
		return row, nil
	}
}

func scopeTenant(ctx context.Context) (string, error) {
	return accountscope.TenantID(ctx)
}
