package cmd

import (
	"context"
	"os"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"
)

var Socket = gcmd.Command{
	Name:  "socket",
	Usage: "mes-socket",
	Brief: "启动 MES WebSocket 演示服务",
	Func: func(ctx context.Context, parser *gcmd.Parser) error {
		address := os.Getenv("SOCKET_ADDR")
		if address == "" {
			address = g.Cfg().MustGet(ctx, "socket.address", "127.0.0.1:8005").String()
		}

		server := g.Server("mes-socket")
		server.SetAddr(address)
		server.BindHandler("GET:/health", func(r *ghttp.Request) {
			r.Response.WriteJson(g.Map{"service": "mes-socket", "status": "ok"})
		})
		server.BindHandler("GET:/ws", func(r *ghttp.Request) {
			connection, err := r.WebSocket()
			if err != nil {
				return
			}
			defer connection.Close()
			connection.SetReadLimit(64 * 1024)

			// 最小演示：将收到的消息原样回传，不处理领料任务。
			for {
				messageType, message, err := connection.ReadMessage()
				if err != nil {
					return
				}
				if err := connection.WriteMessage(messageType, message); err != nil {
					return
				}
			}
		})
		server.Run()
		return nil
	},
}
