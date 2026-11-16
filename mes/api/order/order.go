package order

import (
	"context"

	"multi-system/mes/api/order/v1"
)

type IOrderV1 interface {
	List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error)
}
