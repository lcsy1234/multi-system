package health

import (
	"context"
	v1 "multi-system/wms/api/health/v1"
)

type IHealthV1 interface {
	Health(context.Context, *v1.HealthReq) (*v1.HealthRes, error)
}
