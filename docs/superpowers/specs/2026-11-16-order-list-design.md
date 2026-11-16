# 订单列表设计

日期：2026-11-16

## 目标

在 MES 演示系统「生产管理」下新增「订单列表」：列表走真实前后端接口；详情、编辑先做独立占位页，便于后续补全。

## 范围

### 做

- 侧栏「生产管理」增加「订单列表」菜单项
- 独立订单模块：内存种子数据 + `GET /api/v1/orders`
- 订单列表页（筛选 + 表格），操作列「详情」「编辑」
- 详情页、编辑页占位（展示订单编码、说明、返回列表）

### 不做

- 订单详情 / 更新 API
- 订单与产程强制关联
- MySQL 持久化订单
- 列表多选框
- 订单状态 Tabs（截图无状态列；列表默认展示全部）

## 信息架构与页面流转

页面状态（沿用现有 `App.tsx` 的 `page` 状态机，不引入路由库）：

| page | 含义 |
|------|------|
| `orders` | 订单列表 |
| `order-detail` | 详情占位 |
| `order-edit` | 编辑占位 |

流转：

1. 侧栏点击「订单列表」→ `orders`
2. 列表「详情」→ `order-detail`（携带当前订单 `id`）
3. 列表「编辑」→ `order-edit`（携带当前订单 `id`）
4. 占位页「返回列表」→ `orders`
5. 在详情/编辑时，侧栏选中项仍为 `orders`

## 数据模型

```go
type Order struct {
    ID           string `json:"id"`           // 订单编码，如 SCDD26111320060047
    FactoryName  string `json:"factoryName"`  // 江苏厂
    FactoryCode  string `json:"factoryCode"`  // GC2006
    LineName     string `json:"lineName"`     // 江苏混合二线
    LineCode     string `json:"lineCode"`     // GC2006H02
    ProductName  string `json:"productName"`  // 元气森林-无糖气泡水-白桃 PET480ml*15 纸箱装
    ProductCode  string `json:"productCode"`  // YQ03002
    PlanDate     string `json:"planDate"`     // 2026-11-13 00:00
    PlanQuantity int    `json:"planQuantity"` // 1190
}
```

种子数据约 5～7 条，订单号风格 `SCDD…`，工厂/产线/产品与参考截图一致（江苏厂、天津厂等）。

## API

- `GET /api/v1/orders`
- 响应 `data`：`{ "orders": Order[] }`
- 包装格式与现有接口一致：`{ code, message, data }`（GoFrame `MiddlewareHandlerResponse`）
- 本阶段无查询参数；筛选仅在前端完成

## 后端结构

独立 order 模块（对齐 health / material 的 GoFrame 分层）：

- `mes/internal/model/order.go` — `Order` 类型 + `Orders` 种子
- `mes/api/order/v1/order.go` — `ListReq` / `ListRes`
- `mes/api/order/order.go` — `IOrderV1`
- `mes/internal/controller/order/` — `List` 返回 `model.Orders`
- `mes/internal/cmd/cmd.go` — `group.Bind` 注册 order controller

## 前端结构

- `types.ts`：`Order` 类型；`Page` 扩展为含 `orders` | `order-detail` | `order-edit`
- `api.ts`：`api.orders()`
- `components/OrderList.tsx`：筛选（关键词：订单编码/产品；工厂）+ 表格
- `components/OrderPlaceholder.tsx`：详情/编辑共用占位（`mode: "detail" | "edit"`）
- `App.tsx`：侧栏菜单、页面切换、标题/面包屑
- `useMES.ts`：加载订单列表（或在 App / 独立 hook 中拉取；与 catalog/records 并列）

### 列表列

| 列 | 渲染 |
|----|------|
| 订单编码 | `id` |
| 工厂 | 名称 + 次行编码 |
| 产线 | 名称 + 次行编码 |
| 生产产品项 | 名称 + 次行产品编码 |
| 计划开工日期 | `planDate` |
| 计划生产数量 | `planQuantity` |
| 操作 | 详情、编辑 |

视觉对齐现有产程列表（主文案 + `.secondary` 次行）。

### 占位页

- 标题：订单详情 / 订单编辑
- 展示当前订单编码
- 文案：详情页待完善 / 编辑页待完善
- 按钮：返回列表
- 不请求详情/更新接口

## 错误与空态

- 列表加载失败：沿用页面级 `Alert` + 重试
- 无匹配数据：`Empty`

## 测试

- 手动：侧栏进入列表、筛选、进详情/编辑再返回
- 可选：对 `GET /api/v1/orders` 做一次 curl / 现有风格的简单断言
- 本阶段不强制新增自动化测试文件

## 成功标准

1. 侧栏可见「订单列表」并可进入
2. 列表数据来自 `GET /api/v1/orders`，列含订单编码、工厂、产线等约定字段
3. 详情、编辑可进入占位页并返回列表
4. 不破坏现有产程 / 领料流程
