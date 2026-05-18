package auth

import (
	"net/http"

	rest "github.com/jieyu-c/jieyuc-common/types"
	"github.com/jieyuc/inkforge/services/console/internal/logic/auth"
	"github.com/jieyuc/inkforge/services/console/internal/svc"
)

func LogoutHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		l := auth.NewLogoutLogic(r.Context(), svcCtx)
		resp, err := l.Logout()
		rest.Response(w, resp, err)

	}
}
