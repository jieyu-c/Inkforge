// Code scaffolded by goctl. Safe to edit.

package handler

import (
	rest "github.com/jieyu-c/jieyuc-common/types"
	"github.com/jieyuc/inkforge/services/console/internal/logic/namespace"
	"github.com/jieyuc/inkforge/services/console/internal/svc"
	"net/http"
)

func ListNamespacesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		l := namespace.NewListNamespacesLogic(r.Context(), svcCtx)
		resp, err := l.ListNamespaces()
		rest.Response(w, resp, err)

	}
}
