# MES

生产执行与领料申请。

## 独立启动

在本目录执行（Go 1.24+）：

```sh
go mod download
go run . migrate  # 先创建 MySQL Demo 库与表
go run .
```

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
- `manifest/config/`：本地开发配置，不存放真实凭证。

前端入口：http://127.0.0.1:8004/ 。另开终端运行 `go run . worker` 消费任务。根目录 `make demo` 仅启动 HTTP 与 Redis，另开终端 `make worker` 才开始消费。

前端已迁移至根目录 `mes-web/`（React + TypeScript + Ant Design）。生产页面需先在根目录执行 `make frontend-build`；开发时执行 `make frontend-dev`，访问 5173，支持热更新。后端 `internal/web/web.go` 只负责提供构建产物，不再嵌入前端源码。
