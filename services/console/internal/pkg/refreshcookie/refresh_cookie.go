package refreshcookie

import (
	"net/http"
	"time"

	"github.com/jieyuc/inkforge/services/console/internal/config"
)

func template(cfg config.Config) http.Cookie {
	var ss http.SameSite
	if cfg.Auth.RefreshCookieSameSiteStrict {
		ss = http.SameSiteStrictMode
	} else {
		ss = http.SameSiteLaxMode
	}
	return http.Cookie{
		Name:     cfg.Auth.RefreshCookieName,
		Path:     cfg.Auth.RefreshCookiePath,
		HttpOnly: true,
		Secure:   cfg.Auth.RefreshCookieSecure,
		SameSite: ss,
	}
}

// Set persists the opaque refresh marker; Path is narrowed via config (typically /api/v1/auth).
func Set(w http.ResponseWriter, cfg config.Config, token string, maxAgeSec int) {
	ck := template(cfg)
	ck.Value = token
	ck.MaxAge = maxAgeSec
	http.SetCookie(w, &ck)
}

func Clear(w http.ResponseWriter, cfg config.Config) {
	ck := template(cfg)
	ck.Value = ""
	ck.MaxAge = -1
	ck.Expires = time.Unix(0, 0).UTC()
	http.SetCookie(w, &ck)
}
