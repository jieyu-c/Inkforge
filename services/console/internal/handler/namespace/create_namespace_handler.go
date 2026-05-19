// Code scaffolded by goctl. Safe to edit.

package handler

import (
	rest "github.com/jieyu-c/jieyuc-common/types"
	"github.com/jieyuc/inkforge/services/console/internal/logic/namespace"
	"github.com/jieyuc/inkforge/services/console/internal/svc"
	"github.com/jieyuc/inkforge/services/console/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
)

func CreateNamespaceHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateNamespaceReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := namespace.NewCreateNamespaceLogic(r.Context(), svcCtx)
		resp, err := l.CreateNamespace(&req)
		rest.Response(w, resp, err)

	}
}
