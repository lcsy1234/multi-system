package cmd

import (
	"context"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"
	"multi-system/mes/internal/controller/health"
)

var Main = gcmd.Command{
	Name:  "mes",
	Usage: "mes",
	Brief: "Manufacturing execution and material requisitions",
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
