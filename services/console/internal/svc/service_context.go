package svc

import (
	"errors"

	"github.com/jieyuc/inkforge/services/console/internal/config"
	"github.com/jieyuc/inkforge/services/console/internal/middleware"
	"github.com/jieyuc/inkforge/services/console/internal/model"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/repo"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/rest"
)

type ServiceContext struct {
	Config               config.Config
	DB                   sqlx.SqlConn
	Users                *repo.Users
	Sessions             *repo.Sessions
	ConsoleNamespacesModel model.ConsoleNamespacesModel
	AuthRateLimit        rest.Middleware
	TenantCtx            rest.Middleware
}

func NewServiceContext(c config.Config) *ServiceContext {
	if c.Mysql.DataSource == "" {
		panic("Mysql.DataSource is required")
	}
	if c.Redis.Host == "" {
		logx.Must(errors.New(`Redis.Host is required: auth middleware uses go-zero/core/limit TokenLimiter (Redis-backed)`))
	}

	db := sqlx.NewMysql(c.Mysql.DataSource)
	rds := redis.MustNewRedis(c.Redis)

	return &ServiceContext{
		Config:               c,
		DB:                   db,
		Users:                repo.NewUsers(db),
		Sessions:             repo.NewSessions(db),
		ConsoleNamespacesModel: model.NewConsoleNamespacesModel(db),
		AuthRateLimit:        middleware.NewAuthRateLimitMiddleware(c, rds).Handle,
		TenantCtx:            middleware.NewTenantCtxMiddleware().Handle,
	}
}
