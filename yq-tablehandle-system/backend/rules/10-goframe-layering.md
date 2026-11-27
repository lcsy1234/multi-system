---
paths:
  - "app/template/api/**/*.go"
  - "app/template/internal/**/*.go"
  - "utility/**/*.go"
---

# GoFrame 分层与生成器

## 四层调用链

```text
api/<模块>/v1（结构体 + g.Meta）
   -> controller（转交、组装）
   -> service（logic 生成的接口）
   -> logic（业务、事务、锁、投递）
   -> dao / model（数据访问）
```

- controller 只依赖 `service`，不直接依赖 `dao`；新代码禁止跨层调用。
- logic 不 import `ghttp`/`grpc`、不写响应；失败返回 error，由中间件或 gRPC 层统一转换。
- `utility/` 不得 import `app/template`；需要反向回调时用接口 + 注册。
- `model` 只放数据结构与 `Build` 这类组合逻辑，不依赖 `logic`/`service`。
- api 定义目录只放结构体与校验：可引用 `internal/model` 的数据结构，不引用 logic/dao/controller，不写业务判断。

## 目录归属

| 要加的东西 | 位置 | 细则归属 |
|---|---|---|
| 接口定义（HTTP） | `app/template/api/<模块>/v1/` | [02](02-api-validation.md) |
| proto 与 gRPC 契约 | `app/template/manifest/protobuf/<模块>/<版本>/` | [11](11-grpc-protobuf.md) |
| 控制器 | `app/template/internal/controller/<模块>/` | [02](02-api-validation.md) |
| service（生成） | `app/template/internal/service/` | 生成物，不手改、不放实现 |
| 业务实现（logic） | `app/template/internal/logic/<模块>/` | [04](04-transaction.md)、[05](05-remote-cache.md)、[06](06-async-task.md) |
| dao（生成 + 手写 base） | `app/template/internal/dao/` | [03](03-db-migration.md) |
| model / do / entity | `app/template/internal/model/` | [03](03-db-migration.md) |
| 枚举、编码前缀、队列与任务类型 | `app/template/internal/consts/` | [project.md](project.md)「状态与常量」 |
| 异步任务与生产注册 | `app/template/internal/task/` | [06](06-async-task.md) |
| 子命令装配 | `app/template/internal/cmd/`（http/grpc/worker/migrate） | 本文件 |
| 跨进程能力 | `utility/`（config、database、connector、filter、lock、kafka、feishu、middleware、requestid、infra） | [05](05-remote-cache.md) |
| 迁移 | `app/template/manifest/migration/` | [03](03-db-migration.md) |

## 生成器顺序

在 `app/template` 或仓库根执行（根 Makefile 已代理）：

1. goose 迁移建表（`make migrate-create name=xxx` → `make migrate-up`）→ 在 `app/template/hack/config.yaml` 的 `gen.dao.tables` 登记 → `make dao`；
2. `api/<模块>/v1` 写 Req/Res → `make ctrl` 生成 controller 骨架（`<模块>_new.go`、`<模块>_v1_<动作>.go`）；
3. `internal/logic/<模块>` 实现 → `make service` 生成/更新接口；
4. 在 `internal/cmd/http.go` 的 `/v1` 分组注册 `group.Bind(<模块>.New(...))`；gRPC 在 `internal/cmd/grpc.go` 注册；
5. 契约变更时 `make proto` 重新生成 pb 与 gRPC 控制器骨架。

- 生成物不手改：`dao/internal`、`model/entity`、`model/do`、`service`、`api/<模块>/<模块>.go`、`*.pb.go`。
- `dao/<表>.go` 只允许末尾追加自定义方法；`*_new.go` 生成一次后可改。
- **生成 DAO 只用 `make dao`**：它先 `gf gen dao`，再自动扫描 `internal/dao/internal` 下生成的表并补成嵌入 `baseDao` 的形状（跳过 `goose_*`；v2.10.3 不支持自定义 DAO 模板，只能生成后补）。
  - 绕过 `make dao` 直接跑 `gf gen dao` 会丢掉 `One/List/Page`，`internal/dao/shape_test.go` 会编译失败。
- 改表后按「迁移 → `app/template/hack/config.yaml` 登记 → `make dao`」顺序走；金额列依赖该文件里的 `typeMapping`（`NUMERIC` → `decimal.Decimal`），别在生成物上手工改类型。

## 模块注册

```go
type sDemo struct{}

// internal/logic/logic.go
func init() { service.RegisterDemo(New()) }

func New() *sDemo { return &sDemo{} }
```

- 结构体名 `sXxx`（`gf gen service` 按 `^s([A-Z]\w+)$` 收集，生成接口 `IXxx` 与访问器 `Xxx()`）。
- `init()` 只做注册；网络、DB、客户端初始化放 `utility/infra` 或子命令。
- 新增 logic 包或改导出方法签名/注释后必须 `make service`，不手写接口。

## 数据模型

- `do` 用于写入，`entity` 用于读取绑定；两者均生成，不手改，也只有它们带 `orm` 标签。
- `model/<模块>.go` 放 service 边界的 `XxxInput` / `XxxOutput` / `XxxFilter`；**结构体名一律以这三个后缀结尾**，不出现 `SubmitPlan`、`ExportRow` 这类中间态命名。
  - `Filter` 专指嵌进 `filter.PaginationInput` 的筛选条件（如 `DemoFilter`），本身就是入参的一部分，不必再叫 `XxxFilterInput`。
  - 不把 `entity` / `do` 塞进 `api`。
- **Input/Output/Filter 必须有 `json` 与 `dc`，且一律不带 `orm` 标签**：列名映射属于 `entity` / `do`，边界类型不带。
  - 带 `orm` 会让人以为改列名要同步改 Output，也会诱导把实体直接当接口返回。
  - 扫描依赖列名与字段名的 snake_case ⇄ CamelCase 约定，`internal/guard` 里有守卫。
- `filter` 标签只加在确实需要对外的筛选字段上——没有标签的字段不参与 WHERE，这是安全边界。

## 进程

四个目录入口（`boot/<role>/main.go`），共用 `internal/cmd` 装配：

子命令与端口是事实，见 [`project.md` 进程](project.md#进程)；这里只讲装配。

- 新增角色/子命令同时改 `internal/cmd/<role>.go` 与根 `Makefile` / `hack/hack.mk` 的运行目标。
- 禁止在 HTTP handler 里做全表扫描或长耗时批处理；批处理归 worker。

## 横切关注点

| 关注点 | 位置 | 细则 |
|---|---|---|
| 统一响应 | `utility/middleware.HandlerResponse` | [02 §6](02-api-validation.md) |
| 请求 ID | `utility/middleware.RequestID` + `utility/requestid` | [08](08-logging.md) |
| 日志 | `glog.*` | [08-logging.md](08-logging.md) |
| 配置 | `utility/config` + `manifest/config` + 可选 Nacos | [05](05-remote-cache.md) |
| 锁 | `utility/lock` | [05](05-remote-cache.md) |
| 查询过滤 | `utility/filter` | [03](03-db-migration.md) |
| 内置错误码 | `utility/errcode`（0/200/401/403/404/422/-1） | [02 §7.1](02-api-validation.md) |
| 业务错误码 | `internal/consts/error_code.go` | [02 §7.3](02-api-validation.md) |
| 枚举常量 | `internal/consts` | [project.md](project.md)「状态与常量」 |
