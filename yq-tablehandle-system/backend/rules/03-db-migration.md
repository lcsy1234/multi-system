---
paths:
  - "app/template/manifest/migration/**/*.sql"
  - "app/template/internal/dao/**/*.go"
  - "app/template/internal/model/**/*.go"
---

# 数据库、迁移与查询

## 迁移

- 任何表结构变更都走 goose（`app/template/manifest/migration/`），禁止直接改库再补代码；SQL 通过 `embed.go` 打进二进制。
- 文件名 `YYYYMMDDHHMMSS_<表名>.sql`，由 `make migrate-create name=xxx` 生成；一个文件只做一件事（建表、加字段、加索引分开）。
- Up 与 Down 都必须可执行；Down 按依赖倒序 `DROP`。
- **已上线的迁移不回头改**，一律新增；功能未上线时才允许改初始脚本。
- 迁移必须兼容旧代码运行：加字段给默认值，不立刻删列；先迁移后发布。
- 数据回刷、批量修正另起脚本并单独授权：写明环境、匹配键、预期行数、核对方式与回滚依据，不塞进迁移。
- 执行入口：`make migrate-up` / `migrate-up-one` / `migrate-down` / `migrate-status`（内部是 `template migrate -action=...`）。

## 表设计（逐条对照）

### 主键、命名与快照

- 主键 `BIGSERIAL PRIMARY KEY`；业务编码另存 `code`，表间关联优先 `*_code`。
- 冗余 `*_code` / `*_name` 快照：列表少 JOIN，单据保留历史名。
- 表名、字段名、索引名一律小写蛇形；表名单数、无 `t_` / `tbl_` 前缀。
- 命名：主从 `xxx` / `xxx_item`，明细 `_detail`，标准 `_std` / `_std_rule`，关系 `_rel`，任务 `task_*`。

### 通用列（8 个，硬性要求）

**每张业务表都必须有，一个字段都不能省，没有例外表**（`goose_db_version` 这类框架表除外）：

| 列 | 类型 | 含义 |
|---|---|---|
| `created_by` | `BIGINT NOT NULL DEFAULT 0` | 创建人 id |
| `created_name` | `VARCHAR(64) NOT NULL DEFAULT ''` | 创建人姓名快照 |
| `created_at` | `TIMESTAMPTZ NOT NULL DEFAULT NOW()` | 创建时间 |
| `updated_by` | `BIGINT NOT NULL DEFAULT 0` | 最近更新人 id |
| `updated_name` | `VARCHAR(64) NOT NULL DEFAULT ''` | 更新人姓名快照 |
| `updated_at` | `TIMESTAMPTZ NOT NULL DEFAULT NOW()` | 最近更新时间 |
| `is_deleted` | `BOOLEAN NOT NULL DEFAULT FALSE` | 逻辑删除标记 |
| `version` | `INT NOT NULL DEFAULT 0` | 乐观锁版本号 |

- 新增表在同一个迁移里建齐这 8 列；缺列由 `TestMigrationUniversalColumns` 静态解析迁移拦住，不需要数据库。
- 写入直接用 `g.Map`；`created_at` / `updated_at` 与 `created_by` / `created_name` / `updated_by` / `updated_name` 都在代码里显式写入（新增写创建与更新，修改只写更新）。模板没有身份体系，操作人暂写 `0` / 空串，接入登录后替换。
- 名字用 `_name` 快照而不是 JOIN 实时查：人员离职或改名后，历史记录仍要能读出当时是谁。
- `is_deleted` 只允许置 `true`，不做物理 `DELETE`；`baseDao.One/List/Page` 已内置过滤，手写查询必须自己加 `is_deleted = false`。
- 软删记录仍保留 `biz_key`（唯一索引包含软删行）：幂等回查/更新都过滤 `is_deleted = false`；同键再次创建返回 `900030`，更新已删除记录返回 `900040`，不做隐式复活。
- `version` 的更新只走 `dao.UpdateWithVersion`（`WHERE version = ?` 且 `version = version + 1`，0 行受影响即冲突）；logic 里禁止裸 `.Update()`，由 `TestLogicHasNoRawUpdate` 拦住。

### 注释与默认值

- **每个字段都必须有 `COMMENT ON COLUMN`**，包括主键：`gf gen dao` 的 `description` 直接取自它，没注释生成的 entity 就是 `description:""`。
  - 状态字段把每个取值写进注释。
  - 迁移里的 `--` 注释不能替代 `COMMENT ON COLUMN`——生成器读不到。
- **默认 `NOT NULL DEFAULT`**；只有「未设置 / 已清空」确有三态语义时才允许可空，并在注释写明。
- 时间用 `TIMESTAMPTZ`；纯日期用 `DATE`。不用 `BIGINT` 存 Unix 秒。
- 结构化扩展用 `JSONB`；需要按路径查询的字段建表达式索引。
- 小数用 `NUMERIC`（单位见 [project.md](project.md)）。

### 状态与类型

- 用 `SMALLINT`（固定字符串集合也可用 `VARCHAR` + 注释列出取值），取值进 `internal/consts`，不写裸数字。
- **状态 / 类型取值从 1 起连续编号，禁止用 0**：0 与 Go 零值重合，库里存 0 时分不清「未设置」和「待处理」。
  - 常量从 `1` 开始（如 `XxxStatusDraft = 1`、`RecordStatusEnabled = 1`），连续编号，不加 10/100 间隔；「非正常流转」值（如作废）放末尾非连续位并在注释写明。
  - 列的 `DEFAULT` 不能是 `0`；必须建 `CHECK (<列> <> 0)`——只靠应用层赋值挡不住手工 SQL 与批量脚本。字符串枚举列建 `CHECK (<列> <> '')`。
  - 守卫：`TestStatusConstantsStartAtOne`（常量）、`TestMigrationUniversalColumns`（8 列 + 列注释）；列级 `CHECK` 的集成守卫 `TestStatusColumnsHaveCheck` 待重建。

### 约束与索引

- **外键默认不建**：一致性在 logic 保证（便于分库与批量导入）；确需强约束时在迁移注释写明理由。
- 幂等：需要防重的写接口建 `uk_*`（含部分唯一索引）；唯一键是正确性兜底，锁只是优化。
- 索引不是越多越好：为真实查询路径建，命名 `idx_<表>_<列>`；新增查询先看 `EXPLAIN`。

## 查询

- 读走 dao：`dao.Xxx.One/List/Page`（`baseDao` 提供）+ `filter` 构造器，不在 logic 里手拼 `g.DB()`（事务内除外）。
- 自定义查询追加在 `internal/dao/<表>.go` 末尾，或写成 `filter.QueryBuilder`；不放进 `internal/` 子包。
- 条件结构用 `filter:"<op>"` 标签：`eq`（默认）、`not`、`lt`、`lte`、`gt`、`gte`、`in`、`like`；列名默认取字段名的 snake_case，需要换列就写 `filter:"eq,real_column"`，`filter:"-"` 表示不参与。
- **没有 `filter` 标签的字段不参与 WHERE**——不要为了「方便」给请求结构体的所有字段加标签。
- `like` 必须转义用户输入（`%`、`_`）；若 `utility/filter` 未做转义，先在 filter 包补齐再使用。
- 列名取 `dao.Xxx.Columns()`，不写字符串。
- **列表 / 分页查询必须定义 `model.XxxFilter`**：筛选字段放 `Filter` 结构体，由 `filter.PaginationInput[XxxFilter]` / `filter.ListInput[XxxFilter]` 传入。
  - 禁止在 api 包里定义筛选结构，也禁止在 logic 里手拼 WHERE。
  - `Filter` 实现 `Build(m *gdb.Model) *gdb.Model`（内部调 `filter.BuildFilter`），既是数据也是构造器。
  - 守卫：`TestListQueriesUseFilter`（列表接口必须有 `model.XxxFilter`）。
- **查询结果扫描到 `model.XxxOutput`，不把 `entity` 往 service 边界外传**：`entity` 带 `orm` 标签、跟着表走，跨层传递会把列名绑进上层。
  - `Output` 只保留调用方需要的字段；扫描依赖列名与字段名的 snake_case ⇄ CamelCase 约定。
- 禁止 `SELECT *`；更新只传变更字段或 `OmitNil()`，不整行覆盖。
- `Update`/`Delete` 必须带条件；批量更新后核对 `updated_at` 是否落库。
- **空切片 `IN` 必须显式短路**，否则生成非法 SQL：

```go
if len(codes) == 0 {
    return nil, nil
}
```

- 查询默认带业务过滤（状态、组织、租户）；漏 `where` 是最高频事故。
- 排序字段取 `dao.Xxx.Columns()` 或白名单，不接收前端任意列名。
- 分页必须设上限；深分页用游标（`id > lastId`），不用大 offset。
- 批量写前用同条件 `SELECT COUNT(*)` 预估；超出预期（如 >1000 行）先确认再执行；单批建议 ≤500 行，禁止循环单条 Insert。
- 禁止在 HTTP 请求内做全量扫表；批量归 worker。
- 唯一约束冲突要翻译成业务错误（谁占了、哪张单），不要把 `duplicate key` 原样丢给前端。

## 生成物

- `make dao` 前必须先执行迁移并在 `app/template/hack/config.yaml` 的 `gen.dao.tables` 登记新表；迁移建的表与 `dao/internal` 必须一一对应，由 `TestMigrationTablesMatchDAO` 拦住。
- `dao/internal`、`model/entity`、`model/do` 不手改；`dao/<表>.go` 只允许末尾追加自定义方法（通用清单见 [10](10-goframe-layering.md#生成器顺序)）。
- `make dao.overwrite` 会重刷 `dao/` 索引文件并丢弃手改，只在确认没有自定义方法时执行。
