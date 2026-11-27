---
paths:
  - "app/template/internal/logic/**/*.go"
  - "app/template/internal/dao/**/*.go"
---

# 事务与一致性

- 事务边界在 logic，默认 `g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {...})`；不要为了「保险」给只读方法套事务。
- 主表与明细（头 + item / rule）必须在同一事务提交，不允许先写头再异步补明细。
- 事务内只做数据库读写；Kafka、Asynq、RPC、文件上传、HTTP 一律移到提交之后。
- 事务回调内必须用回调传入的 `ctx`，不要用外层 `ctx`，否则拿不到 tx。
- 事务内不得吞错：返回 `error` 即回滚；不允许 `if err != nil { glog.Error(...) }` 后继续。
- 不可撤销的副作用在事务成功后投递；参数在事务内收集成切片，提交后统一 `Enqueue`。
- 跨系统动作顺序：先远端、后本地。远端失败时本地不动；本地失败时远端已有单据，由重试或对账补偿。
- 禁止在长事务里做远程调用、`time.Sleep`、循环重试；确需执行时说明一致性与补偿策略。
- 状态流转先校验当前状态，再用**条件更新**（`WHERE status = ? AND version = ?`）落库；锁只是第一道闸，不能替代条件更新。
- 条件更新必须查 `RowsAffected`：0 行要区分「已完成 / 不存在 / 状态冲突」，返回可读错误，不能一律当成功。
- 事务内不打「成功」日志——事务可能回滚，成功日志放在提交之后。
- 提交后投递不等于必达：写清失败重试、补偿或落表兜底（见 [06-async-task.md](06-async-task.md)），不宣称消息必达。
- 跨库不做分布式事务；最终一致 = 状态校验 + 补偿 + 幂等。

## 事务内怎么用 dao

- 事务内的读写用回调里的 `tx.Model(...)`；`dao.Xxx.One/List/Page` 用的是默认连接，**在事务内用它们会读到事务外的数据**。
- 需要「事务内 + filter 构造器」时，写成 `filter.Build(tx.Model(...), builder...)`，不要绕过事务直接用 `dao`。

## 乐观锁与并发

- 并发写同一行的默认方案是 `version` 条件更新：`WHERE id = ? AND version = ?`，成功后 `version = version + 1`。
- 统一走 `dao.UpdateWithVersion`（见 [03](03-db-migration.md)），不在 logic 里手写 `.Update()`。
- 冲突时返回可读文案（「单据已被他人变更，请刷新」），不后台静默重试、不覆盖后写者。
- 需要悲观锁时用 `Model.LockUpdate()`（`SELECT ... FOR UPDATE`），说明为什么乐观锁不够，且锁必须落在同一事务内。

## 取号

- 单号不用工具包，直接用**库内序列表**：`INSERT ... ON CONFLICT (类型, 年月) DO UPDATE SET last_val = last_val + 1` 再读回，两条语句在同一事务内完成（行锁保持到提交）。
- 前缀常量放 `internal/consts`，禁止在调用处写裸字符串。
- 取号会按序列键串行化，锁粒度限定在 `(类型, 年月)` 这一行，不做全局锁。
- 禁止 `max(no)+1`、随机数、「先查再插」。
- 序号允许跳号、不允许重号；回滚不退还。
