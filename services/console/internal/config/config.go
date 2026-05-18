package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	JwtAuth struct {
		AccessSecret string `json:",optional"`
		AccessExpire int64  `json:",default=900"`
	}
	Redis redis.RedisConf `json:",optional"`
	Mysql struct {
		DataSource string
	}
	Auth struct {
		PublicBaseURL                 string `json:",optional"`
		RefreshTTLSeconds             int64  `json:",default=604800"`
		RefreshCookieName             string `json:",default=jf_refresh"`
		RefreshCookiePath             string `json:",default=/api/v1/auth"`
		RefreshCookieSecure           bool   `json:",optional"`
		RefreshCookieSameSiteStrict   bool   `json:",optional"` // if true Strict, else Lax
		BcryptCost                    int    `json:",default=10"`
		LockFailures                  int    `json:",default=8"`
		LockDurationMinutes           int    `json:",default=15"`
		MinPasswordLength             int    `json:",default=10"`
		ForgotPasswordThrottleSeconds int    `json:",optional"` // unused in MVP phone path
		ResetTokenTTLMinutes          int    `json:",optional"` // unused
		SAMLAssertionMaxAgeMinutes    int    `json:",optional"` // unused
	}
	RateLimit struct {
		AuthIPRequestsPerMinute    float64 `json:",default=30"`
		AuthIPBurst                int     `json:",default=60"`
		AuthPhoneRequestsPerMinute float64 `json:",default=10"`
		AuthPhoneBurst             int     `json:",default=20"`
	}
}
