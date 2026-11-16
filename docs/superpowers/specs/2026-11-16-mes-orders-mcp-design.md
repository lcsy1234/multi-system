# MES 订单列表 MCP 设计

日期：2026-11-16

## 目标

用 TypeScript 实现一个本地 stdio MCP Server，暴露 `list_orders` 工具，供 Cursor / Codex 等 AI 客户端调用；数据来自真实 MES `GET /api/v1/orders`。

## 范围

### 做

- 独立包 `mes-orders-mcp/`（Node + TypeScript + `@modelcontextprotocol/sdk`）
- 工具 `list_orders`（可选 `keyword`、`factory`）
- 通过 HTTP 请求 MES API
- README：本机安装、Cursor / Codex 配置示例、学习说明

### 不做

- 详情 / 编辑工具
- 内置种子回退
- HTTP/SSE 远程传输
- 鉴权

## 架构

```text
AI Client (Cursor / Codex)
  → spawn: node dist/index.js  (stdio JSON-RPC)
  → mes-orders-mcp
  → GET {MES_BASE_URL}/api/v1/orders
  → 过滤后返回 JSON 文本
```

## 工具约定

**名称：** `list_orders`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `keyword` | string | 否 | 匹配 `id` / `productName` / `productCode`（不区分大小写包含） |
| `factory` | string | 否 | 按 `factoryName` 精确匹配 |

**成功：** `content: [{ type: "text", text: JSON.stringify({ count, orders }, null, 2) }]`

**失败：** `isError: true`，文案说明 MES 未启动或请求失败（提示检查 `MES_BASE_URL`，默认 `http://127.0.0.1:8004`）。

## 目录结构

```text
mes-orders-mcp/
  package.json
  tsconfig.json
  src/
    index.ts      # MCP 入口：注册工具 + stdio connect
    orders.ts     # fetch + filter
  README.md
```

依赖：`@modelcontextprotocol/sdk`、`zod`；开发依赖 `typescript`、`@types/node`。

脚本：`build` → `tsc`；`start` → `node dist/index.js`。

## 客户端配置

### Cursor

项目或用户 MCP 配置（示例）：

```json
{
  "mcpServers": {
    "mes-orders": {
      "command": "node",
      "args": ["/ABS/PATH/multi-system/mes-orders-mcp/dist/index.js"],
      "env": {
        "MES_BASE_URL": "http://127.0.0.1:8004"
      }
    }
  }
}
```

也可在仓库放置 `.cursor/mcp.json`（路径用本机绝对路径或文档中说明需替换）。

### Codex

在 Codex MCP 配置中注册同名 stdio server（`command` + `args` + `env` 同上）。以当前 Codex 文档中的 MCP 配置项为准；README 给出通用 stdio 示例。

## 学习要点（README 简述）

1. MCP Server 通过 stdio 与客户端通信，不要往 stdout 打普通日志（会污染协议；日志用 stderr）。
2. Tool = AI 可调用的函数；inputSchema 用 zod 描述。
3. 本示例把业务 API 包成 Tool，AI 无需知道 HTTP 细节。

## 成功标准

1. `npm run build` 成功
2. MES 运行时，客户端能列出并调用 `list_orders` 得到订单
3. MES 未运行时，返回可读错误
4. Cursor 与 Codex 均可按 README 配置调用
