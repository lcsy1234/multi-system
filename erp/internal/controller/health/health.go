package health

import (
	"context"
	"multi-system/erp/api/health"
	v1 "multi-system/erp/api/health/v1"
)

type ControllerV1 struct{}

func NewV1() health.IHealthV1 { return &ControllerV1{} }

func (c *ControllerV1) Health(ctx context.Context, req *v1.HealthReq) (*v1.HealthRes, error) {
	return &v1.HealthRes{Service: "erp", Status: "ok", Role: "ERP document integration simulator"}, nil
}
