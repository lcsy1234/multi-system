package model

type Order struct {
	ID           string `json:"id"`
	FactoryName  string `json:"factoryName"`
	FactoryCode  string `json:"factoryCode"`
	LineName     string `json:"lineName"`
	LineCode     string `json:"lineCode"`
	ProductName  string `json:"productName"`
	ProductCode  string `json:"productCode"`
	PlanDate     string `json:"planDate"`
	PlanQuantity int    `json:"planQuantity"`
}

var Orders = []Order{
	{"SCDD26111320060047", "江苏厂", "GC2006", "江苏混合二线", "GC2006H02", "元气森林-无糖气泡水-白桃 PET480ml*15 纸箱装", "YQ03002", "2026-11-13 00:00", 1190},
	{"SCDD26111320060048", "天津厂", "GC2001", "天津碳酸一线", "GC2001T01", "元气森林-无糖气泡水-白桃 PET480ml*15 纸箱装", "YQ03002", "2026-11-13 00:00", 10000},
	{"SCDD26111320060049", "江苏厂", "GC2006", "江苏无菌三线", "GC2006W03", "燃茶-乌龙茶-玄米乌龙 PET500ml*15 纸箱装", "YQ02003", "2026-11-13 00:00", 100000},
	{"SCDD26111220060050", "江苏厂", "GC2006", "江苏混合二线", "GC2006H02", "冰茶-红茶-冰红茶 PET500ml*15 纸箱装", "YQ04001", "2026-11-12 08:00", 5000},
	{"SCDD26111120010051", "天津厂", "GC2001", "天津无菌二线", "GC2001W02", "茉莉花茶 PET500ml*15 纸箱装", "YQ02005", "2026-11-11 00:00", 8000},
	{"SCDD26111020060052", "江苏厂", "GC2006", "江苏无菌三线", "GC2006W03", "青柠红茶 PET500ml*15 纸箱装", "YQ02008", "2026-11-10 00:00", 3200},
}
