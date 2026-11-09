package cmd

import (
	"context"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"
	"multi-system/scm/internal/controller/health"
)

var Main = gcmd.Command{
	Name:  "scm",
	Usage: "scm",
	Brief: "Transfers and production order fulfillment",
	Func: func(ctx context.Context, parser *gcmd.Parser) error {
		s := g.Server()
		s.Group("/", func(group *ghttp.RouterGroup) {
			group.Middleware(ghttp.MiddlewareHandlerResponse)
			group.Bind(health.NewV1())
		})
		s.Run()
		return nil
	},
}
