package tablehandle

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/xuri/excelize/v2"
	"io"
	"multi-system/yq-tablehandle-system/backend/app/tablehandle/internal/logic/report"
	"multi-system/yq-tablehandle-system/backend/utility/idgen"

	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var validID = regexp.MustCompile("^[a-f0-9]{32}$")
var openOptions = excelize.Options{UnzipSizeLimit: 32 << 20, UnzipXMLSizeLimit: 4 << 20}

func writeJSON(c *ghttp.Request, status int, value any) {
	c.Response.WriteHeader(status)
	c.Response.WriteJson(value)
}
func Bind(server *ghttp.Server, dir string) {
	r := server.Group("/api")
	bindAdmin(r, conversionController{}, scriptController{})
	r.GET("/health", func(c *ghttp.Request) { writeJSON(c, 200, map[string]any{"status": "ok"}) })
	r.GET("/options", func(c *ghttp.Request) {

		mode := "本地演示"
		if os.Getenv("DEEPSEEK_API_KEY") != "" {
			mode = "DeepSeek"
		}
		writeJSON(c, 200, report.Options(mode))
	})
	r.GET("/sample", func(c *ghttp.Request) {
		b, e := report.Sample()
		if e != nil {
			writeJSON(c, 500, map[string]any{"error": e.Error()})
			return
		}
		defer b.Close()
		buf, e := b.WriteToBuffer()
		if e != nil {
			writeJSON(c, 500, map[string]any{"error": e.Error()})
			return
		}
		c.Response.Header().Set("Content-Disposition", "attachment; filename=factory-template.xlsx")
		c.Response.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Response.Write(buf.Bytes())
	})
	r.POST("/uploads", func(c *ghttp.Request) {
		c.Request.Body = http.MaxBytesReader(c.Response.RawWriter(), c.Request.Body, 9<<20)
		source, file, e := c.Request.FormFile("file")
		if e == nil {
			defer source.Close()
		}
		if e != nil || !strings.EqualFold(filepath.Ext(file.Filename), ".xlsx") || file.Size > 8<<20 {
			writeJSON(c, 400, map[string]any{"error": "请选择不超过 8 MB 的 .xlsx 文件"})
			return
		}
		id := idgen.Hex128()
		p := filepath.Join(dir, "up-"+id+".xlsx")
		if e = saveUpload(p, source); e != nil {
			writeJSON(c, 500, map[string]any{"error": "上传失败"})
			return
		}
		b, e := excelize.OpenFile(p, openOptions)
		if e != nil {
			os.Remove(p)
			writeJSON(c, 400, map[string]any{"error": "无法读取工作簿，请检查文件格式"})
			return
		}
		defer b.Close()
		writeJSON(c, 201, map[string]any{"id": id, "sheets": b.GetSheetList()})
	})
	r.POST("/run", func(c *ghttp.Request) {
		var in struct {
			ID, Sheet string
			Filters   report.Filters
		}
		c.Request.Body = http.MaxBytesReader(c.Response.RawWriter(), c.Request.Body, 8192)
		if json.NewDecoder(c.Body).Decode(&in) != nil || !validID.MatchString(in.ID) {
			writeJSON(c, 400, map[string]any{"error": "请求无效"})
			return
		}
		if e := in.Filters.Validate(); e != nil {
			writeJSON(c, 400, map[string]any{"error": e.Error()})
			return
		}
		b, e := excelize.OpenFile(filepath.Join(dir, "up-"+in.ID+".xlsx"), openOptions)
		if e != nil {
			writeJSON(c, 404, map[string]any{"error": "模板已失效，请重新上传"})
			return
		}
		defer b.Close()
		c.Response.Header().Set("Content-Type", "text/event-stream")
		c.Response.Header().Set("Cache-Control", "no-cache")
		c.Response.Header().Set("X-Accel-Buffering", "no")
		send := func(kind string, v any) {
			raw, _ := json.Marshal(v)
			fmt.Fprintf(c.Response.BufferWriter, "event: %s\ndata: %s\n\n", kind, raw)
			c.Response.Flush()
		}
		ctx, cancel := context.WithTimeout(c.Context(), 4*time.Minute)
		defer cancel()
		send("start", map[string]any{"message": "已接收填报任务"})
		result, e := report.Run(ctx, b, in.Sheet, idgen.Hex128(), dir, in.Filters, func(tool, msg string) { send("step", map[string]any{"tool": tool, "message": msg}) })
		if e != nil {
			send("failure", map[string]any{"message": e.Error()})
			return
		}
		send("result", result)
	})
	r.GET("/results/:id", func(c *ghttp.Request) {
		id := c.GetRouter("id").String()
		if !validID.MatchString(id) {
			c.Response.WriteStatus(404)
			return
		}
		p := filepath.Join(dir, id+".xlsx")
		if _, e := os.Stat(p); e != nil {
			c.Response.WriteStatus(404)
			return
		}
		c.Response.ServeFileDownload(p, "元气森林工厂填报结果.xlsx")
	})
}

func saveUpload(path string, source io.Reader) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	_, err = io.Copy(file, source)
	closeErr := file.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		os.Remove(path)
	}
	return err
}
