package main

import (
	"github.com/gogf/gf/v2/os/gctx"
	"multi-system/mes/internal/cmd"
)

// main 启动 MES 主进程，提供页面和 HTTP 接口，不启动异步任务消费。
func main() {
	if err := cmd.Main.AddCommand(&cmd.Migrate, &cmd.AsyncDemo); err != nil {
		panic(err)
	}
	cmd.Main.Run(gctx.GetInitCtx())
}
