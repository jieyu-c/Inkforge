package user

import (
	"net/http"

	rest "github.com/jieyu-c/jieyuc-common/types"
	"github.com/jieyuc/inkforge/services/console/internal/logic/user"
	"github.com/jieyuc/inkforge/services/console/internal/svc"
)

func MeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		l := user.NewMeLogic(r.Context(), svcCtx)
		resp, err := l.Me()
		rest.Response(w, resp, err)

	}
}
