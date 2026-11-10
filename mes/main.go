package main

import (
	"github.com/gogf/gf/v2/os/gctx"

	"multi-system/mes/internal/cmd"
)

func main() {
	// 主进程默认启动 HTTP 服务；异步消费者由 listener.go 独立启动。
	if err := cmd.Main.AddCommand(&cmd.Migrate, &cmd.AsyncDemo); err != nil {
		panic(err)
	}
	cmd.Main.Run(gctx.GetInitCtx())
}
