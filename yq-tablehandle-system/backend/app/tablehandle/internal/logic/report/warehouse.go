package report

import (
	"fmt"
	"slices"
	"strings"
	"time"
)

const production = "dws_factory_production_daily"
const equipment = "dwd_equipment_oee_daily"
const inbound = "db_inventory_inbound"

var factories = []string{"滁州工厂", "肇庆工厂", "天津工厂"}
var lines = []string{"L01", "L02"}
var skus = []string{"白桃味气泡水", "青柠味气泡水", "燃茶乌龙"}

const day = "日期"
const plant = "工厂名称"
const line = "产线编号"
const sku = "SKU名称"

var dimensions = []string{day, plant, line, sku}

type Source struct {
	Name, Grain string
	Fields      map[string]string
}

var catalog = []Source{
	fructoseSource(),
	{production, "日期 × 工厂 × 产线 × SKU", map[string]string{"计划产量": "计划生产箱数", "实际产量": "实际下线箱数", "合格产量": "合格品箱数", "损耗箱数": "实际减合格", "达成率": "实际合计 / 计划合计", "良品率": "合格合计 / 实际合计"}},
	{equipment, "日期 × 工厂 × 产线", map[string]string{"运行时长": "运行小时数", "停机时长": "停机小时数", "OEE": "按运行时长加权的设备综合效率", "能耗": "耗电量（度）"}},
	{inbound, "入库日期 × 工厂 × SKU", map[string]string{"入库箱数": "已完成入库箱数", "质检状态": "同组出现多种状态时显示混合"}},
}

type Record map[string]any
type Filters struct{ Factory, Start, End, Line, SKU, Note string }

func today() time.Time        { return time.Now().In(time.FixedZone("CST", 8*3600)) }
func date(t time.Time) string { return t.Format(time.DateOnly) }
func (f Filters) Validate() error {
	a, e := time.Parse(time.DateOnly, f.Start)
	b, ee := time.Parse(time.DateOnly, f.End)
	if e != nil || ee != nil || b.Before(a) || b.Sub(a) > 365*24*time.Hour {
		return fmt.Errorf("请选择最多 366 天的有效日期范围")
	}
	if !slices.Contains(factories, f.Factory) || (f.Line != "" && !slices.Contains(lines, f.Line)) || (f.SKU != "" && !slices.Contains(skus, f.SKU)) || len([]rune(f.Note)) > 600 {
		return fmt.Errorf("筛选条件无效或说明过长")
	}
	return nil
}
func seed(now time.Time) map[string][]Record {
	db := map[string][]Record{fructose: seedFructose(now)}
	for d := 0; d < 14; d++ {
		dt := date(now.AddDate(0, 0, d-13))
		for i, p := range factories {
			for k, s := range skus {
				db[inbound] = append(db[inbound], Record{day: dt, plant: p, sku: s, "入库箱数": float64(2100 + d*7 + i*8 + k), "质检状态": "合格"})
			}
			for j, l := range lines {
				run := float64(17 + (i+j+d)%5)
				db[equipment] = append(db[equipment], Record{day: dt, plant: p, line: l, "运行时长": run, "停机时长": 24 - run, "OEE": 0.81 + float64(d%7)/100, "能耗": run * 114})
				for k, s := range skus {
					plan := float64(1000 + i*89 + j*101 + k*61)
					actual := plan + float64((d*31+k*7)%99-42)
					loss := float64(3 + (i+j+d+k)%13)
					db[production] = append(db[production], Record{day: dt, plant: p, line: l, sku: s, "计划产量": plan, "实际产量": actual, "合格产量": actual - loss, "损耗箱数": loss})
				}
			}
		}
	}
	return db
}
func canon(h string) string {
	h = strings.NewReplacer("_箱", "", "（箱）", "", "(箱)", "", "_小时", "", "（小时）", "", "_度", "", "（度）", "", " ", "", "（%）", "", "(%)", "").Replace(strings.TrimSpace(h))
	switch h {
	case "入库日期", "生产日期":
		return day
	case "工厂":
		return plant
	case "产线":
		return line
	case "SKU", "产品名称":
		return sku
	}
	return h
}
func groupRecords(db []Record, f Filters, dims []string) ([]Record, map[string]bool) {
	groups := map[string]Record{}
	var ordered []Record
	days := map[string]bool{}
	for _, r := range db {
		dt := r[day].(string)
		if dt < f.Start || dt > f.End || r[plant] != f.Factory || (f.Line != "" && r[line] != f.Line) || (f.SKU != "" && r[sku] != f.SKU) {
			continue
		}
		days[dt] = true
		var parts []string
		for _, k := range dims {
			parts = append(parts, fmt.Sprint(r[k]))
		}
		key := strings.Join(parts, "\x00")
		g := groups[key]
		if g == nil {
			g = Record{}
			groups[key] = g
			ordered = append(ordered, g)
			for _, k := range dims {
				g[k] = r[k]
			}
		}
		for k, v := range r {
			if slices.Contains(dimensions, k) {
				continue
			}
			switch n := v.(type) {
			case float64:
				prev, _ := g[k].(float64)
				if k == "OEE" {
					n *= r["运行时长"].(float64)
				}
				g[k] = prev + n
			case string:
				if old := g[k]; old != nil && old != n {
					g[k] = "混合"
				} else {
					g[k] = n
				}
			}
		}
	}
	for _, g := range ordered {
		if p, ok := g["计划产量"].(float64); ok {
			a := g["实际产量"].(float64)
			g["达成率"] = a / p
			g["良品率"] = g["合格产量"].(float64) / a
		}
		if e, ok := g["OEE"].(float64); ok {
			g["OEE"] = e / g["运行时长"].(float64)
		}
	}
	return ordered, days
}
