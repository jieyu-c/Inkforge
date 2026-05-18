// Code scaffolded by goctl. Safe to edit.

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"github.com/jieyuc/inkforge/services/console/internal/config"
	"github.com/jieyuc/inkforge/services/console/internal/handler"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/apperr"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/requestscope"
	"github.com/jieyuc/inkforge/services/console/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

var configFile = flag.String("f", "etc/console-api.yaml", "the config file")

func main() {
	flag.Parse()

	httpx.SetErrorHandlerCtx(func(ctx context.Context, err error) (int, any) {
		var he *apperr.HTTP
		if errors.As(err, &he) {
			return he.Status, apperr.Body{Code: he.Code, Message: he.Message}
		}
		return 500, apperr.Body{Code: "INTERNAL_ERROR", Message: "An unexpected error occurred"}
	})

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	server.Use(requestscope.Middleware)

	serviceCtx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, serviceCtx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
