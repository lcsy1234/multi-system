package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// The daily view has one inventory balance per factory/day and three pairs of
// shift usage columns. Never aggregate balances or temperatures across days.
const fructose = "dws_fructose_receipt_usage_daily"

var fructoseFields = []string{"接收日期", "接收量", "结存量", "果糖生产日期", "接收人", "接收管路检查记录", "1号罐储存温度", "2号罐储存温度", "使用日期", "一组白班用量", "一组夜班用量", "二组白班用量", "二组夜班用量", "三组白班用量", "三组夜班用量", "日结存"}
var fructoseGroups = []struct{ Line, SKU string }{{lines[0], skus[0]}, {lines[1], skus[1]}, {lines[0], skus[2]}}

func fructoseSource() Source {
	fields := map[string]string{"接收日期": "到货验收日期", "接收量": "当日接收量（吨）", "结存量": "当日期初库存（吨，演示口径）", "果糖生产日期": "所收原料的生产日期", "接收人": "模拟验收人员", "接收管路检查记录": "模拟管路检查结果", "1号罐储存温度": "1号储罐温度（℃）", "2号罐储存温度": "2号储罐温度（℃）", "使用日期": "领用日期", "日结存": "期初库存加接收量减少全厂白夜班用量（吨）"}
	for i, g := range fructoseGroups {
		for shift := 0; shift < 2; shift++ {
			fields[fructoseFields[9+2*i+shift]] = g.Line + " / " + g.SKU + " 班次使用量（吨）"
		}
	}
	return Source{fructose, "工厂 × 日期，每日一行，三组线别品项各含白夜班", fields}
}
func seedFructose(now time.Time) []Record {
	var rows []Record
	for i, p := range factories {
		balance := 250 + i*20 // tenths of a tonne
		for d := 0; d < 14; d++ {
			t := now.AddDate(0, 0, d-13)
			dt := date(t)
			received := 105 + (i+d)%8
			r := Record{day: dt, plant: p, "接收日期": dt, "使用日期": dt, "果糖生产日期": date(t.AddDate(0, 0, -3)), "接收量": float64(received) / 10, "结存量": float64(balance) / 10, "接收人": []string{"模拟员工甲", "模拟员工乙"}[d%2], "接收管路检查记录": "√", "1号罐储存温度": 25 + (d+i)%3, "2号罐储存温度": 26 + d%2}
			balance += received
			for k := 9; k < 15; k++ {
				qty := 14 + (d+k)%7
				balance -= qty
				r[fructoseFields[k]] = float64(qty) / 10
			}
			r["日结存"] = float64(balance) / 10
			rows = append(rows, r)
		}
	}
	return rows
}
func groupSelected(i int, f Filters) bool {
	g := fructoseGroups[i]
	return (f.Line == "" || f.Line == g.Line) && (f.SKU == "" || f.SKU == g.SKU)
}
func fructoseRecords(db []Record, f Filters) ([]Record, map[string]bool) {
	var rows []Record
	days := map[string]bool{}
	for _, r := range db {
		dt := r[day].(string)
		if dt < f.Start || dt > f.End || r[plant] != f.Factory {
			continue
		}
		v := Record{}
		for k, x := range r {
			v[k] = x
		}
		matched := false
		for i := range fructoseGroups {
			if groupSelected(i, f) {
				matched = true
			} else {
				delete(v, fructoseFields[9+2*i])
				delete(v, fructoseFields[10+2*i])
			}
		}
		if matched {
			rows = append(rows, v)
			days[dt] = true
		}
	}
	return rows, days
}
func headerText(s string) string { return strings.Join(strings.Fields(s), "") }
func fructoseHeader(b *excelize.File, s string, row int) bool {
	labels := [][]string{{"接收日期", "果糖量（T）", "", "生产日期", "接收人", "接收管路检查记录", "储存温度(℃)", "", "使用日期", "线别：品项：", "", "线别：品项：", "", "线别：品项：", "", "日结存（T）"}, {"", "接收量", "结存量", "", "", "", "1#", "2#", "", "使用量（T）", "", "使用量（T）", "", "使用量（T）"}, {"", "", "", "", "", "", "", "", "", "白班", "夜班", "白班", "夜班", "白班", "夜班"}}
	for offset, cols := range labels {
		for i, want := range cols {
			if want == "" {
				continue
			}
			v, e := b.GetCellValue(s, cell(i+1, row+offset))
			if e != nil || headerText(v) != want {
				return false
			}
		}
	}
	return true
}
func fillFructoseLabels(j *Job) error {
	if j.report.Rows == 0 {
		return nil
	}
	for i, g := range fructoseGroups {
		if !groupSelected(i, j.filters) {
			continue
		}
		a := cell(10+2*i, j.layout.Header)
		v, e := j.book.GetCellValue(j.sheet, a)
		if e != nil {
			return e
		}
		f, e := j.book.GetCellFormula(j.sheet, a)
		if e != nil {
			return e
		}
		if headerText(v) == "线别：品项：" && f == "" {
			if e = j.book.SetCellValue(j.sheet, a, fmt.Sprintf("线别：%s\n品项：%s", g.Line, g.SKU)); e != nil {
				return e
			}
		}
	}
	a := cell(10, j.layout.Header-1)
	v, e := j.book.GetCellValue(j.sheet, a)
	if e != nil {
		return e
	}
	formula, e := j.book.GetCellFormula(j.sheet, a)
	if e != nil {
		return e
	}
	if headerText(v) == "年月日—年月日" && formula == "" {
		return j.book.SetCellValue(j.sheet, a, j.filters.Start+" — "+j.filters.End)
	}
	return nil
}
