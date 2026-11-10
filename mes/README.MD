# MES

生产执行与领料申请。

## 独立启动

在本目录执行（Go 1.24+）：

```sh
go mod download
go run . migrate  # 先创建 MySQL Demo 库与表
go run .
```

主进程也可以通过 `go run ./boot/main.go` 启动。两种方式都只提供页面与 HTTP 接口，不启动异步任务消费。

## 三个独立进程入口

```text
boot/
  main.go      # cmd.Main：MES 页面与 HTTP 接口，8004
  socket.go    # cmd.Socket：WebSocket 消息回显示例，8005
  listener.go  # cmd.Listener：Redis/Asynq 任务消费
```

在 MES 目录分别打开终端，按需启动：

```sh
REDIS_ADDR=127.0.0.1:6380 go run ./boot/main.go
go run ./boot/socket.go
REDIS_ADDR=127.0.0.1:6380 go run ./boot/listener.go
```

主进程和 listener 必须连接同一个 Redis；6380 是 `make demo` 使用的 Redis。
socket 的地址由 `socket.address` 配置，支持 `SOCKET_ADDR` 覆盖；消息入口为 `/ws`，收到什么消息就回传什么消息。
三个进程独立启动、独立停止，`make demo` 不会启动 socket 或 listener。

`socket.go` 和 `listener.go` 使用 `//go:build ignore`，避免默认构建时同目录多个 `main()` 冲突；上面的按文件运行及 Makefile 的按文件编译仍会正常构建它们。

- 健康检查：http://127.0.0.1:8004/health
- Swagger：http://127.0.0.1:8004/swagger
- OpenAPI：http://127.0.0.1:8004/api.json

采用 GoFrame v2.10.2，拥有独立 go.mod、配置和进程。
当前只演示保存领料申请、Asynq 入队、worker 读取申请并打印日志，不调用 SCM。
使用与接口见上级 `docs/async-material-demo.md`。
领料链路与后续阶段见上级目录 `docs/material-issue-flow.md`。

## 目录约定

- `api/`：版本化请求、响应和接口声明。
- `internal/controller/`：HTTP 入口。
- `internal/logic/`：业务逻辑。
- `internal/service/`：需要时增加业务接口。
- `internal/client/`：调用其他系统的客户端。
- `internal/model/`、`internal/dao/`：数据模型与数据库访问。
- `internal/cmd/`：服务启动及路由绑定。
- `boot/main.go`：MES 主进程入口，调用 `cmd.Main`，与根目录 `main.go` 等效。
- `boot/socket.go`：独立 WebSocket 进程入口，调用 `cmd.Socket`。
- `boot/listener.go`：独立消费进程入口，调用 `cmd.Listener`。
- `utility/worker/producer.go`、`consumer.go`：Asynq 生产和消费封装。
- `manifest/config/`：本地开发配置，不存放真实凭证。

前端入口：http://127.0.0.1:8004/ 。另开终端运行 `go run ./boot/listener.go` 消费任务。根目录 `make demo` 仅启动 HTTP 与 Redis，另开终端 `make listener` 才开始消费。

使用 Demo Redis 6380 时，在本目录运行 `REDIS_ADDR=127.0.0.1:6380 go run ./boot/listener.go`。
消费进程仅通过上述独立入口或根目录 `make listener` 手动启动。
根目录 `make listener` 使用独立二进制 `.bin/mes-listener`，`make worker` 保留为兼容别名。

前端已迁移至根目录 `mes-web/`（React + TypeScript + Ant Design）。生产页面需先在根目录执行 `make frontend-build`；开发时执行 `make frontend-dev`，访问 5173，支持热更新。后端 `internal/web/web.go` 只负责提供构建产物，不再嵌入前端源码。
