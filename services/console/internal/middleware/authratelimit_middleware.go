// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/jieyuc/inkforge/services/console/internal/config"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/apperr"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/phone"
	"github.com/zeromicro/go-zero/core/limit"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest/httpx"
)

const authPrefix = "/api/v1/auth"
const maxPeekBody = 1 << 20

type limiterLRU struct {
	cap     int
	mu      sync.Mutex
	order   []string
	entries map[string]*limit.TokenLimiter
}

func newLimiterLRU(cap int) *limiterLRU {
	if cap <= 0 {
		cap = 4096
	}
	return &limiterLRU{
		cap:     cap,
		order:   make([]string, 0, cap),
		entries: make(map[string]*limit.TokenLimiter),
	}
}

func (l *limiterLRU) allow(ctx context.Context, store *redis.Redis, rate, burst int, key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	lim, ok := l.entries[key]
	if !ok {
		lim = limit.NewTokenLimiter(rate, burst, store, key)
		if len(l.order) >= l.cap {
			evict := l.order[0]
			l.order = l.order[1:]
			delete(l.entries, evict)
		}
		l.entries[key] = lim
		l.order = append(l.order, key)
	} else {
		for i, k := range l.order {
			if k == key {
				l.order = append(l.order[:i], l.order[i+1:]...)
				break
			}
		}
		l.order = append(l.order, key)
	}
	return lim.AllowNCtx(ctx, time.Now(), 1)
}

func perSecondFromPerMinute(perMinute float64) int {
	if perMinute <= 0 {
		return 1
	}
	v := int(math.Ceil(perMinute / 60.0))
	if v < 1 {
		return 1
	}
	return v
}

func burstAtLeast(burst, rate int) int {
	if burst < 1 {
		burst = 1
	}
	if burst < rate {
		return rate
	}
	return burst
}

func authRouteKind(path string) string {
	if len(path) < len(authPrefix)+2 || path[:len(authPrefix)] != authPrefix {
		return ""
	}
	trim := strings.TrimPrefix(path, authPrefix)
	trim = strings.TrimPrefix(trim, "/")
	idx := strings.IndexByte(trim, '/')
	if idx >= 0 {
		trim = trim[:idx]
	}
	switch trim {
	case "register", "login", "refresh", "logout":
		return trim
	default:
		return ""
	}
}

func sanitizeKeyPart(s string) string {
	if s == "" {
		return "_"
	}
	b := []byte(strings.TrimSpace(s))
	for i := range b {
		switch b[i] {
		case '{', '}', ':', '\n', '\r', ' ', '\t':
			b[i] = '_'
		}
	}
	return string(b)
}

func write429(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusTooManyRequests)
	body := apperr.Body{
		Code:    "RATE_LIMITED",
		Message: "Too many requests",
	}
	_ = json.NewEncoder(w).Encode(body)
}

type phonePeek struct {
	Phone string `json:"phone"`
}

func peekLoginRegisterPhone(r *http.Request) (readCloser io.ReadCloser, canon string, err error) {
	var buf bytes.Buffer
	if _, err = io.Copy(&buf, io.LimitReader(r.Body, maxPeekBody)); err != nil {
		_ = r.Body.Close()
		return io.NopCloser(bytes.NewReader(nil)), "", err
	}
	_ = r.Body.Close()
	readCloser = io.NopCloser(bytes.NewReader(buf.Bytes()))

	var peek phonePeek
	if err := json.Unmarshal(buf.Bytes(), &peek); err != nil {
		return readCloser, "", nil
	}
	return readCloser, phone.Canonical(strings.TrimSpace(peek.Phone)), nil
}

type AuthRateLimitMiddleware struct {
	store      *redis.Redis
	ipCache    *limiterLRU
	phoneCache *limiterLRU
	ipRate     int
	ipBurst    int
	phoneRate  int
	phoneBurst int
}

func NewAuthRateLimitMiddleware(cfg config.Config, store *redis.Redis) *AuthRateLimitMiddleware {
	ipRate := perSecondFromPerMinute(cfg.RateLimit.AuthIPRequestsPerMinute)
	ipBurst := burstAtLeast(cfg.RateLimit.AuthIPBurst, ipRate)
	phoneRate := perSecondFromPerMinute(cfg.RateLimit.AuthPhoneRequestsPerMinute)
	phoneBurst := burstAtLeast(cfg.RateLimit.AuthPhoneBurst, phoneRate)

	return &AuthRateLimitMiddleware{
		store:      store,
		ipCache:    newLimiterLRU(8192),
		phoneCache: newLimiterLRU(8192),
		ipRate:     ipRate,
		ipBurst:    ipBurst,
		phoneRate:  phoneRate,
		phoneBurst: phoneBurst,
	}
}

func (m *AuthRateLimitMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		kind := authRouteKind(r.URL.Path)
		if kind == "" {
			next(w, r)
			return
		}

		ctx := r.Context()
		ipKey := sanitizeKeyPart(httpx.GetRemoteAddr(r))
		ipRLKey := "console_auth_ip_" + kind + "_" + ipKey
		if !m.ipCache.allow(ctx, m.store, m.ipRate, m.ipBurst, ipRLKey) {
			write429(w)
			return
		}

		if kind == "register" || kind == "login" {
			bodyRC, canon, peekErr := peekLoginRegisterPhone(r)
			r.Body = bodyRC
			if peekErr != nil {
				next(w, r)
				return
			}
			if canon != "" {
				acctRLKey := "console_auth_acct_" + kind + "_" + sanitizeKeyPart(canon)
				if !m.phoneCache.allow(ctx, m.store, m.phoneRate, m.phoneBurst, acctRLKey) {
					write429(w)
					return
				}
			}
		}

		next(w, r)
	}
}
