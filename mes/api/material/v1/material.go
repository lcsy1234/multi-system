package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	"multi-system/mes/internal/model"
)

type CatalogReq struct {
	g.Meta `path:"/api/v1/catalog" method:"get" tags:"Material" summary:"Demo production data"`
}
type CatalogRes struct {
	Processes  []model.Process   `json:"processes"`
	Materials  []model.Material  `json:"materials"`
	Warehouses []model.Warehouse `json:"warehouses"`
}
type CreateReq struct {
	g.Meta `path:"/api/v1/material-requisitions" method:"post" tags:"Material" summary:"Create material requisition and enqueue task"`
	model.Submission
}
type CreateRes struct {
	Record *model.Record `json:"record"`
}
type ListReq struct {
	g.Meta `path:"/api/v1/material-requisitions" method:"get" tags:"Material" summary:"List requisitions and async state"`
}
type ListRes struct {
	Records []model.Record `json:"records"`
}
