# SCM

调拨、生产订单履约与领料协调。

## 独立启动

在本目录执行（Go 1.24+）：

```sh
go mod download
go run .
```

- 健康检查：http://127.0.0.1:8003/health
- Swagger：http://127.0.0.1:8003/swagger
- OpenAPI：http://127.0.0.1:8003/api.json

采用 GoFrame v2.10.2，拥有独立 go.mod、配置和进程。
当前仅提供可运行骨架；业务接口、数据库和跨系统通信尚未实现。
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
