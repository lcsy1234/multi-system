package health

import (
	"context"

	"multi-system/mes/api/health/v1"
)

func (c *ControllerV1) Health(ctx context.Context, req *v1.HealthReq) (res *v1.HealthRes, err error) {
	return &v1.HealthRes{Service: "mes", Status: "ok", Role: "Manufacturing execution and material requisitions"}, nil
}
