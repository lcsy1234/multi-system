---
paths:
  - "**/*_test.go"
  - "**/go.mod"
  - "Makefile"
  - "app/template/Makefile"
---

# 测试与审查

## 从验收到证据

- 每项验收条件先定验证方式；修缺陷先写失败用例复现，重构先确认原有行为。
- 断言可观察的输出、状态、副作用与异常语义；不用私有方法名、源码字符串或方法数量代替行为。
- 优先表驱动测试，用例名说明场景（`TestBuildFilter/when_value_is_zero_should_skip`）。
- Mock 只隔离外部依赖，不 Mock 被测业务逻辑。
- 必须覆盖：状态流转的非法分支、事务回滚后无副作用、锁竞争与超时、外部报文异常（字段缺失/类型不符）、重复提交幂等。
- 单元测试证明不了 SQL、事务、锁与异步投递；涉及这些边界时跑集成测试，跑不了就标「未验证」。

## 测试文件位置

- 本仓库现状是**同包测试**：`utility/` 下各包自带的 `*_test.go`（config、filter、lock、middleware、feishu、kafka、connector/grpcx、connector/httpx）。
- 新增单元测试放被测包内、与被测文件同目录；需要从包外视角时用 `<包名>_test` 包名。
- 需要真实 PostgreSQL / Redis / Kafka 的用例加构建标签，默认不跑：


- 跑集成用例前先准备真实依赖：PostgreSQL / Redis / Kafka 的本机或外部实例，并说明用了哪个；跳过 ≠ 通过。
- 不把测试塞进纯数据结构包（没有行为可测）。

## 执行

| 层级 | 证明什么 | 怎么跑 | 跑不了时 |
|---|---|---|---|
| 编译 | 类型、依赖、生成代码一致 | `make check`（角色文件构建 + 包级 build，排除 `boot/`） | 编不过 = 阻塞交付 |
| 静态检查 | 缺陷与格式 | `go vet ./...`；改动文件 `gofmt -l` 无输出 | 新增问题阻塞，存量记基线 |
| 单元测试 | 纯逻辑、边界、错误分支 | `make test`（排除 `boot/`） | 标「未验证」 |
| 集成测试 | SQL、事务、锁、序列化、异步 | `go test -tags=integration ./...` + 真实依赖 | 标「未验证」并写明缺什么 |
| 人工验收 | 网关、真机、外部服务 | 按验收清单 | 不开工或留人工 |

- 各层级分别报告，互不替代；「打包跳过测试」不算通过。
- `make check`（角色构建 + build/vet + 单元守卫）是提交前的最小集合；`make test` / `make vet` / `make fmt` 等价于上表。
- `make build` 在仓库根用 `go build` 产出 `bin/template-<role>`，不需要 gf CLI；`make -C app/template build` 才走 `gf build`（需 gf CLI）。没有 CLI 时用 `make check` 证明编译（包级命令已排除 `boot/`），并写明用的是哪条命令。
- 未配置且未实际执行的 `golangci-lint` 不得写成「已通过」。
- 使用项目现有命令，不为这次验证引入新框架；覆盖率门槛只在 CI 明确配置时成立。
- 工具链问题（本机 Go 低于 `go.mod` 要求）按 [`project.md` 验证命令](project.md#验证命令)处理，不因此跳过验证。

## 守卫测试：把踩过的坑变成会失败的检查

只写进文档的约定会退化；能自动失败的检查不会。遇到「同一个坑踩第二次」时，**加一条守卫测试**，而不是补一段说明。规则与守卫的对应关系登记在 [`README.md` 硬性要求速查](README.md#硬性要求速查)。

运行方式：

```bash
make check                                                     # 角色构建 + build/vet + 单元守卫
cd app/template && go test -tags=integration ./...                    # 集成守卫（需真实 PG）
```

### 单元守卫（`internal/guard`，无需数据库）

| 守卫 | 断言 | 拦住的坑 |
|---|---|---|
| `TestAPIFieldsAreDocumented` | `api/` 下所有结构体的具名字段都有 `json` 与 `dc` | 字段名漂移；按 `dc` 反射匹配时静默取空 |
| `TestModelFieldTags` | `internal/model` 顶层字段有 `json` + `dc`，且不许有 `orm` | Output 被当成实体；改列名时误改边界类型 |
| `TestModelNaming` | `internal/model` 每个结构体以 `Input` / `Output` / `Filter` 结尾 | 分不清 service 边界类型与内部结构 |
| `TestAPIImportsNoEntity` | `api/` 不 import `model/entity` 或 `model/do` | Res 直接暴露实体，改列名就破坏调用方 |
| `TestMigrationNaming` | 迁移文件名为 `时间戳_表名.sql`，一个文件只建一张表 | 迁移不可读、无法按表回溯 |
| `TestLogicHasNoRawDB` | `internal/logic` 里不出现 `g.DB().Model` | 绕过 dao 拼查询，过滤与事务规则失效 |
| `TestListQueriesUseFilter` | 列表/分页接口必须有 `model.XxxFilter` | 筛选字段散落在 api 包，绕开 `filter` 标签与列名约定 |
| `TestLogicHasNoRawUpdate` | `internal/logic` 里不出现 `.Update()` | 绕过乐观锁，并发请求都能写成功 |
| `TestStatusConstantsStartAtOne` | 状态/类型常量从 `1` 起，且同组最小值为 1 | 0 与零值混淆、编号出现空洞 |
| `TestMoneyUsesDecimal` | 扫描 `model` / `entity` / `do` 拒绝 `float64`；`config.example` 保留 decimal/numeric → `decimal.Decimal` 的 `typeMapping` | 有人改回 `float64` 或删掉生成映射，精度丢失 |
| `TestMigrationStatusColumnsHaveCheck` | 迁移出现 status/type 列时，必须带 `CHECK (status <> 0)` | 状态列缺少列级约束，脏数据可写入 |
| `TestMigrationConstraintsIntegration`（integration） | 真实 PG 应用迁移后检查通用列、注释、唯一索引与 CHECK | 静态解析漏掉类型/默认值/约束 |
| `TestBusinessMethodEntryLogs` | controller / logic 的导出方法入口第一条语句必须是 `glog.Infof` 调用 | 新方法没有入口日志，排障时断链 |
| `TestReadmeGoSnippetsParse` | README 的每个 `go` 代码块必须能按文件或函数体解析 | 示例出现语法错误，读者无法编译 |
| `TestMigrationUniversalColumns` | 静态解析迁移：每张业务表有 8 个通用列，每列有 `COMMENT ON COLUMN` | 新表漏列、漏注释（替代原集成版通用列/注释守卫） |
| `TestControllerMethodLength` | controller 方法 ≤ 20 行（AST 统计） | controller 变大、塞进业务分支 |
| `TestToProtoDemoItem` / `TestToProtoDemoItems` | gRPC 转换：标量字段、nil 时间戳、列表顺序 | controller 手工拼 proto 漂移 |
| `TestAllTableDAOsEmbedBaseDao` | `internal/dao/*.go` 的每个表 DAO 都嵌入 `*baseDao` | 新增表绕过 make dao / 形状脚本漏表 |
| `TestMigrationTablesMatchDAO` | 迁移建的表与 `dao/internal` 的 table 一一对应 | 建了表没生成 DAO，或 DAO 指向已删表 |
| `TestReadmeEntryPaths` | README 只写一种入口方案，四种角色文件/目录齐全 | 文档与入口漂移 |
| `TestEntryDocsConsistency` | 合并入口不存在时，README/AGENTS/rules 的纯文本与目录树也不得再提 `main.go`、`go build ./app/template` | 元守卫只看反引号路径，漏掉纯文本漂移 |
| `TestReadmeModelTypes` | README 示例里的 `model.Xxx` 必须真实存在于 `internal/model` | 文档抄错类型名，读者编译失败 |
| `TestDocsLifecycleADR` | ADR 编号、Status 合法；superseded 必须带日期与原因 | 决策被删除，或只标 superseded 不写为什么 |
| `TestDocsLifecycleGlossary` | 每个术语块有非空 `_Avoid_` | 词表退化成只有词条、没有边界 |
| `TestDocsLifecycleIssues` | `docs/issues/README.md` 存在；每个 feature 有带 `Status:` 的 spec.md | 关闭/终止的 spec 无状态或目录被删 |
| `TestMetaSpeedTableGuardsExist` | 速查表点名的 `Test*` 必须在 `internal/guard` 真实存在 | 检查名悬空、规则引用已删测试 |

### 集成守卫（`//go:build integration`，需真实 PostgreSQL）

| 守卫 | 断言 | 拦住的坑 |
|---|---|---|
| `TestMigrationConstraintsIntegration` | 在真实 PG 上应用迁移，检查通用列、列注释、`biz_key` 唯一索引与状态列 CHECK | 静态解析漏掉类型、默认值、NOT NULL、约束 |

> 集成守卫需要 `TEST_DATABASE_DSN`；默认 `make check` 只做编译，不执行。新增业务域时同步扩展该用例。

**覆盖率的两个层次：**`TestMetaRuleCoverage` 同时输出 `static`（有任意自动检查）与 `behavior`（速查表标「静态+集成」且存在 `*Integration` 测试）。当前 static 11/11、behavior 3/11；static 100% 不代表行为已验证。

### 形状锁

`internal/dao/shape_test.go` 不含逻辑，只有编译期引用（`var _ = DemoItem.Page`）：生成器把 `baseDao` 冲掉时**编译就失败**，而不是运行时报 nil。`TestAllTableDAOsEmbedBaseDao` 再静态扫描所有表 DAO，新增表漏补形状也会失败。

`make dao` 生成后会自动补齐形状。v2.10.3 的 CLI 不支持自定义 DAO 模板（旧模板已删除），别指望模板，也别绕过 `make dao` 单独跑生成器。

加守卫的判据：这个坑能被静态断言（标签、文件名、符号引用、调用形状、库结构）就加守卫；只能靠人判断的，才写进对应专题。

## 范围与存量失败基线

- 修改前记录命令、工作区状态、Go 版本与失败项；修改后同条件复跑，结论才有可比性。
- 新增或加重的失败必须修；无关的既有失败单独报告，须有修改前证据。
- 禁止关检查、忽略退出码、跳过用例来制造通过；既有问题阻塞验证时，该项明确标「未验证」。
- `gofmt -l` 只对本次改动文件要求空输出；全仓格式问题只记基线。
- 仅改文档：核对引用、命令与一致性，不跑无关业务测试。

## 提交前自查

按命中专题检查，结论留在审查记录，不写进代码注释：

- [ ] 分层、身份来源、数据范围符合 [02](02-api-validation.md) / [10](10-goframe-layering.md)，无越权、无「从参数信任身份」。
- [ ] 事务 `ctx`、状态条件、提交后副作用符合 [04](04-transaction.md) / [06](06-async-task.md)；失败窗口有补偿，消费可幂等。
- [ ] SQL 参数绑定、批量写、空集合短路、生成物与迁移同步符合 [03](03-db-migration.md)。
- [ ] 外部调用超时、失败分类、日志脱敏符合 [05](05-remote-cache.md) / [08](08-logging.md) / [00](00-core.md)。
- [ ] proto 字段号未复用、生成文件已重跑 `make proto`（如涉及），符合 [11](11-grpc-protobuf.md)。
- [ ] 新增或明显变长的方法已按 [09-code-structure.md](09-code-structure.md) 做结构复查；成功、参数错误、权限不足、状态冲突均有测试或说明。

## 审查结论

- 给出具体 `文件:行`、可达路径、影响与证据；区分功能缺陷、明确规则违规与可选改进。
- 不把个人偏好或其他项目模板写成缺陷；结论落到「改哪一行、为什么」。

## 交付报告

固定四段，缺项说明原因：

```markdown
## 改动
- <模块/文件>：<做了什么，为什么>

## 验证
| 命令 | 结果 |
|---|---|
| `<实际执行的命令及范围>` | 通过 / 失败（新增或既有基线）/ 未执行（原因） |

## 未覆盖 / 阻塞
- <项>：<缺什么环境，谁跟进>
```

- 有验收清单就逐项关联条件、位置与证据。
- 分别说明代码完成、自动化检查、运行时验收与发布前置条件；没跑的验证不写成通过，环境阻塞不算失败但要写明。
