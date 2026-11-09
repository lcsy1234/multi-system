// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package material

import (
	"context"

	"multi-system/mes/api/material/v1"
)

type IMaterialV1 interface {
	Catalog(ctx context.Context, req *v1.CatalogReq) (res *v1.CatalogRes, err error)
	Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error)
	List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error)
}
