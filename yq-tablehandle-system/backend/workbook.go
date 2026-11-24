package main

import (
	"fmt"
	"github.com/xuri/excelize/v2"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

func cell(c, r int) string { s, _ := excelize.CoordinatesToCellName(c, r); return s }

type Slot struct {
	Row      int
	Existing map[string]string
	Blanks   map[string]string
}
type Layout struct {
	Source   string `json:",omitempty"`
	Sheets   []string
	Sheet    string
	Header   int
	Fields   []string
	Rows     []Slot
	Formulas []string
}

// A vertical total ends the detail region. A horizontal SUM in a detail row
// protects only its own cell; neighboring metrics can still be populated.
var aggregateRange = regexp.MustCompile(`(?i)^=?\s*(?:SUM|AVERAGE)\(\s*\$?[A-Z]+\$?([0-9]+):\$?[A-Z]+\$?([0-9]+)`)

func priorRowTotal(formula string, row int) bool {
	parts := aggregateRange.FindStringSubmatch(formula)
	if parts == nil {
		return false
	}
	start, _ := strconv.Atoi(parts[1])
	end, _ := strconv.Atoi(parts[2])
	return start < row && end < row
}
func inspect(b *excelize.File, s string) (Layout, error) {
	p := Layout{Sheets: b.GetSheetList(), Sheet: s}
	it, err := b.Rows(s)
	if err != nil {
		return p, err
	}
	defer it.Close()
	var rows [][]string
	for it.Next() {
		if len(rows) >= 5000 {
			return p, fmt.Errorf("工作表超过 5000 行")
		}
		values, e := it.Columns()
		if e != nil {
			return p, e
		}
		rows = append(rows, values)
	}
	if err := it.Error(); err != nil {
		return p, err
	}
	// Find the first plausible header with at least one known business field.
	for i, r := range rows {
		if len(r) > 80 {
			return p, fmt.Errorf("工作表超过 80 列")
		}
		for j := range r {
			f, e := b.GetCellFormula(s, cell(j+1, i+1))
			if e != nil {
				return p, e
			}
			if f != "" {
				p.Formulas = append(p.Formulas, cell(j+1, i+1))
			}
		}
		if p.Header == 0 && len(r) > 0 && headerText(r[0]) == "接收日期" {
			if !fructoseHeader(b, s, i+1) {
				if len(r) >= 16 && headerText(r[15]) == "日结存（T）" {
					return p, fmt.Errorf("果糖记录表的班次或线别表头已修改，请上传空白原模板")
				}
			} else {
				p.Source = fructose
				p.Header = i + 1
				p.Fields = fructoseFields
			}
		}
		if p.Header == 0 && len(r) >= 2 {
			for _, h := range r {
				k := canon(h)
				known := slices.Contains(dimensions, k)
				for _, src := range catalog {
					known = known || src.Fields[k] != ""
				}
				if known {
					p.Header = i + 1
					p.Fields = r
					break
				}
			}
		}
	}
	if p.Header == 0 {
		return p, fmt.Errorf("未识别到含日期、工厂或业务指标的表头")
	}
	seen := map[string]bool{}
	for _, h := range p.Fields {
		if h != "" && seen[h] {
			return p, fmt.Errorf("表头重复：%s", h)
		}
		seen[h] = true
	}
	merges, err := b.GetMergeCells(s)
	if err != nil {
		return p, err
	}
	end := len(rows)
	if end == p.Header {
		end++
	}
	start := p.Header + 1
	if p.Source == fructose {
		start = p.Header + 3
		footer := 0
		for i := start; i <= len(rows); i++ {
			if len(rows[i-1]) > 0 && strings.HasPrefix(strings.TrimSpace(rows[i-1][0]), "备注") {
				footer = i
				break
			}
		}
		if footer == 0 {
			return p, fmt.Errorf("果糖模板缺少备注行，无法确定明细边界")
		}
		end = footer - 1
	}
	for i := start; i <= end; i++ {
		row := Slot{i, map[string]string{}, map[string]string{}}
		summary := false
		for j, h := range p.Fields {
			if h == "" {
				continue
			}
			a := cell(j+1, i)
			v, e := b.GetCellValue(s, a)
			if e != nil {
				return p, e
			}
			f, e := b.GetCellFormula(s, a)
			if e != nil {
				return p, e
			}
			if strings.Contains(v, "合计") || strings.Contains(v, "汇总") || strings.Contains(v, "小计") || priorRowTotal(f, i) {
				summary = true
			}
			if v != "" {
				row.Existing[h] = v
				if canon(h) == day {
					if t, e := b.GetCellType(s, a); e == nil && (t == excelize.CellTypeNumber || t == excelize.CellTypeUnset) {
						raw, _ := b.GetCellValue(s, a, excelize.Options{RawCellValue: true})
						var n float64
						if _, e := fmt.Sscan(raw, &n); e == nil && n > 0 && n < 100000 {
							props, _ := b.GetWorkbookProps()
							is1904 := props.Date1904 != nil && *props.Date1904
							if dt, e := excelize.ExcelDateToTime(n, is1904); e == nil {
								row.Existing[h] = date(dt)
							}
						}
					}
				}
			}
			if v != "" || f != "" {
				continue
			}
			merged := false
			for _, m := range merges {
				x, y, _ := excelize.CellNameToCoordinates(m.GetStartAxis())
				xx, yy, _ := excelize.CellNameToCoordinates(m.GetEndAxis())
				if j+1 >= x && j+1 <= xx && i >= y && i <= yy {
					merged = true
					break
				}
			}
			if !merged {
				row.Blanks[h] = a
			}
		}
		if summary {
			break
		}
		if len(row.Blanks) > 0 {
			p.Rows = append(p.Rows, row)
		}
	}
	return p, nil
}
func sample(now time.Time) (*excelize.File, error) {
	b := excelize.NewFile()
	s := "生产日报"
	if e := b.SetSheetName("Sheet1", s); e != nil {
		return nil, e
	}
	headers := []string{"日期", "工厂名称", "计划产量_箱", "实际产量_箱", "合格产量_箱", "损耗箱数", "达成率", "良品率"}
	if e := b.MergeCell(s, "A1", "H1"); e != nil {
		return nil, e
	}
	b.SetCellValue(s, "A1", "元气森林  /  工厂生产日报")
	b.SetRowHeight(s, 1, 34)
	for j, h := range headers {
		b.SetCellValue(s, cell(j+1, 3), h)
	}
	for i := 0; i < 14; i++ {
		b.SetCellValue(s, cell(1, i+4), date(now.AddDate(0, 0, i-13)))
		b.SetRowHeight(s, i+4, 23)
	}
	b.SetCellValue(s, "A18", "合计")
	for _, c := range []string{"C", "D", "E", "F"} {
		b.SetCellFormula(s, c+"18", "SUM("+c+"4:"+c+"17)")
	}
	style, e := b.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Color: "FFFFFF"}, Fill: excelize.Fill{Type: "pattern", Color: []string{"175B49"}, Pattern: 1}})
	if e != nil {
		return nil, e
	}
	b.SetCellStyle(s, "A1", "H1", style)
	b.SetCellStyle(s, "A3", "H3", style)
	b.SetCellStyle(s, "A18", "H18", style)
	pct, e := b.NewStyle(&excelize.Style{NumFmt: 10})
	if e != nil {
		return nil, e
	}
	b.SetCellStyle(s, "G4", "H17", pct)
	b.SetColWidth(s, "A", "H", 19)
	return b, nil
}
