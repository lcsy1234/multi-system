package cmd

import (
	"context"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"
	"multi-system/mes/internal/boot"
	"multi-system/mes/internal/controller/health"
	"multi-system/mes/internal/controller/material"
	"multi-system/mes/internal/storage"
	"multi-system/mes/internal/web"
	"os"
)

var Main = gcmd.Command{
	Name:  "mes",
	Usage: "mes",
	Brief: "Manufacturing execution and material requisitions",
	Func: func(ctx context.Context, parser *gcmd.Parser) error {
		service, err := boot.New(ctx)
		if err != nil {
			return err
		}
		defer boot.Close(service)
		materialController := material.NewV1()
		materialController.Service = service
		s := g.Server()
		if address := os.Getenv("HTTP_ADDR"); address != "" {
			s.SetAddr(address)
		}
		assets := web.Handler()
		for _, path := range []string{"/", "/assets/*"} {
			s.BindHandler("GET:"+path, func(r *ghttp.Request) { assets.ServeHTTP(r.Response.BufferWriter, r.Request) })
		}
		s.Group("/", func(group *ghttp.RouterGroup) {
			group.Middleware(ghttp.MiddlewareHandlerResponse)
			group.Bind(health.NewV1(), materialController)
		})
		s.Run()
		return nil
	},
}

var Worker = gcmd.Command{Name: "worker", Usage: "worker", Brief: "Run MES Asynq material worker", Func: func(ctx context.Context, parser *gcmd.Parser) error { return boot.Worker(ctx) }}

var Migrate = gcmd.Command{Name: "migrate", Brief: "Create MES demo MySQL schema", Func: func(ctx context.Context, parser *gcmd.Parser) error { return storage.Migrate(ctx) }}
