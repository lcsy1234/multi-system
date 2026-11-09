# multi-system

四个独立 GoFrame v2.10.2 项目：ERP（8001）、WMS（8002）、SCM（8003）、MES（8004）。根目录通过 go.work 组织开发。

当前学习范围只包含：**MES 保存领料申请 → Redis/Asynq 入队 → worker 查询申请并打印日志**。SCM、WMS、ERP 保留项目骨架。

## 使用

需要 Go 1.24+、Node.js 20.19+（或 22.12+）、Python 3、make、redis-server 和已启动的 MySQL。

```sh
make demo
```

打开 http://127.0.0.1:8004/ ，在“生产管理 → 生产产程”提交领料申请。此时任务保持排队，**不会自动启动 worker**。

另开终端执行：

```sh
make worker
```

观察终端消费日志，以及页面“已排队 → 已消费”的变化。当前不调用 SCM，不包含 outbox、故障注入或手动重投。

- [入队与消费教程](docs/async-material-demo.md)
- [MySQL 存储说明](docs/database-proposal.md)
- [原始业务规划（后续阶段）](docs/material-issue-flow.md)

## 命令

```sh
make db-init       # 初始化 MES MySQL 库与表
make dev           # HTTP 服务，使用自有 Redis（默认 6379），不启动 worker
make worker        # 消费任务，默认使用 Demo Redis 6380
make frontend-dev  # React 开发服务 5173（支持热更新，代理到 MES 8004）
make frontend-build # TypeScript 检查与生产构建
make test
make vet
make integration   # 隔离的 MySQL/Redis 入队消费验收
make smoke         # 四服务健康检查；需先停止占用 8001–8004 的 Demo
```

Demo Redis 6380 的数据在 `.data/redis/`，HTTP 日志在 `.logs/`，worker 日志直接输出到终端。Ctrl+C 分别停止对应终端中的进程。

MES MySQL 由 `MES_MYSQL_DSN` 指定，默认 `root@tcp(127.0.0.1:3306)/multi_system_mes_demo`；不启动或停止已有 MySQL。数据库连接和表结构见存储说明。

每个服务均可在各自目录执行 `go run .`，并提供 `/health`、`/swagger`、`/api.json`。MES 前端已独立为 [mes-web/](mes-web/README.md)，采用 React + TypeScript + Vite + Ant Design。`make demo` 会先构建前端，MES 从 `mes-web/dist/` 提供生产页面；开发时可另启 `make frontend-dev`。不再将页面源文件内嵌到 Go 程序。

申请列表默认手动刷新，需要自动观察任务状态时，可打开“自动刷新（5 秒）”开关。
