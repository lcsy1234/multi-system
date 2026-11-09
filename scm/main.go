package main

import (
	"github.com/gogf/gf/v2/os/gctx"

	"multi-system/scm/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}
