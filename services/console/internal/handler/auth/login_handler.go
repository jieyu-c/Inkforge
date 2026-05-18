package auth

import (
	"net/http"

	rest "github.com/jieyu-c/jieyuc-common/types"
	"github.com/jieyuc/inkforge/services/console/internal/logic/auth"
	"github.com/jieyuc/inkforge/services/console/internal/svc"
	"github.com/jieyuc/inkforge/services/console/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func LoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LoginReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := auth.NewLoginLogic(r.Context(), svcCtx)
		resp, err := l.Login(&req)
		rest.Response(w, resp, err)

	}
}
