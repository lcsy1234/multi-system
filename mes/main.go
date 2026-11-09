package main

import (
	"github.com/gogf/gf/v2/os/gctx"

	"multi-system/mes/internal/cmd"
)

func main() {
	if err := cmd.Main.AddCommand(&cmd.Worker, &cmd.Migrate); err != nil {
		panic(err)
	}
	cmd.Main.Run(gctx.GetInitCtx())
}
