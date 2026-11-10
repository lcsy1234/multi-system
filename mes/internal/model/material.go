package model

// Decimal quantities travel as strings to avoid binary floating-point rounding.
type Line struct {
	MaterialCode string `json:"materialCode"`
	Quantity     string `json:"quantity"`
	Remark       string `json:"remark"`
}
type Submission struct {
	ClientRequestID string `json:"clientRequestId"`
	ProcessID       string `json:"processId"`
	WarehouseID     string `json:"warehouseId"`
	Date            string `json:"date"`
	Remark          string `json:"remark"`
	Items           []Line `json:"items"`
}
type Event struct {
	Time    string `json:"time"`
	Message string `json:"message"`
}
type Record struct {
	Submission
	ID        string  `json:"id"`
	Number    string  `json:"number"`
	Status    string  `json:"status"`
	LastError string  `json:"lastError"`
	Attempts  int     `json:"attempts"`
	CreatedAt string  `json:"createdAt"`
	UpdatedAt string  `json:"updatedAt"`
	Events    []Event `json:"events"`
}
type Process struct {
	ID       string `json:"id"`
	Order    string `json:"order"`
	Factory  string `json:"factory"`
	Line     string `json:"line"`
	Product  string `json:"product"`
	PlanDate string `json:"planDate"`
	Status   string `json:"status"`
}
type Material struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	Spec      string `json:"spec"`
	Unit      string `json:"unit"`
	Category  string `json:"category"`
	Suggested string `json:"suggested"`
}
type Warehouse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var Processes = []Process{
	{"PR20261109005", "MO20261109005", "四川工厂", "四川无菌五线", "玄米乌龙 · 领料练习 A PET500ml × 15", "2026-11-09", "已开工"},
	{"PR20261109006", "MO20261109006", "四川工厂", "四川无菌三线", "茉莉花茶 · 领料练习 B PET500ml × 15", "2026-11-09", "已开工"},
	{"PR20261109007", "MO20261109007", "天津工厂", "天津无菌二线", "乌龙茶 · 领料练习 C PET500ml × 15", "2026-11-09", "已开工"},
	{"PR20261109001", "MO20261109001", "四川工厂", "四川无菌五线", "乌龙茶 · 玄米乌龙 PET500ml × 15", "2026-11-09", "已开工"},
	{"PR20261109002", "MO20261109002", "天津工厂", "天津碳酸一线", "白桃气泡水 PET500ml × 15", "2026-11-09", "已开工"},
	{"PR20261108003", "MO20261108003", "四川工厂", "四川无菌三线", "茉莉花茶 PET500ml × 15", "2026-11-08", "变更中"},
	{"PR20261107004", "MO20261107004", "天津工厂", "天津无菌二线", "青柠红茶 PET500ml × 15", "2026-11-07", "已完工"},
}
var Materials = []Material{
	{"RM001", "茶叶萃取液", "食品级 / 20kg 桶", "kg", "生产原料", "120"},
	{"RM002", "白砂糖", "一级 / 50kg 袋", "kg", "生产原料", "80"},
	{"PK001", "PET 瓶胚", "500ml / 28mm", "个", "生产包材", "3000"},
	{"PK002", "包装纸箱", "15 瓶装 / 标准版", "个", "生产包材", "200"},
	{"QC001", "取样瓶", "无菌 / 100ml", "个", "品保用料", "12"},
}
var Warehouses = []Warehouse{{"WH001", "四川原料仓"}, {"WH002", "四川包材仓"}, {"WH003", "天津综合仓"}}
