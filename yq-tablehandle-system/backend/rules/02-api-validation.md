---
paths:
  - "app/template/api/**/*.go"
  - "app/template/internal/controller/**/*.go"
---

# 接口与校验

接口契约以本仓库为准，不是通用 REST。单 HTTP 服务（`template http`），分组 `/v1`，没有 admin/api 双端分权。

阅读顺序：**位置与生成 → 命名 → g.Meta → 身份 → 字段 → 响应 → 错误码 → v 规则 → 回调 → 检查与报告**。

## 1 位置与生成

- api 定义：`app/template/api/<模块>/v1/<模块>.go`（Req/Res 手写）；`app/template/api/<模块>/<模块>.go` 是生成的 `I<模块>V1` 接口，**不手改**。
- controller：`app/template/internal/controller/<模块>/`。`<模块>_new.go` 生成一次后可改；`<模块>_v1_<动作>.go` 由 `make ctrl` 生成骨架，手写实现。
- 新模块必须在 `internal/cmd/http.go` 的 `/v1` 分组 `group.Bind(...)` 注册，否则访问不到；完整顺序见 [10-goframe-layering.md](10-goframe-layering.md) 的「生成器顺序」。

## 2 命名

一个动作一对结构体，前缀相同，GoFrame 去掉 `Req` 得到方法名。

| 动作 | 结构体 | method | path 后缀 |
|---|---|---|---|
| 分页列表 | `GetListReq` / `GetListRes` | `post` | `/get-list` |
| 详情 | `GetOneReq` / `GetOneRes` | `post` | `/get-one` |
| 创建 | `CreateReq` / `CreateRes` | `post` | `/create` |
| 修改 | `UpdateReq` / `UpdateRes` | `post` | `/update` |
| 删除 | `DeleteReq` / `DeleteRes` | `post` | `/delete` |

- 标准 CRUD 用上表名字；其他动作用帕斯卡动词短语，path 用同一短语的 kebab-case：`UpdateStatusReq` → `/update-status`。
- 每个 Req 都有同前缀的 Res；Res 可以是空结构体。

## 3 g.Meta

每个 Req 嵌入 `g.Meta`，带齐 `path`、`method`、`tags`、`summary`。

- `path` 以 `/` 开头、全小写 kebab-case，同一模块共用一个资源前缀（如 `/demo/items`）；不带版本段（版本在分组 `/v1`）。
- `method` 一律 `post`：不使用 `get`、`delete`、`put`、`patch`，查询、详情、删除也不例外。
- 存量 `get` 是待迁移债务（`/demo/items` 列表就是）：不新增、不扩散，改动该接口时一并改为 `post` 并同步调用方。
- `tags` 是中文业务名，同一模块每个接口相同；`summary` 是中文，说明动作，同一文件内彼此不同。
- 同一分组内 `path` 唯一。

## 4 身份与数据范围

- **本仓库尚未实现登录**：`cmd/http.go` 只挂了 `RequestID` 与 `HandlerResponse`，没有鉴权中间件、没有账号体系（见 [`project.md` 未落地项](project.md#未落地项tbd禁止臆造)）。
- 登录落地前，**任何从请求参数或请求头信任身份的写法都是认证绕过**，不得进入主干，也不得作为「临时方案」合入。
- 身份只从 ctx 取（将来由登录中间件写入）；controller 不接受前端传入的 `accountId`、部门、越权过滤条件。
- 新增接口默认要求已认证；确需匿名（`/healthz`、`/readyz`、外部回调）时在 `cmd/http.go` 显式挂到无鉴权路径，并在评审说明理由。
- 详情与写操作先校验归属：能查到 ≠ 有权操作。
- 数据范围在 logic 里收口，不在 controller 拼 WHERE。

## 5 请求字段

- `json` 用驼峰，与业务含义一致。
- 字段说明**统一用 `dc`**，不要用 `description`：`dc:"状态，1=启用，2=停用"`；`description` 只允许出现在生成物里。
- 筛选字段放在独立结构体（如 `ListItemsFilter`），用 `filter:"<op>"` 声明可筛列；**没有 `filter` 标签的字段不参与 WHERE**，这是安全边界，不是遗漏。
- 分页列表的 `page` 从 1 起，`pageSize` 校验 `between:1,300`、默认 20；不另起一套分页字段。
- `v` 标签只做形状校验，规则见第 8 节；必填字段必须带 `v`。
- 可选筛选字段用指针（零值表示未传），必填写入字段用值类型；判据见 [01-go.md](01-go.md) 的「指针还是值」。
- 嵌套结构体的导出字段同样要有 `json` 与 `dc`。
- **`api/` 下每个具名字段都必须同时有 `json` 与 `dc`**（含共享回执字段），由 `TestAPIFieldsAreDocumented` 强制；`model` 侧由 `TestModelFieldTags` 强制且禁止 `orm`。

## 6 响应与 controller

对外 JSON 固定为 [`middleware.APIResponse`](../utility/middleware/response.go)：`code`、`message`、`data`、`requestId`；由 `HandlerResponse` 统一写出。

```go
func (c *ControllerV1) CreateItem(ctx context.Context, req *v1.CreateItemReq) (res *v1.CreateItemRes, err error)
```

- 业务结果放进 `res` 后 return；controller 不调用 `WriteJson`、`WriteStatus`。
- Res 是 `data` 的结构，不再包一层 `code`/`message`。
- **Res 的字段来自 `model.XxxOutput`，不直接嵌 `entity` / `do`**：service 返回 `Output`，controller 负责映射成 Res（字段可同名，语义必须对得上）。
  - 把实体直接当 `data` 返回，等于把库结构当契约；`Output` 不带 `orm`，也不该出现在 SQL 里。
- controller 只做参数转交与结果组装，方法不超过 20 行（`TestControllerMethodLength`）；不写事务、SQL、业务分支；只允许入口 Info 日志（rules/08），不写业务日志。
- 非标准输出（文件流、健康检查）用 `r.Response.WriteStatusExit` 直写，并确认已绕过统一包裹。
- 生产响应不暴露堆栈、SQL、内部路径、密钥；`HandlerResponse` 已对非 client-error 码脱敏。
- 写接口必须有防重：业务键 + 唯一约束 + 条件更新，不假设调用方只来一次。

## 7 错误码

分两层：**gcode 内置码**（框架与基础设施用）与**业务码**（六位，业务规则校验用）。

### 7.1 内置码

| code | 含义 | 触发场景 |
|------|---|---|
| 0    | 成功 | 正常返回 |
| 200  | 操作成功 | 仅兼容老系统逻辑，新接口不用 |
| 401  | 权限不足 | 未登录、凭证错误、越权统一 401 |
| 403  | 禁止访问 | 框架/网关已判定拒绝时透传 |
| 404  | 资源不存在 | 路径不存在、方法不匹配，或业务对象查不到（gcode 65 映射） |
| 422  | 字段验证失败 | `v` 标签在绑定阶段返回 |
| -1   | 系统错误 | “系统错误，请稍后重试！”；兜底 |
- 内置码唯一定义在 `utility/errcode`：`CodeSuccess` / `CodeOKCompat` / `CodeUnauthorized` / `CodeForbidden` / `CodeResourceNotFound` / `CodeFieldInvalid` / `CodeSystemError`；业务码仍登记在 `internal/consts/error_code.go`（见 7.3）。两处都引用常量，不写裸数字。
- 用 `gerror.NewCode` / `gerror.NewCodef` 构造，不自造未登记整数码。
- 裸 `gerror.New` 按系统错误兜底，码为 `-1`（`gcode.CodeNil`），对外提示「系统错误，请稍后重试！」；要展示文案就必须给 client-error 码。
- 422 只给字段校验；gf 的字段级码——`v` 校验 51、参数非法 53、参数缺失 54——由 `HandlerResponse` 统一映射成 422 输出，message 保留原码文案，不要直接返回 51/53/54。
- gf 的未授权码 61（`gcode.CodeNotAuthorized`）由 `HandlerResponse` 统一映射成 401 输出，文案保留（如「未登录」）。
- gf 的未命中码 65（`gcode.CodeNotFound`）由 `HandlerResponse` 统一映射成 404 输出，文案保留（如「资源不存在」）。

### 7.2 401 / 403 / 404 的写出

- 401/403/404 与 422 一样是**内置码**，统一走 `{code,message,data}` 信封，HTTP 状态保持 200，不直写状态码、不用 `WriteStatusExit`。
- 401/403 由鉴权中间件 `r.SetError(...)` 写出，controller 不自行拼状态响应。
- 本项目「权限不足」统一 401，不靠响应码区分「未登录 / 凭证错误 / 越权」，差异只进日志。
- 403 只透传框架/网关的拒绝，不在业务代码里新造 403 分支。
- 404 表示资源不存在：路由/方法不匹配由状态处理器返回，业务对象查不到由 gcode 65 映射而来。
- **现状**：`61 → 401`、`65 → 404` 已由 `HandlerResponse` 映射；`403` 还没有映射，只能靠框架/网关透传；路由不存在时仍是 ghttp 默认 404 页（不套信封）。见 [`project.md` 未落地项](project.md#未落地项tbd禁止臆造)。

### 7.3 业务码（六位）

- 六位定长：**前三位模块码 + 后三位操作位**；操作位从 0 起每段 10 个号，`xxx000` 为该模块通用失败。
- 模块码段按业务域分配，不跨段借用；未列出的段（如 4xx）不自行占用，需要时先在这里登记：

| 模块码段 | 业务域 |
|---|---|
| 1xx | 中台 |
| 2xx | 采购 |
| 3xx | 生产 |
| 5xx | 计划 |
| 6xx | W |
| 9xx | 演示（demo） |

- 常量登记在 `internal/consts/error_code.go`，用 `gerror.NewCodef(consts.CodeXxx, ...)` 构造；不写裸整数，不跳号回填。
- 已登记模块：**900 演示**（`900010` 名称重复、`900020` 版本冲突、`900030` 幂等键被软删记录占用、`900040` 记录不存在或已删除）；新增模块先在此登记，再在 `internal/consts/error_code.go` 落常量。
- 同一操作的多种失败原因共用一码，靠文案区分到具体字段；确需前端分支时再单独分配操作位。
- 只用于业务规则校验失败：字段校验用 422，未分类用 -1，库失败 52。
- 六位业务码（100000–999999）的文案由 `HandlerResponse` 直接透传，不需要额外登记；新增码只要落在该区间即可。

## 8 v 校验规则

### 8.1 内置规则

| 分类 | 规则 | 写法 | 说明 |
|---|---|---|---|
| 必填 | `required` | `v:"required"` | 零值即失败：`""`、`0`、`false`、`nil` |
| 条件必填 | `required-if` | `v:"required-if:pushMethod,2"` | 另一字段等于该值时必填 |
| 长度 | `length` | `v:"length:1,50"` | 仅字符串；数组长度用 `arrLen` |
| 范围 | `between` | `v:"between:1,300"` | 含边界 |
| 范围 | `min` / `max` | `v:"min:0"` | 数值大小或长度 |
| 类型 | `integer` / `float` | `v:"required|integer"` | 整数用 `integer`；`numeric` 会放行数字字符串，少用 |
| 类型 | `array` / `json` | `v:"required|array"` | 数组、JSON 字符串 |
| 日期 | `date` / `date-format` | `v:"date-format:Y-m-d"` | 日期与格式 |
| 枚举 | `in` / `not-in` | `v:"required|in:1,2,3"` | 状态、类型等有限取值 |
| 格式 | `regex` | `v:"regex:^\\d{6}$"` | Go 正则，反斜杠写两次 |
| 控制 | `bail` | `v:"bail|required|length:1,50"` | 首条失败即停止 |

- `length` 只用于字符串；数字字段用 `min`/`max`/`between`/`in`。
- 不加 `bail` 时所有规则都会执行，`message` 是多条错误用 `;` 拼接。

### 8.2 项目自定义

- 目前没有自定义校验规则。新增时集中放在服务内的 `validate` 包（按需新建），`init()` 里 `gvalid.RegisterRule` 注册；规则名全局唯一，报错文案用中文并指出字段与实际值。

## 9 外部回调的信任边界

- 无鉴权 ≠ 来源可信：先核实对端协议、入口中间件与部署边界，再定认证方案。
- 认证在写入、入队、ACK 之前完成；失败即拒绝，不降级放行。
- 不凭单个请求头或 `X-Forwarded-For` 信任来源；依赖网关时确认外部无法绕过直连。
- 认证、防重放、幂等分别处理：签名正确 ≠ 未重放。
- 受理成功才返回 ACK；入队或落库失败不伪装成功。

## 10 检查与报告

1. **圈定范围**：用户点名的文件或模块；没点名则只看本次 diff 里 `api/**/v1/*.go` 与对应 controller。
2. **逐个对照**：每个 Req 连同 Res、`g.Meta`、校验标签、controller 方法一起看；同分组内查 `path` 是否重复。
3. **给结论**：每个 Req 归入「通过 / 必须改（阻塞合并） / 建议改（记录即可）」，写明文件、结构体、违反的规则与应改成什么。
4. **只改必要范围**：只动 Req/Res 与手写 controller；api 接口文件由生成器维护，改完执行 `make ctrl`。
