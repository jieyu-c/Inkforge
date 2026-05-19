package accountscope

import (
	"context"

	"github.com/jieyuc/inkforge/services/console/internal/pkg/apperr"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/authctx"
)

type tenantInternalKey struct{}

// WithTenant records the authenticated account isolation domain (tenant_id). Personal MVP mirrors user id.
func WithTenant(parent context.Context, tenantID string) context.Context {
	return context.WithValue(parent, tenantInternalKey{}, tenantID)
}

func tenantFromTyped(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(tenantInternalKey{}).(string)
	return v, ok && v != ""
}

// TenantID returns the authenticated tenant/account isolation identifier.
func TenantID(ctx context.Context) (string, error) {
	if tid, ok := tenantFromTyped(ctx); ok && tid != "" {
		return tid, nil
	}
	return tenantFromJWTClaims(ctx)
}

func tenantFromJWTClaims(ctx context.Context) (string, error) {
	uid, ok := authctx.StringClaim(ctx, authctx.KeysUserID)
	if !ok || uid == "" {
		return "", apperr.Unauthorized("UNAUTHORIZED", "Invalid or missing access token")
	}
	return uid, nil
}
