package consolepath

import (
	"net/http"
	"strings"

	"github.com/jieyuc/inkforge/services/console/internal/pkg/apperr"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/nslug"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// TrimmedNsSlug parses the :nsSlug binding from the path.
func TrimmedNsSlug(r *http.Request) (string, error) {
	var p struct {
		NsSlug string `path:"nsSlug"`
	}
	if err := httpx.ParsePath(r, &p); err != nil {
		return "", apperr.BadRequest("BAD_REQUEST", "Invalid path parameters")
	}
	slug := strings.TrimSpace(p.NsSlug)
	return nslug.MustValidate(slug)
}
