package requestscope

import (
	"context"
	"net/http"
)

type ctxKey struct{}

// HTTP binds the writer and raw request onto the incoming context chain (middleware only).
type HTTP struct {
	W http.ResponseWriter
	R *http.Request
}

// With stores w/r for downstream logic/handlers built by goctl.
func With(parent context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	return context.WithValue(parent, ctxKey{}, &HTTP{W: w, R: r})
}

// Peek returns injected HTTP values; absent when middleware is not installed.
func Peek(ctx context.Context) (*HTTP, bool) {
	v, ok := ctx.Value(ctxKey{}).(*HTTP)
	return v, ok && v != nil && v.W != nil && v.R != nil
}
