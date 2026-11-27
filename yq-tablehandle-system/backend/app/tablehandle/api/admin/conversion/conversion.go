package conversion

import (
	"context"
	v1 "multi-system/yq-tablehandle-system/backend/app/tablehandle/api/admin/conversion/v1"
)

type IConversionV1 interface {
	Convert(context.Context, *v1.ConvertReq) (*v1.ConvertRes, error)
	ConvertShrimp(context.Context, *v1.ConvertShrimpReq) (*v1.ConvertShrimpRes, error)
}
