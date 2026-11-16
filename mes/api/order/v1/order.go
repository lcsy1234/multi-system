package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	"multi-system/mes/internal/model"
)

type ListReq struct {
	g.Meta `path:"/api/v1/orders" method:"get" tags:"Order" summary:"List demo production orders"`
}
type ListRes struct {
	Orders []model.Order `json:"orders"`
}
