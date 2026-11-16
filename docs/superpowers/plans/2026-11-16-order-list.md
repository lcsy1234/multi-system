# Order List Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add an Order List under Production Management with a real `GET /api/v1/orders` backend and frontend list; detail/edit are placeholder pages.

**Architecture:** Independent GoFrame `order` module with in-memory seed data (same pattern as catalog). Frontend extends the existing `page` state machine in `App.tsx`, adds `OrderList` and `OrderPlaceholder` components, and loads orders via `api.orders()`.

**Tech Stack:** Go / GoFrame, React + TypeScript + Ant Design, Vite

## Global Constraints

- No detail/update APIs in this phase
- No MySQL persistence for orders
- No row checkboxes; no status tabs
- Follow existing response wrapper `{ code, message, data }`
- Chinese UI copy; respond to user in simplified Chinese

---

### Task 1: Backend Order model + List API

**Files:**
- Create: `mes/internal/model/order.go`
- Create: `mes/api/order/v1/order.go`
- Create: `mes/api/order/order.go`
- Create: `mes/internal/controller/order/order.go`
- Create: `mes/internal/controller/order/order_new.go`
- Create: `mes/internal/controller/order/order_v1_list.go`
- Modify: `mes/internal/cmd/cmd.go`

**Interfaces:**
- Produces: `model.Order`, `model.Orders`, `GET /api/v1/orders` → `{ orders: Order[] }`

- [ ] **Step 1: Add Order model and seed data**

Create `mes/internal/model/order.go`:

```go
package model

type Order struct {
	ID           string `json:"id"`
	FactoryName  string `json:"factoryName"`
	FactoryCode  string `json:"factoryCode"`
	LineName     string `json:"lineName"`
	LineCode     string `json:"lineCode"`
	ProductName  string `json:"productName"`
	ProductCode  string `json:"productCode"`
	PlanDate     string `json:"planDate"`
	PlanQuantity int    `json:"planQuantity"`
}

var Orders = []Order{
	{"SCDD26111320060047", "江苏厂", "GC2006", "江苏混合二线", "GC2006H02", "元气森林-无糖气泡水-白桃 PET480ml*15 纸箱装", "YQ03002", "2026-11-13 00:00", 1190},
	{"SCDD26111320060048", "天津厂", "GC2001", "天津碳酸一线", "GC2001T01", "元气森林-无糖气泡水-白桃 PET480ml*15 纸箱装", "YQ03002", "2026-11-13 00:00", 10000},
	{"SCDD26111320060049", "江苏厂", "GC2006", "江苏无菌三线", "GC2006W03", "燃茶-乌龙茶-玄米乌龙 PET500ml*15 纸箱装", "YQ02003", "2026-11-13 00:00", 100000},
	{"SCDD26111220060050", "江苏厂", "GC2006", "江苏混合二线", "GC2006H02", "冰茶-红茶-冰红茶 PET500ml*15 纸箱装", "YQ04001", "2026-11-12 08:00", 5000},
	{"SCDD26111120010051", "天津厂", "GC2001", "天津无菌二线", "GC2001W02", "茉莉花茶 PET500ml*15 纸箱装", "YQ02005", "2026-11-11 00:00", 8000},
	{"SCDD26111020060052", "江苏厂", "GC2006", "江苏无菌三线", "GC2006W03", "青柠红茶 PET500ml*15 纸箱装", "YQ02008", "2026-11-10 00:00", 3200},
}
```

- [ ] **Step 2: Add API definition and controller**

`mes/api/order/v1/order.go`:

```go
package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	"multi-system/mes/internal/model"
)

type ListReq struct {
	g.Meta `path:"/api/v1/orders" method:"get" tags:"Order" summary:"List demo production orders"`
}
type ListRes struct {
	Orders []model.Order `json:"orders"`
}
```

`mes/api/order/order.go`:

```go
package order

import (
	"context"
	"multi-system/mes/api/order/v1"
)

type IOrderV1 interface {
	List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error)
}
```

`mes/internal/controller/order/order.go` — empty package file or leave as package comment only.

`mes/internal/controller/order/order_new.go`:

```go
package order

import "multi-system/mes/api/order"

type ControllerV1 struct{}

var _ order.IOrderV1 = (*ControllerV1)(nil)

func NewV1() *ControllerV1 { return &ControllerV1{} }
```

`mes/internal/controller/order/order_v1_list.go`:

```go
package order

import (
	"context"
	v1 "multi-system/mes/api/order/v1"
	"multi-system/mes/internal/model"
)

func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	return &v1.ListRes{Orders: model.Orders}, nil
}
```

- [ ] **Step 3: Register controller in cmd.go**

In `mes/internal/cmd/cmd.go`, import `order` controller and bind:

```go
group.Bind(health.NewV1(), materialController, order.NewV1())
```

- [ ] **Step 4: Verify API**

With MES server running:

```bash
curl -s http://127.0.0.1:8000/api/v1/orders | head -c 400
```

Expected: JSON with `code: 0` and `data.orders` array containing `SCDD…` ids.

- [ ] **Step 5: Commit**

```bash
git add mes/internal/model/order.go mes/api/order mes/internal/controller/order mes/internal/cmd/cmd.go
git commit -m "feat: add orders list API with demo seed data"
```

---

### Task 2: Frontend types, API, and order loading

**Files:**
- Modify: `mes-web/src/types.ts`
- Modify: `mes-web/src/api.ts`
- Modify: `mes-web/src/useMES.ts`

**Interfaces:**
- Consumes: `GET /api/v1/orders`
- Produces: `Order` type, `api.orders()`, `useMES().orders` / `ordersLoading` / reload includes orders

- [ ] **Step 1: Extend types**

In `types.ts`:

```ts
export interface Order {
  id: string;
  factoryName: string;
  factoryCode: string;
  lineName: string;
  lineCode: string;
  productName: string;
  productCode: string;
  planDate: string;
  planQuantity: number;
}

export type Page =
  | "processes"
  | "requests"
  | "orders"
  | "order-detail"
  | "order-edit";
```

- [ ] **Step 2: Add api.orders**

In `api.ts`:

```ts
orders: (signal?: AbortSignal) =>
  request<{ orders: Order[] }>("/api/v1/orders", { signal }),
```

Import `Order` in `api.ts`.

- [ ] **Step 3: Load orders in useMES**

Mirror `loadCatalog` / `refresh` pattern: state `orders`, `ordersError`, `loadOrders`, include in `reload` and initial `useEffect`. Return `orders` and fold `ordersError` into `error`.

- [ ] **Step 4: Commit**

```bash
git add mes-web/src/types.ts mes-web/src/api.ts mes-web/src/useMES.ts
git commit -m "feat: add order types and frontend orders fetch"
```

---

### Task 3: OrderList + OrderPlaceholder + App wiring

**Files:**
- Create: `mes-web/src/components/OrderList.tsx`
- Create: `mes-web/src/components/OrderPlaceholder.tsx`
- Modify: `mes-web/src/App.tsx`

**Interfaces:**
- Consumes: `Order[]`, navigate callbacks
- Produces: list UI with 详情/编辑; placeholder pages with 返回列表

- [ ] **Step 1: Create OrderList**

Props: `orders`, `loading`, `onDetail(order)`, `onEdit(order)`.

Filter card: keyword (id / productName / productCode), factory (unique `factoryName`). Table columns per spec; actions call `onDetail` / `onEdit`.

- [ ] **Step 2: Create OrderPlaceholder**

Props: `mode: "detail" | "edit"`, `orderId: string`, `onBack: () => void`.

Show title, order id, “待完善” message, 返回列表 button.

- [ ] **Step 3: Wire App.tsx**

- Menu child `{ key: "orders", label: "订单列表" }`
- `onClick` / `navigate` handle `orders`
- When `page` is `order-detail` | `order-edit`, keep menu selectedKeys `["orders"]`
- Title / eyebrow / breadcrumb for list vs placeholders
- Render `OrderList` or `OrderPlaceholder` instead of process/request tables when on order pages
- State: `selectedOrderId` for placeholder

- [ ] **Step 4: Manual verify**

Open MES web → 订单列表 → see SCDD rows → 详情 / 编辑 → 返回.

- [ ] **Step 5: Commit**

```bash
git add mes-web/src/components/OrderList.tsx mes-web/src/components/OrderPlaceholder.tsx mes-web/src/App.tsx
git commit -m "feat: add order list menu with detail/edit placeholders"
```

---

## Spec coverage check

| Spec item | Task |
|-----------|------|
| Sidebar 订单列表 | Task 3 |
| GET /api/v1/orders + seed | Task 1 |
| List columns (编码/工厂/产线/…) | Task 3 |
| Detail/edit placeholders | Task 3 |
| No detail/update API | (out of scope) |
| Frontend fetch | Task 2 |
