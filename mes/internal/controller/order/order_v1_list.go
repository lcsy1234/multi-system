package order

import (
	"context"

	v1 "multi-system/mes/api/order/v1"
	"multi-system/mes/internal/model"
)

func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	return &v1.ListRes{Orders: model.Orders}, nil
}
