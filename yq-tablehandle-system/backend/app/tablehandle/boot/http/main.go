package main

import (
	"github.com/gogf/gf/v2/os/gctx"
	"multi-system/yq-tablehandle-system/backend/app/tablehandle/internal/cmd"
)

func main() { cmd.Http.Run(gctx.GetInitCtx()) }
