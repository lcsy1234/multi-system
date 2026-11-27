---
paths:
  - "**/*.go"
---

# 日志与告警

- 记录日志只用 `glog.*`；禁止 `g.Log()`（仅装配处可用来取实例）、`fmt.Print`、标准库 `log`。
- `utility/infra.Init` 启动时执行 `glog.SetDefaultLogger(g.Log())`，`glog.*` 才会继承 `logger` 配置；新进程入口必须经过 `infra.Init`，否则级别与输出配置不生效（glog 默认 `Debug`）。
- **所有手写业务方法的入口都必须打一条日志**：controller / logic / task / utility 中带 `ctx` 的手写方法，第一条语句用 `glog.Infof(ctx, "进入方法名", 业务键...)`。入口日志统一 `Info`，确保生产默认级别可见。生成代码（`dao/internal`、`entity`/`do`、service 接口）与纯 helper 除外。
- 入口日志只带方法名、业务键/关键参数（如 id、bizKey、queue），禁止完整请求体、密码、token；controller / logic 由 `TestBusinessMethodEntryLogs` 强制，task / utility 同口径暂由评审兜底。
- 同一个方法在出口再打一条完成/失败日志（defer 或最后 return 前）；入口与出口成对。长耗时允许「开始 + 完成」两条；事务内不打成功日志，提交后打。
- 优先结构化字段：`glog.Info(ctx, "bizNo", bizNo, "status", status)`；用 `Infof` 时不得把整个结构体 `%v` 进去。
- 失败带 `err` 对象而不是 `err.Error()`；`requestId` 由 `utility/requestid` 与中间件自动带上，不手写。

## 字段

- 必带业务键（单号、任务号、账号、工厂）；状态变更带 `from` / `to` / `operator`。

| 场景 | 必带字段 |
|---|---|
| 业务动作 | 业务键、`duration_ms` |
| 状态变更 | 业务键、`from`、`to`、`operator` |
| 失败 | 业务键、`err` |
| 批处理 / 循环 | `total`、`success`、`failed`（循环外一条） |
| 外部调用 | `target`、`timeout_ms`、`elapsed_ms`、`err` |
| 入参 | 业务键、关键状态/类型、条数（如 `itemCount`） |

## 级别

- `Debug` 本地与排障（生产默认 `Info`）；`Info` 业务里程碑（创建、提交、审批、支付），一次业务一条；`Warning` 可恢复降级（重试后成功、下游短暂不可用）。
- `Error` 只用于三类：吞错继续执行、异步/定时/后台 goroutine（无 HTTP 出口）、补偿与对账失败；业务失败返回 error 即可，不自己打 Error。
- `middleware.HandlerResponse` 已对非 client-error 码记 Error 并脱敏，不要在 controller 里重复打印同一个错误；gRPC 侧在 controller 里记一条即可。

## 脱敏

- 禁止打印密码、token、身份证、手机号、完整银行卡号与完整请求体。
- 确需明细排障时：降到 `Debug`，只打白名单字段并截断（单字段 ≤ 512 字符），不把整个结构体直接 `%+v` 或 `json.Marshal` 后整段输出。
- 保留标识时固定打码：手机号 `138****0000`，身份证前 6 后 4，银行卡后 4，token / 密码 / 密钥整体 `***`。
- 请求头只记 `X-Request-Id` / `traceparent`；查询参数同样先脱敏。

## 采样与聚合

- 循环与批处理只在循环外打一条汇总（`total` / `success` / `failed`），禁止逐条。
- 高频成功路径（轮询、每请求必经分支）降 `Debug` 或按 1/N 采样，采样率进配置，不散落常量。
- 同一错误按「错误签名 + 业务键」去重并附 `count` 输出一条；状态变更、金额、支付记录**不采样、不去重**。

## 告警

- 本仓库没有自动告警中间件：需要告警的失败必须显式调用 `utility/feishu` 的 `Manager`（`SendText` / `SendCard`），并异步发送，忽略 `context.Canceled`。
- 循环内必须聚合成一条再发；未配置 webhook 时记 Warn 后跳过，不重试。
- 不假设「打了 Error 就会有人收到」——日志与告警是两条链路。
