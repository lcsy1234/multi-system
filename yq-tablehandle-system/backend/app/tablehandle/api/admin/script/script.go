package script

import (
	"context"
	v1 "multi-system/yq-tablehandle-system/backend/app/tablehandle/api/admin/script/v1"
)

type IScriptV1 interface {
	Optimize(context.Context, *v1.OptimizeReq) (*v1.OptimizeRes, error)
}
