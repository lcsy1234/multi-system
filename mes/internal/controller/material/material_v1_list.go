package material

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"multi-system/mes/api/material/v1"
)

func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	rows, err := c.Service.List(ctx)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeOperationFailed, "MySQL 不可用，请检查连接并执行 make db-init："+err.Error())
	}
	return &v1.ListRes{Records: rows}, nil
}
