package main

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

func main() {
	server := g.Server()
	server.Group("/api", func(group *ghttp.RouterGroup) {
		group.GET("/health", func(request *ghttp.Request) {
			request.Response.WriteJson(g.Map{"status": "ok"})
		})
	})
	server.Run()
}
