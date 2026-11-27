package tablehandle

import (
	"github.com/gogf/gf/v2/os/gctx"
	"multi-system/yq-tablehandle-system/backend/app/tablehandle/internal/cmd"
)

func RunHTTP() { cmd.Http.Run(gctx.GetInitCtx()) }
