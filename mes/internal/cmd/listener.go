package cmd

import (
	"context"

	"github.com/gogf/gf/v2/os/gcmd"
	"multi-system/mes/internal/boot"
	"multi-system/mes/internal/listener"
	"multi-system/mes/utility/worker"
)

var Listener = gcmd.Command{
	Name:  "listener",
	Usage: "mes listener",
	Brief: "启动 MES 异步领料任务监听服务",
	Func: func(ctx context.Context, parser *gcmd.Parser) error {
		config, err := boot.ConsumerConfig(ctx)
		if err != nil {
			return err
		}

		service, err := boot.New(ctx)
		if err != nil {
			return err
		}
		defer boot.Close(service)

		if err := service.DB.PingContext(ctx); err != nil {
			return err
		}

		return worker.StartConsumer(
			boot.RedisConfig(ctx),
			config,
			listener.NewMux(service),
		)
	},
}
