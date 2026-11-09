package material

import (
	"context"

	"multi-system/mes/api/material/v1"
	"multi-system/mes/internal/model"
)

func (c *ControllerV1) Catalog(ctx context.Context, req *v1.CatalogReq) (res *v1.CatalogRes, err error) {
	return &v1.CatalogRes{Processes: model.Processes, Materials: model.Materials, Warehouses: model.Warehouses}, nil
}
