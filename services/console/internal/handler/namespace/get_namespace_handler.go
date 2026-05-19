// Code scaffolded by goctl. Safe to edit.

package handler

import (
	"net/http"

	rest "github.com/jieyu-c/jieyuc-common/types"
	"github.com/jieyuc/inkforge/services/console/internal/logic/namespace"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/consolepath"
	"github.com/jieyuc/inkforge/services/console/internal/svc"
)

func GetNamespaceHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug, err := consolepath.TrimmedNsSlug(r)
		if err != nil {
			rest.Response(w, nil, err)
			return
		}
		l := namespace.NewGetNamespaceLogic(r.Context(), svcCtx)
		resp, err := l.GetNamespace(slug)
		rest.Response(w, resp, err)
	}
}
