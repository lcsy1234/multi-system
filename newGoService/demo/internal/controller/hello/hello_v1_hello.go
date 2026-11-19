package hello

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"demo/api/hello/v1"
)

func (c *ControllerV1) Hello(ctx context.Context, req *v1.HelloReq) (res *v1.HelloRes, err error) {
	database, err := g.DB().GetValue(ctx, "SELECT DATABASE()")
	if err != nil {
		return nil, err
	}
	g.RequestFromCtx(ctx).Response.Writeln("Hello World! Connected database: " + database.String())
	return
}
