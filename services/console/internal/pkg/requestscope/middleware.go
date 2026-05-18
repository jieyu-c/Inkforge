package requestscope

import "net/http"

// Middleware injects [*HTTP] into request context before route handlers and logic execute.
func Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		next(w, r.WithContext(With(r.Context(), w, r)))
	}
}
