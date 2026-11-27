package cmd

import (
	"context"
	"os"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/joho/godotenv"
	"multi-system/yq-tablehandle-system/backend/app/tablehandle/internal/controller/tablehandle"
)

var Http = gcmd.Command{Name: "http", Func: func(ctx context.Context, parser *gcmd.Parser) error {
	_ = godotenv.Load("../.env")
	dir, err := os.MkdirTemp("", "factory-reports-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	server := g.Server("tablehandle")
	address := os.Getenv("TABLEHANDLE_HTTP_ADDR")
	if address == "" {
		address = "127.0.0.1:8191"
	}
	server.SetAddr(address)
	server.SetClientMaxBodySize(9 << 20)
	tablehandle.Bind(server, dir)
	server.Run()
	return nil
}}
