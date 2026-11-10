package main

import (
	"github.com/gogf/gf/v2/os/gctx"
	"multi-system/mes/internal/cmd"
)

func main() {
	ctx := gctx.GetInitCtx()
	cmd.Listener.Run(ctx)
}
