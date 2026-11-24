package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/xuri/excelize/v2"
)

func fixture(t *testing.T) *Job {
	t.Helper()
	now := time.Date(2026, 11, 25, 0, 0, 0, 0, time.UTC)
	b, e := sample(now)
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { b.Close() })
	return &Job{book: b, sheet: "生产日报", id: "result", dir: t.TempDir(), db: seed(now), filters: Filters{Factory: factories[0], Start: "2026-11-12", End: "2026-11-25"}, trace: func(string, string) {}}
}
func TestReportPreservesTemplate(t *testing.T) {
	t.Setenv("DEEPSEEK_API_KEY", "")
	j := fixture(t)
	j.filters.Note = "只填实际产量和良品率，汇总行保留公式"
	j.filters.Start = "2026-11-11"
	j.filters.End = "2026-11-12"
	style, _ := j.book.GetCellStyle(j.sheet, "H4")
	if e := j.run(t.Context()); e != nil {
		t.Fatal(e)
	}
	if j.report.Rows != 1 || len(j.report.Warnings) == 0 {
		t.Fatal(j.report)
	}
	out, e := excelize.OpenFile(filepath.Join(j.dir, "result.xlsx"))
	if e != nil {
		t.Fatal(e)
	}
	defer out.Close()
	for a, expected := range map[string]string{"C4": "", "D4": "6459", "D5": "", "A4": "2026-11-12", "E4": ""} {
		v, e := out.GetCellValue(j.sheet, a)
		if e != nil || v != expected {
			t.Fatalf("%s: %q", a, v)
		}
	}
	raw, _ := out.GetCellValue(j.sheet, "H4", excelize.Options{RawCellValue: true})
	ratio, _ := strconv.ParseFloat(raw, 64)
	if math.Abs(ratio-6432.0/6459) > 1e-9 {
		t.Fatal(ratio)
	}
	if v, _ := out.GetCellStyle(j.sheet, "H4"); v != style {
		t.Fatal("percentage style changed")
	}
	if f, _ := out.GetCellFormula(j.sheet, "D18"); f != "SUM(D4:D17)" {
		t.Fatal(f)
	}
	if m, _ := out.GetMergeCells(j.sheet); len(m) != 1 || m[0].GetEndAxis() != "H1" {
		t.Fatal(m)
	}
}
func TestNoDataAndSQL(t *testing.T) {
	t.Setenv("DEEPSEEK_API_KEY", "")
	j := fixture(t)
	j.filters.Start = "2020-01-01"
	j.filters.End = j.filters.Start
	if e := j.run(t.Context()); e != nil || j.report.Rows != 0 {
		t.Fatal(e, j.report)
	}
	j = fixture(t)
	j.phase = 2
	if _, e := j.query(t.Context(), &Query{Table: production, SQL: "DELETE FROM reports"}); e == nil {
		t.Fatal("SQL accepted")
	}
}
func TestDeepSeekToolLoop(t *testing.T) {
	j := fixture(t)
	calls := 0
	names := []string{"parse_excel_template", "list_data_sources", "query_data_warehouse", "fill_excel_template"}
	args := []any{Empty{}, Empty{}, Query{Table: production, Mapping: map[string]string{"日期": day, "工厂名称": plant, "实际产量_箱": "实际产量"}}, Fill{Sheet: j.sheet}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Tools []any
			Model string
		}
		if e := json.NewDecoder(r.Body).Decode(&request); e != nil || len(request.Tools) != 4 || request.Model != "deepseek-chat" {
			t.Errorf("invalid tool request: %v", e)
		}
		if calls >= len(names) {
			http.Error(w, "extra call", 400)
			return
		}
		arg, _ := json.Marshal(args[calls])
		name := names[calls]
		calls++
		call := map[string]any{"type": "function", "id": fmt.Sprint(calls), "function": map[string]string{"name": name, "arguments": string(arg)}}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"id": "demo-completion", "model": "deepseek-chat", "object": "chat.completion", "choices": []any{map[string]any{"finish_reason": "tool_calls", "message": map[string]any{"role": "assistant", "tool_calls": []any{call}}}}})
	}))
	defer server.Close()
	t.Setenv("DEEPSEEK_API_KEY", "test-fixture")
	t.Setenv("DEEPSEEK_BASE_URL", server.URL)
	t.Setenv("DEEPSEEK_MODEL", "")
	if e := j.run(t.Context()); e != nil {
		t.Fatal(e)
	}
	if calls != 4 || j.report.Rows != 14 {
		t.Fatal(calls, j.report)
	}
}

func TestOtherSources(t *testing.T) {
	t.Setenv("DEEPSEEK_API_KEY", "")
	for _, tc := range []struct{ table, header, value string }{{equipment, "运行时长_小时", "35"}, {inbound, "入库箱数", "6303"}} {
		t.Run(tc.table, func(t *testing.T) {
			j := fixture(t)
			b := excelize.NewFile()
			defer b.Close()
			j.book = b
			j.sheet = "Sheet1"
			for a, v := range map[string]string{"A1": "日期", "B1": tc.header, "A2": j.filters.Start} {
				if err := b.SetCellValue(j.sheet, a, v); err != nil {
					t.Fatal(err)
				}
			}
			j.filters.End = j.filters.Start
			if err := j.run(t.Context()); err != nil {
				t.Fatal(err)
			}
			if j.report.Source != tc.table || j.report.Rows != 1 || j.report.Preview[0][tc.header] != tc.value {
				t.Fatal(j.report)
			}
		})
	}
}

func TestFructoseWorkbook(t *testing.T) {
	t.Setenv("DEEPSEEK_API_KEY", "")
	j := fixture(t)
	b, e := excelize.OpenFile("testdata/fructose-demo.xlsx")
	if e != nil {
		t.Fatal(e)
	}
	defer b.Close()
	j.book = b
	j.sheet = b.GetSheetList()[0]
	if e = j.run(t.Context()); e != nil {
		t.Fatal(e)
	}
	if j.report.Rows != 14 || j.report.Source != fructose || len(j.layout.Rows) != 17 {
		t.Fatal(j.report)
	}
	out, e := excelize.OpenFile(filepath.Join(j.dir, "result.xlsx"))
	if e != nil {
		t.Fatal(e)
	}
	defer out.Close()
	for a, want := range map[string]string{"A7": j.filters.Start, "I20": j.filters.End, "B7": "10.5", "P7": "25.1", "A21": "", "J4": "线别：L01\n品项：白桃味气泡水", "A26": "经理："} {
		got, e := out.GetCellValue(j.sheet, a)
		if e != nil || got != want {
			t.Fatalf("%s = %q, want %q", a, got, want)
		}
	}
}
