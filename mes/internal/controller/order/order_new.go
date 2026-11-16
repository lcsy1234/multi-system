package order

import "multi-system/mes/api/order"

type ControllerV1 struct{}

var _ order.IOrderV1 = (*ControllerV1)(nil)

func NewV1() *ControllerV1 {
	return &ControllerV1{}
}
