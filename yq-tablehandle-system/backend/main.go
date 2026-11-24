package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/xuri/excelize/v2"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func token() string {
	var b [16]byte
	if _, e := rand.Read(b[:]); e != nil {
		panic(e)
	}
	return hex.EncodeToString(b[:])
}

var validID = regexp.MustCompile("^[a-f0-9]{32}$")
var openOptions = excelize.Options{UnzipSizeLimit: 32 << 20, UnzipXMLSizeLimit: 4 << 20}

func router(dir string) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/api/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.GET("/api/options", func(c *gin.Context) {
		now := today()
		mode := "本地演示"
		if os.Getenv("DEEPSEEK_API_KEY") != "" {
			mode = "DeepSeek"
		}
		c.JSON(200, gin.H{"factories": factories, "lines": lines, "skus": skus, "start": date(now.AddDate(0, 0, -13)), "end": date(now), "mode": mode})
	})
	r.GET("/api/sample", func(c *gin.Context) {
		b, e := sample(today())
		if e != nil {
			c.JSON(500, gin.H{"error": e.Error()})
			return
		}
		defer b.Close()
		buf, e := b.WriteToBuffer()
		if e != nil {
			c.JSON(500, gin.H{"error": e.Error()})
			return
		}
		c.Header("Content-Disposition", "attachment; filename=factory-template.xlsx")
		c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
	})
	r.POST("/api/uploads", func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 9<<20)
		file, e := c.FormFile("file")
		if e != nil || !strings.EqualFold(filepath.Ext(file.Filename), ".xlsx") || file.Size > 8<<20 {
			c.JSON(400, gin.H{"error": "请选择不超过 8 MB 的 .xlsx 文件"})
			return
		}
		id := token()
		p := filepath.Join(dir, "up-"+id+".xlsx")
		if e = c.SaveUploadedFile(file, p); e != nil {
			c.JSON(500, gin.H{"error": "上传失败"})
			return
		}
		b, e := excelize.OpenFile(p, openOptions)
		if e != nil {
			os.Remove(p)
			c.JSON(400, gin.H{"error": "无法读取工作簿，请检查文件格式"})
			return
		}
		defer b.Close()
		c.JSON(201, gin.H{"id": id, "sheets": b.GetSheetList()})
	})
	r.POST("/api/run", func(c *gin.Context) {
		var in struct {
			ID, Sheet string
			Filters   Filters
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 8192)
		if c.ShouldBindJSON(&in) != nil || !validID.MatchString(in.ID) {
			c.JSON(400, gin.H{"error": "请求无效"})
			return
		}
		if e := in.Filters.validate(); e != nil {
			c.JSON(400, gin.H{"error": e.Error()})
			return
		}
		b, e := excelize.OpenFile(filepath.Join(dir, "up-"+in.ID+".xlsx"), openOptions)
		if e != nil {
			c.JSON(404, gin.H{"error": "模板已失效，请重新上传"})
			return
		}
		defer b.Close()
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("X-Accel-Buffering", "no")
		send := func(kind string, v any) {
			raw, _ := json.Marshal(v)
			fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", kind, raw)
			c.Writer.Flush()
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 4*time.Minute)
		defer cancel()
		j := Job{book: b, sheet: in.Sheet, id: token(), dir: dir, filters: in.Filters, db: seed(today()), trace: func(tool, msg string) { send("step", gin.H{"tool": tool, "message": msg}) }}
		send("start", gin.H{"message": "已接收填报任务"})
		if e = j.run(ctx); e != nil {
			send("failure", gin.H{"message": e.Error()})
			return
		}
		send("result", j.report)
	})
	r.GET("/api/results/:id", func(c *gin.Context) {
		id := c.Param("id")
		if !validID.MatchString(id) {
			c.Status(404)
			return
		}
		p := filepath.Join(dir, id+".xlsx")
		if _, e := os.Stat(p); e != nil {
			c.Status(404)
			return
		}
		c.FileAttachment(p, "元气森林工厂填报结果.xlsx")
	})
	return r
}
func main() {
	_ = godotenv.Load("../.env")
	dir, e := os.MkdirTemp("", "factory-reports-")
	if e != nil {
		log.Fatal(e)
	}
	defer os.RemoveAll(dir)
	log.Println("工厂填报服务 http://127.0.0.1:8191")
	if e = router(dir).Run("127.0.0.1:8191"); e != nil {
		log.Print(e)
	}
}
