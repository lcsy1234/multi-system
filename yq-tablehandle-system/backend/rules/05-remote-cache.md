---
paths:
  - "app/template/internal/logic/**/*.go"
  - "utility/connector/**/*.go"
  - "utility/kafka/**/*.go"
  - "utility/lock/**/*.go"
  - "utility/config/**/*.go"
---

# 外部调用、Redis 与锁

## 外部调用

- HTTP 走 `utility/connector/httpx`，gRPC 走 `utility/connector/grpcx`：客户端按名字从注册表取（`NewRegistry` → `New`/`Get`），**不在 logic 里 new `http.Client` 或手写拨号**。
- 新增对端在 `manifest/config/config.example.yaml` 的 `http.services` / `grpc.clients` 登记名字与地址，代码只引用名字。
- 所有外部调用必须设超时（连接 5s、读取 10s 起步；长耗时接口单独配置并说明依据）；禁止无限等待。
- 外部报文优先定义结构体解析；确需 map 时，JSON 数字只能断言 `.(float64)`，禁止 `.(int)`；金额、单据 ID、时间戳用 `json.Number` 或字符串，避免超过 2^53 丢精度。
- 类型断言必须检查 `ok`，失败要返回可读错误或留日志，不允许用 `_` 吞掉。
- 必须检查响应 `code`/业务状态：非成功转成 `error`，不能只记日志；区分「传输失败」「业务失败」「数据缺失」。
- 下游超时转成 `gerror.NewCodef(gcode.CodeOperationFailed, "调用%s超时，请稍后重试", 对端)`；直接抛原始 error 会被兜底成 `-1` 并脱敏。
- 合法空结果 ≠ 失败：三类结果分开返回，不把「对端没有数据」当异常重试。
- 只有幂等请求才重试；重试与退避有上限，写请求重试前先证明幂等。
- 客户端集中放 `utility/connector` 并复用连接，不在 logic 每次 new。

## Redis

- Redis 是基础设施，不是业务读缓存：禁止 Cache-Aside，业务读以 PostgreSQL 为准；进程内短缓存用 `gcache`。
- 用法限定为：锁、计数、限流、短命状态；TTL 必须显式设置。
- key 用业务唯一键并带域前缀（如 `user:refresh:<id>`），不用全局键。

## 锁

- 用 `utility/lock`：`l := lock.New("业务key")` → `l.Lock(ctx)` → `defer l.Unlock(ctx)`。
- 锁必须带 TTL（默认 10s），长任务用 `Refresh` 续期；禁止无 TTL 的锁。
- 获取失败的错误是 `lock.ErrNotAcquired`（等待超时）与 `lock.ErrNotHeld`（已不持有），分别给出可读提示，不静默放行。
- 等待策略用 `WithWaitTimeout`，不要自己写 `for + time.Sleep` 重试。
- 多资源加锁按固定顺序，避免交叉死锁。
- 锁只是优化：正确性由唯一约束与条件更新兜底（见 [04-transaction.md](04-transaction.md)）。

## Kafka

- 生产端用 `utility/kafka` 的命名 Producer（`kafka.Producer(name)`），消费端用 `utility/kafka` 的 Consumer 注册 handler；不在业务里直接 new `kafka.Writer/Reader`。
- 消息都是 at-least-once：消费幂等由业务保证，不假设「只投一次」。
- 投递在事务提交后（见 [04-transaction.md](04-transaction.md)）；投递失败记日志并留补偿，不宣称必达。
- topic、group、broker 走配置，不进代码。

## 配置

- 运行配置在 `app/template/manifest/config/config.example.yaml`（本地复制成 `config.yaml`）；环境变量按配置路径的大写、点转下划线形式覆盖，可直接写在 `.env` 或部署环境中。
- 强类型配置定义在 `utility/config`，改字段要同步：结构体、默认值、`config.example.yaml`、Nacos 样例五处。
- Nacos 启用时**整份替换**本地配置：远程文档必须包含全部配置块，缺块会以零值运行；改本地样例时同步改远程样例。
- 敏感项走环境变量或配置中心；不提交真实密钥、webhook、证书。
- 环境差异（地址、桶名、开关、租户）放配置，不硬编码。

## 飞书

- 发送走 `utility/feishu` 的 `Manager`（`SendText` / `SendCard`），webhook 与签名密钥来自配置。
- webhook 为空时视为未配置：记 Warn 后跳过，不 panic、不重试。
- 告警类消息异步发送，不阻塞主链路（细则见 [08-logging.md](08-logging.md)）。
