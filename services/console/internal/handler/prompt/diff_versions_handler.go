// Code scaffolded by goctl. Safe to edit.

package handler

import (
	"net/http"

	rest "github.com/jieyu-c/jieyuc-common/types"
	"github.com/jieyuc/inkforge/services/console/internal/logic/prompt"
	"github.com/jieyuc/inkforge/services/console/internal/svc"
	"github.com/jieyuc/inkforge/services/console/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func DiffVersionsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DiffVersionsReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := prompt.NewDiffVersionsLogic(r.Context(), svcCtx)
		resp, err := l.DiffVersions(&req)
		rest.Response(w, resp, err)
	}
}
