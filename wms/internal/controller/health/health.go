package health

import (
	"context"
	"multi-system/wms/api/health"
	v1 "multi-system/wms/api/health/v1"
)

type ControllerV1 struct{}

func NewV1() health.IHealthV1 { return &ControllerV1{} }

func (c *ControllerV1) Health(ctx context.Context, req *v1.HealthReq) (*v1.HealthRes, error) {
	return &v1.HealthRes{Service: "wms", Status: "ok", Role: "Warehouse inventory and material issuing"}, nil
}
