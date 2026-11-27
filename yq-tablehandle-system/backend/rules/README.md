# 规则路由

改代码、表结构、配置或规则时，按顺序读：

1. [`00-core.md`](00-core.md)——团队红线；
2. [`09-code-structure.md`](09-code-structure.md)——结构与职责；
3. [`07-testing-review.md`](07-testing-review.md)——怎么验证。

再按下方路由表补读命中项。只回答问题、不改仓库时不读。[`project.md`](project.md) 与 [`02-api-validation.md`](02-api-validation.md) 写明的差异优先；未写明的冲突先确认。

按任务语义读取全部命中项，不只看修改的文件名。`paths` 只辅助编辑器自动加载，不能替代这张路由表；范围扩大时补读。

## 硬性要求速查

下表是**违反即失败**的条款，每条都有自动检查兜着。改代码前扫一眼，提交前跑 `make check`（`build + vet + 单元守卫`）；集成守卫需要真实 PostgreSQL。

| # | 要求 | 检查 | 级别 | 详解 |
|---|---|---|---|---|
| 1 | 生成物不手改；改表先迁移再 `make dao` | `internal/dao/shape_test.go`（编译期） | 静态 | [10](10-goframe-layering.md) |
| 2 | 金额一律 `decimal.Decimal`，库内 `NUMERIC(18,2)` | `TestMoneyUsesDecimal`（AST 扫描 + typeMapping） | 静态 | [01](01-go.md) |
| 3 | `model` 结构体名必须以 `Input` / `Output` / `Filter` 结尾 | `TestModelNaming` | 静态 | [10](10-goframe-layering.md) |
| 4 | `api` 字段必须有 `json` + `dc`；`model` 字段必须有 `json` + `dc` 且不许有 `orm`；`api` 不引用 `entity`/`do` | `TestAPIFieldsAreDocumented`、`TestModelFieldTags`、`TestAPIImportsNoEntity` | 静态 | [02](02-api-validation.md) |
| 5 | 每张业务表必须有 8 个通用列（`created_by`/`created_name`/`updated_by`/`updated_name` + `is_deleted` + `version`） | `TestMigrationUniversalColumns`；`TestMigrationConstraintsIntegration` | 静态+集成 | [03](03-db-migration.md) |
| 6 | 每个字段必须有 `COMMENT ON COLUMN`（含主键） | `TestMigrationUniversalColumns`；`TestMigrationConstraintsIntegration` | 静态+集成 | [03](03-db-migration.md) |
| 7 | 状态/类型从 `1` 起连续编号，禁止 `0`，并建 `CHECK (status <> 0)` | `TestStatusConstantsStartAtOne`、`TestMigrationStatusColumnsHaveCheck`；`TestMigrationConstraintsIntegration` | 静态+集成 | [03](03-db-migration.md) |
| 8 | 更新只走 `dao.UpdateWithVersion`，logic 里禁止裸 `.Update()` | `TestLogicHasNoRawUpdate` | 静态 | [03](03-db-migration.md) |
| 9 | logic 不直接 `g.DB().Model`，读走 `dao` + `filter` | `TestLogicHasNoRawDB` | 静态 | [03](03-db-migration.md) |
| 10 | 迁移文件名 `YYYYMMDDHHMMSS_<表名>.sql`，一个文件只做一件事 | `TestMigrationNaming` | 静态 | [03](03-db-migration.md) |
| 11 | 列表/分页查询必须有 `model.XxxFilter`，不得在 api 包定义筛选结构 | `TestListQueriesUseFilter` | 静态 | [03](03-db-migration.md) |

守卫清单与新增守卫的判据见 [07](07-testing-review.md#守卫测试把踩过的坑变成会失败的检查)。

## 按任务读取

| 任务范围 | 必读规则 |
|---|---|
| 改代码、表结构、配置或规则 | [`00-core.md`](00-core.md) |
| 编码、重构、审查或方案设计 | [`09-code-structure.md`](09-code-structure.md) |
| Go 代码或设计 | [`01-go.md`](01-go.md) |
| GoFrame 分层、生成器、目录归属、进程 | [`10-goframe-layering.md`](10-goframe-layering.md) |
| api 定义、controller、请求响应契约 | [`02-api-validation.md`](02-api-validation.md) |
| gRPC、protobuf、proto 生成 | [`11-grpc-protobuf.md`](11-grpc-protobuf.md) |
| 迁移、表设计、dao/model、查询与过滤 | [`03-db-migration.md`](03-db-migration.md) |
| 写流程、事务、一致性 | [`04-transaction.md`](04-transaction.md) |
| 外部调用（HTTP/gRPC/Kafka/飞书）、Redis、锁 | [`05-remote-cache.md`](05-remote-cache.md) |
| Asynq、事件、延迟任务、周期任务 | [`06-async-task.md`](06-async-task.md) |
| 日志、错误上报、飞书告警 | [`08-logging.md`](08-logging.md) |
| 功能实现、缺陷修复、重构、测试、审查 | [`07-testing-review.md`](07-testing-review.md) |

完成标准：命中规则已应用；验收条件已核对；`make check` 通过；实际验证、失败与未执行项已明确报告。

## 维护约定

- 编号即职责：`00`–`11` 不重排、不改含义；新增专题用下一个空号（`12` 起），不动既有编号。
- 只在拦不住真实问题时加规则；每条要能回答「是 / 否」，写清怎么做，不写态度。
- **能被静态断言的规则必须配守卫**（标签、命名、文件名、符号引用、调用形状进 `internal/guard`；数据与库结构进集成测试），并在上面的速查表登记。只写文档不配检查的规则，默认会退化。
- 事实归 [`project.md`](project.md)，做法归专题，不在两边重复；两边冲突时以 `project.md` 与本文件为准。
- 改标题必须同步引用它的锚点（本文件、其他专题）。
- 元守卫：`TestMetaRuleGuardsExist`（规则点名的 `Test*` 必须存在，已移除的须在该行标注「待重建/已移除」）、`TestMetaGuardRegistered`（`internal/guard` 的守卫必须在规则中登记）、`TestMetaDocsReferences`（文档里的链接、`make` 目标、`go run` 路径与仓库路径必须存在）、`TestMetaRuleCoverage`（速查表每行必须有自动检查或显式标「手动/待重建」，静态 11/11，集成 3/11（迁移约束，需 PostgreSQL），手动 0）、`TestMetaSpeedTableGuardsExist`（速查表点名的 `Test*` 必须在 `internal/guard` 真实存在）。`TestMeta*` 不参与反向登记检查。
