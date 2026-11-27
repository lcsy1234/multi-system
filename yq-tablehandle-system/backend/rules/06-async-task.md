---
paths:
  - "app/template/internal/task/**/*.go"
  - "app/template/internal/logic/**/*.go"
  - "app/template/internal/consts/**/*.go"
  - "utility/kafka/**/*.go"
---

# 异步任务与调度

## 先判断同步还是异步

| 特征 | 选择 | 不要 |
|---|---|---|
| 同库多步写需要原子 | 同一个数据库事务 | 把事务后的动作当成原子的一部分 |
| 外部结果必须立即反馈 | 事务外同步调用 + 超时 | 当成跨系统强一致 |
| 可重试、耗时长、依赖外部 | Asynq，**提交后**投递 | 回滚前投递、事务内等结果 |
| 到点执行、延迟生效 | `Enqueue` 时传 `asynq.ProcessIn/ProcessAt` | `time.Sleep`、进程内定时器 |
| 只是刷新界面 / 发通知 | `low` 队列 | 阻塞主链路 |
| 跨系统最终一致 | 状态校验 + 补偿 + 幂等 | 分布式事务、假设必达 |

## Asynq

- 队列名与任务类型集中在 `internal/consts`：`QueueCritical/QueueDefault/QueueLow`、`TaskTypeXxx`，命名「域:动作」（如 `feishu:notify`）。新增任务同时登记常量、构造函数、handler 与注释。
- 队列权重在 `asynq.queues` 配置；按语义选队列，禁止全部塞 `default`。
- 生产者是**全局命名实例**，在 `internal/task` 里按名字解析（`task.Producer()`）；首次调用时懒创建并缓存，之后 handler 内复用；**不在请求里重复创建**。
- 投递统一 `producer.Enqueue(ctx, taskObject, queue)`；副作用必须在事务提交后投递（事务纪律见 [04-transaction.md](04-transaction.md)）。
- Handler 返回值决定重试：返回 `error` 交给 asynq 按配置重试；不可重试的失败（参数非法、永久拒绝）返回 `asynq.SkipRetry` 并告警，不占重试配额。
- 超过重试上限进 archived（不是 DLQ）；投递或重试耗尽只记日志、不回滚业务，并提供手动重放入口。
- 任务内不开大事务、不做长耗时远程调用；重活拆小或改批量。
- 监控页 `/admin/asynq` 由 `asynq.monitor.enabled` 控制，生产确认 `readOnly` 配置。

## 新增任务的清单

1. `internal/consts`：登记任务类型（队列在投递时选）。
2. `internal/task`：写 payload 结构体与 `NewXxxTask(payload)` 构造函数，序列化失败返回 error。
3. `internal/task`：写 handler，签名 `func(context.Context, *asynq.Task) error`。
4. `internal/task/worker.go` 的 `Processors`：把任务类型映射到 handler，漏注册 = 任务永远不消费。
5. `cmd/worker` 无需改动（Server 统一装配）。

## 幂等

- Handler 幂等的判定要落在「本事件负责的效果」上，不是「有没有收到过」；同库写入由唯一约束、条件更新或行锁保护。
- 处理顺序：① 校验 payload 与业务键，非法直接 `SkipRetry` → ② 事务内校验状态并加锁 → ③ 条件更新后查 `RowsAffected` → ④ 远程推送单独判幂等键，失败必须有补偿。
- 本地写入全部完成才算幂等返回；部分完成要能把剩余部分补完。
- 无统一幂等中间件时用「业务键 + 唯一约束 + 状态条件」组合，不臆造框架（见 [project.md 的「未落地项」](project.md#未落地项tbd禁止臆造)）。
- 至少验证：非法 payload、不存在、状态冲突、顺序重复、并发重复、回滚重试；并发与回滚需真实数据库，跑不了标「未验证」。

## 消息（Kafka）

- Kafka 是独立通道，与 Asynq 不混用：Asynq 管本进程内部的重试型副作用，Kafka 管跨系统的事件流。
- 都是 at-least-once；消费幂等由业务保证，不假设「只投一次」。
- offset / commit 语义以 `utility/kafka` 的封装为准，改消费逻辑前先读实现。
- 不假设 Outbox、事务消息或全链路可靠投递；集成现状见 [project.md 的「外部集成现状」](project.md#外部集成现状)。
- 新增消费者评估分区、副本与重复消费的影响；topic、group 走配置。

## 定时

- 本仓库没有扫表进程与调度中心：周期任务用 asynq 的 `ProcessIn` 自循环投递，或外部 cron 调接口。
- 多实例部署下周期任务必须加锁（`utility/lock`），否则每台机器都会跑一遍。
