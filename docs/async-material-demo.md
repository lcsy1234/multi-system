# 第一步：领料任务入队与消费

当前只保留下面这条链路：

```text
MES 页面提交领料申请
  → MySQL 保存申请与明细
  → Asynq Client 直接将任务放进 Redis
  → 单独启动 MES worker
  → worker 根据申请 ID 查询 MySQL、打印日志
  → 页面显示“已消费”
```

不调用 SCM，不包含签名、outbox 发布器、故障注入和手动重投接口。SCM、WMS、ERP 保留原来的 GoFrame 骨架。

## 先看排队

需要 MySQL 已启动，以及 Go、Node.js 20.19+、Python 3、make、redis-server。前端源码在 `mes-web/`，使用 React + TypeScript + Ant Design。

```sh
make demo
```

打开 http://127.0.0.1:8004/ ，在“生产管理 → 生产产程”点击领料并提交。此时没有启动 worker，申请应保持“已排队”。

`make demo` 创建 MES 的独立 Demo 库和初始表，构建 React 前端，并启动专用 Redis 6380 和四个 HTTP 服务，**不会自动启动 worker**。

## 再看消费

保持第一个终端运行，另开终端执行：

```sh
make worker
```

终端会打印：

```text
[领料任务消费] applicationId=... number=... process=... items=...
```

点击页面“刷新数据”，状态变为“已消费”；也可以打开“自动刷新（5 秒）”开关。Ctrl+C 可单独停止 worker，再提交一笔申请，就能再次观察排队。

`make worker` 默认使用 Demo Redis 6380。使用已有 Redis 时需要两边指向同一个实例：

```sh
# 例如自有 Redis 6379，HTTP 与消费端都显式配置
REDIS_ADDR=127.0.0.1:6379 make dev
REDIS_ADDR=127.0.0.1:6379 make worker
```

## 按顺序读代码

1. `mes/internal/controller/material/material_v1_create.go`：创建领料申请的 HTTP 入口。
2. `mes/internal/logic/material/service.go` 的 `Create`：保存申请，构造 `TaskPayload`，调用 `Publisher.Push`。
   `mes/internal/taskqueue/publisher.go` 统一封装 JSON 序列化、任务创建、Redis 入队及重复任务处理；当前设置不自动重试，成功任务保留 24 小时。
3. `mes/internal/boot/boot.go` 的 `Worker`：注册队列和任务处理函数。
4. `mes/internal/logic/material/worker.go` 的 `HandleTask`：反序列化、查询申请、打印日志。

任务类型是 `mes:material:consume`，队列是 `mes-material`。任务内容只有：

```json
{"applicationId":"申请 UUID"}
```

申请明细保存在 MySQL，不复制到任务里。当前显式配置 `MaxRetry(0)`，先观察一次消费，不展开失败重试。返回 `nil` 表示任务成功，错误会进入 Asynq 的失败归档。

为观察消费进展保留 MySQL 状态和简短日志；这一步没有跨库、跨系统业务处理。相同请求 ID 用固定 TaskID 去重。

## 配置与验证

- Redis：`REDIS_ADDR`、`REDIS_PASSWORD`，配置默认值在 MES `manifest/config/config.yaml`。
- MySQL：`MES_MYSQL_DSN`，默认 `root@tcp(127.0.0.1:3306)/multi_system_mes_demo`。
- HTTP：MES 默认 8004，可用 `HTTP_ADDR` 单独覆盖。
- `make test`、`make vet`：编译、校验规则测试与静态检查。
- `make integration`：用临时 Redis 和随机 MySQL 测试库，验证不开 worker 时保持排队，启动后消费并打印日志。不会清理现有业务库。

## 本阶段的限制

MySQL 保存和 Redis 入队是两次独立写入；中间进程退出可能出现已保存但未入队。这是为了先看清入队和消费，有意暂不处理的可靠性问题。以后再单独学习 outbox、自动重试与 SCM 调用。
