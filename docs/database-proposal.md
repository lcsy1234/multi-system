# 当前 MySQL 存储

用户已选择 MySQL。当前仅 MES 使用数据库，SCM 已恢复为原始骨架。

- 地址：本机 `127.0.0.1:3306`。
- 数据库：`multi_system_mes_demo`。
- Navicat 连接：`study-multiSys`。
- 连接变量：`MES_MYSQL_DSN`。
- SQL：`mes/internal/storage/schema.sql`。

当前使用三张表：

| 表 | 用途 |
| --- | --- |
| `material_requisition` | 领料申请、排队/消费状态 |
| `material_requisition_line` | 物料明细与 DECIMAL(18,4) 数量 |
| `integration_event_log` | 提交和消费过程的日志 |

执行 `make db-init` 可创建初始结构。默认连接本机 root 无密码；自有账号通过环境变量设置，代码不保存真实密码。会话使用 UTC，页面按 Asia/Shanghai 展示。

当前没有 outbox 写入或调度逻辑。之前已创建的 `outbox_event`、SCM 库、额外字段和历史数据保留在 MySQL 中，本次撤回代码不删库、不删表、不清空数据。
