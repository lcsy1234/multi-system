# template-gf 项目事实

本文件只写**本仓库的事实与差异**；做法与条款在 [`README.md`](README.md) 的专题里，不在此重复。冲突时以本文件与专题为准，判断不出优先级先问人。

## 技术栈

| 项 | 事实 |
|---|---|
| module | `codeup.aliyun.com/yuanqisenlin/template-gf` |
| 语言 | Go，`go.mod` 声明 `go 1.23.0` |
| 框架 | GoFrame v2.10.3（HTTP + gRPC） |
| 数据库 | PostgreSQL（`contrib/drivers/pgsql`，`lib/pq`） |
| 缓存 | Redis（`contrib/nosql/redis`；asynq 另有独立连接块） |
| 迁移 | goose v3，SQL 内嵌进二进制 |
| 异步 | Asynq v0.25 + asynqmon（监控页 `/admin/asynq`） |
| 消息 | Kafka（`segmentio/kafka-go`，封装在 `utility/kafka`） |
| 配置 | `manifest/config` 文件适配器；可选 Nacos 适配器（整份替换）；环境变量按大写路径覆盖 |
| 接口契约 | HTTP 用 `g.Meta`；gRPC 用 `manifest/protobuf` + `gf gen pb` |
| 飞书 | 自定义机器人 webhook + 签名（`utility/feishu`），不接飞书审批 |

## 进程

四个入口（`boot/<role>/main.go`）、四个角色：

| 子命令 | 职责 | 默认端口 |
|---|---|---|
| `http` | HTTP API、`/healthz`、`/readyz`、asynq 监控 | `:8080` |
| `grpc` | gRPC API | `:9090` |
| `worker` | Asynq 消费 | — |
| `migrate` | goose 迁移（`-action=up/down/status/...`） | — |

## 本仓库约定

| 项 | 约定 |
|---|---|
| 应用目录 | `app/template`，`boot/<role>/main.go` 四个发布入口，装配共用 `internal/cmd` |
| 进程 | `http`(`:8080`) / `grpc`(`:9090`) / `worker` / `migrate` |
| 数据库 | PostgreSQL；时间 `TIMESTAMPTZ`；金额见「业务域」 |
| 迁移 | `YYYYMMDDHHMMSS_<表名>.sql`，`make migrate-*`，只增不改 |
| 分层 | api → controller → service → logic → dao |
| 生成器 | `make dao/ctrl/service/proto`；`make dao` 生成后自动补齐 `baseDao` |
| 测试 | 与被测包同目录；集成用例加 `//go:build integration` |
| 错误码 | gcode 内置码 + 六位业务码；业务码文案由中间件原样透传，见 [02 §7.3](02-api-validation.md) |
| 文案 | 未接 i18n，直接中文 |

## 目录事实

```
app/template/
  boot/<role>/main.go      四个发布入口（http / grpc / worker / migrate）
  api/<模块>/{<模块>.go, v1/}   生成接口 + Req/Res + pb
  internal/cmd/            http / grpc / worker / migrate 子命令
  internal/controller/     控制器（HTTP 的 ControllerV1、gRPC 的 Controller）
  internal/service/        生成的 service 接口与注册器
  internal/logic/          业务实现 + logic.go 注册
  internal/dao/            baseDao（手写）+ 生成表 DAO + internal/
  internal/model/          Input/Output + do + entity
  internal/consts/         队列名、任务类型、枚举与编码前缀
  internal/task/           Asynq 生产者/Server/handler + kafka
  manifest/config/         config.example.yaml
  manifest/migration/      goose SQL + embed.go
  manifest/protobuf/       proto 契约
  manifest/deploy/         docker compose（postgres / redis / kafka）
utility/                   跨进程能力：config database connector filter lock kafka feishu middleware requestid infra
hack/                      make 片段（hack.mk、hack-cli.mk）、dao_shape.py（make dao 内部使用）
```

## 生成器

| 命令 | 作用 | 前置 |
|---|---|---|
| `make dao` | 由库表生成 `dao/`、`model/do`、`model/entity` | 迁移已执行；`app/template/hack/config.yaml` 的 `gen.dao.tables` 已登记 |
| `make dao.overwrite` | 重刷 `dao/` 索引文件（丢弃手改） | 确认没有自定义方法 |
| `make ctrl` | 由 `api/` 生成 controller 骨架 | api Req/Res 已写好 |
| `make service` | 由 `logic/` 生成 service 接口 | 新 logic 包已注册 |
| `make proto` | 由 `manifest/protobuf` 生成 pb 与 gRPC 骨架 | 已装 `protoc`、`protoc-gen-go`、`protoc-gen-go-grpc`、`gf` CLI |

- 生成物清单、生成顺序与「不手改」的边界见 [10 生成器顺序](10-goframe-layering.md#生成器顺序)。
- `make` 目标多数依赖 `cli.install`（会下载 gf CLI）；离线环境用直接命令替代，并在报告里写明。

## 状态与常量

- 状态、类型、队列名、任务类型、单号前缀一律进 `internal/consts`；禁止在 logic 里写裸数字与裸字符串。
- 命名：状态 `XxxStatus` + 取值，队列 `QueueXxx`，任务类型 `TaskTypeXxx`，编码前缀 `XxxSeqPrefix`。
- 同一前缀只定义一次，调用处只引用常量。
- **状态 / 类型从 `1` 起连续编号，禁止 `0`**；库里用 `CHECK (<列> <> 0)` 兜底，常量由 `TestStatusConstantsStartAtOne` 守住（见 [03](03-db-migration.md#状态与类型)）。

## 业务域

- 模板自带 `demo` 域（`demo_item`）演示 api → controller → service → logic → dao 全链路；报销、主数据、登录等业务域已移除，新增时按 [10](10-goframe-layering.md) 的分层与生成顺序落地。
- 金额用 `NUMERIC(18,2)` + `decimal.Decimal`；单号/编码前缀进 `consts`；状态与类型常量从 `1` 起，不在迁移或代码里写魔数。

## 外部集成现状

| 对端 | 方式 | 现状 | 注意 |
|---|---|---|---|
| PostgreSQL | `utility/database` + gdb | 可用 | 本机或外部 PostgreSQL |
| Redis / Asynq | `app/template/internal/task` + `asynq.redis` | 可用 | 队列权重在 `asynq.queues` |
| Kafka | `utility/kafka` | 可用 | topic / group 走配置 |
| Nacos | `utility/config` 适配器 | 可用（默认关） | 启用时**整份替换**本地配置 |
| 飞书自定义机器人 | `utility/feishu` | 可用 | webhook 为空时记 Warn 跳过 |
| 外部 HTTP 对端 | `utility/connector/httpx` | 可用（配置驱动） | 地址在 `http.services` |
| 外部 gRPC 对端 | `utility/connector/grpcx` | 可用（懒连接） | 地址在 `grpc.clients` |
| 飞书审批 / 企业微信 | — | **无** | 不要假设有回调能力 |
| 对象存储（OSS 等） | — | **无客户端** | 附件存储方案未定，先确认再实现 |
| 银行 / 税务 / 财务系统 | — | **无** | 相关需求先确认接口 |

## 未落地项（TBD，禁止臆造）

- **登录与身份体系**：无 SSO、无账号密码、无鉴权中间件；HTTP 链只有 `RequestID` + `HandlerResponse`。
- **接口级 RBAC 与数据范围框架**。
- **401/403/404 内置码**：`61 → 401`、`65 → 404` 已由 `HandlerResponse` 映射；`403` 仍只靠框架/网关透传，路由不存在时仍是 ghttp 默认 404 页（不套信封）。见 [02 §7.2](02-api-validation.md)。
- **i18n**：`manifest/i18n` 是空目录，`v` 标签与错误文案目前是硬编码。
- **统一幂等键（Idempotency-Key）中间件**。
- **Outbox / 可靠事件投递**。
- **CI 与 `golangci-lint` 配置**：仓库暂无流水线，也没有 `golangci-lint` 配置。

涉及以上：Spec 列为未决项，先做不依赖它的部分，或停下来确认；不编造接口、表结构、配置项。

## 验证命令

从仓库根执行：

```bash
make check          # 角色构建 + build/vet（排除 boot）+ 单元守卫（internal/guard + dao 形状锁）
make test           # go test（排除 boot）
gofmt -l .          # 只对本次改动文件要求空输出
```

`boot/` 下是四个目录入口（`boot/<role>/main.go`），每个目录一个 `package main`；
`go build ./...` / `go vet ./...` / `go test ./...` 可直接运行；
Makefile 已统一排除 `./boot`，直接跑标准命令时请自行加
`$(go list ./... | grep -v /boot)`。

Make 等价目标：`make build`（`go build` 四个角色到 `bin/template-<role>`）、`make test`、`make check`、`make vet`、`make fmt`。

集成用例需要真实 PostgreSQL（`go test -tags=integration ./...`）；本机可用已有 PostgreSQL 实例，跑之前先确认 `5432` 可达。

**工具链**：`go.mod` 要求 1.23.0。本机 Go 低于该版本时会尝试下载工具链并写 `~/go`、`~/Library/Caches`；受限环境（沙箱、无写权限 HOME）会失败，把缓存指到仓库内即可：

```bash
GOPATH=$PWD/.gopath GOMODCACHE=$PWD/.gopath/pkg/mod \
GOCACHE=$PWD/.gocache GOTOOLCHAIN=go1.23.0 go build ./...
```

`.gopath/`、`.gocache/` 不应提交。

## 已知验证限制

- 仓库没有 CI 流水线，也没有 `golangci-lint` 配置；静态检查以 `go vet` 为准。
- 集成测试需要真实 PostgreSQL（本机或外部实例）；没有可用实例时相关用例标「未执行」，不写「通过」。
- 迁移是否可执行、可回滚，只能在有数据库时验证；没跑过就写「未执行」，不写「通过」。
- `make dao/ctrl/service/proto` 依赖 gf CLI 与 protoc 插件，离线环境跑不了，标「未执行」并写明缺什么。
