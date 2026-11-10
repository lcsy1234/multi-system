//go:build ignore

// 独立入口：go run ./boot/socket.go。按文件编译，不参与 boot 目录的默认构建。
package main

import (
	"github.com/gogf/gf/v2/os/gctx"
	"multi-system/mes/internal/cmd"
)

func main() {
	cmd.Socket.Run(gctx.GetInitCtx())
}
