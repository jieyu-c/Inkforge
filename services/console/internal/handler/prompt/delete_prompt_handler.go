// Code scaffolded by goctl. Safe to edit.

package handler

import (
	"net/http"

	rest "github.com/jieyu-c/jieyuc-common/types"
	"github.com/jieyuc/inkforge/services/console/internal/logic/prompt"
	"github.com/jieyuc/inkforge/services/console/internal/svc"
)

func DeletePromptHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		l := prompt.NewDeletePromptLogic(r.Context(), svcCtx)
		resp, err := l.DeletePrompt()
		rest.Response(w, resp, err)

	}
}
