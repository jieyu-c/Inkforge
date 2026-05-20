// Code scaffolded by goctl. Safe to edit.

package handler

import (
	rest "github.com/jieyu-c/jieyuc-common/types"
	"github.com/jieyuc/inkforge/services/console/internal/logic/prompt"
	"github.com/jieyuc/inkforge/services/console/internal/svc"
	"github.com/jieyuc/inkforge/services/console/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
)

func PatchChannelPointerHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PatchChannelPointerReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := prompt.NewPatchChannelPointerLogic(r.Context(), svcCtx)
		resp, err := l.PatchChannelPointer(&req)
		rest.Response(w, resp, err)

	}
}
