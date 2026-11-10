package cmd

import (
	"context"

	"github.com/gogf/gf/v2/os/gcmd"
	"multi-system/mes/internal/asynctask"
	"multi-system/mes/internal/bootstrap"
	"multi-system/mes/utility/worker"
)

var Listener = gcmd.Command{
	Name:  "listener",
	Usage: "mes listener",
	Brief: "启动 MES 异步领料任务监听服务",
	Func: func(ctx context.Context, parser *gcmd.Parser) error {
		service, err := bootstrap.New(ctx)
		if err != nil {
			return err
		}
		defer bootstrap.Close(service)

		if err := service.DB.PingContext(ctx); err != nil {
			return err
		}

		asynctask.Init(service)
		return worker.StartConsumer(ctx, asynctask.Processors())
	},
}
