package middleware

import (
	"net/http"

	"github.com/jieyuc/inkforge/services/console/internal/pkg/accountscope"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/authctx"
)

type TenantCtxMiddleware struct{}

func NewTenantCtxMiddleware() *TenantCtxMiddleware {
	return &TenantCtxMiddleware{}
}

func (m *TenantCtxMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if uid, ok := authctx.StringClaim(ctx, authctx.KeysUserID); ok && uid != "" {
			ctx = accountscope.WithTenant(ctx, uid)
		}
		next(w, r.WithContext(ctx))
	}
}
