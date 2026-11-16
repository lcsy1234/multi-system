# mes-orders-mcp

用 **MCP（Model Context Protocol）** 把 MES 订单列表暴露给 Cursor / Codex 等 AI 客户端。

## 你在学什么

```text
AI（Cursor / Codex）
  │  stdio：JSON-RPC（不要往 stdout 打普通日志）
  ▼
本 MCP Server（Node + TypeScript）
  │  HTTP
  ▼
MES  GET /api/v1/orders
```

- **MCP Server**：被客户端拉起的本地进程，提供 Tools
- **Tool**：AI 可调用的函数；本仓库只有 `list_orders`
- **stdio**：客户端通过标准输入输出和 Server 通信（本地最常见）

## 前置条件

1. Node.js 20+
2. MES 已启动（默认 `http://127.0.0.1:8004`）

```bash
# 在 multi-system 根目录
make demo
# 或仅启动已编译的 MES：
# REDIS_ADDR=127.0.0.1:6380 .bin/mes
```

## 安装与构建

```bash
cd mes-orders-mcp
npm install
npm run build
```

## 工具：list_orders

| 参数 | 说明 |
|------|------|
| `keyword`（可选） | 匹配订单编码 / 产品名 / 产品编码 |
| `factory`（可选） | 工厂名称精确匹配，如 `江苏厂` |

环境变量：

| 变量 | 默认 | 说明 |
|------|------|------|
| `MES_BASE_URL` | `http://127.0.0.1:8004` | MES 根地址 |

## 在 Cursor 里用

仓库已提供 `.cursor/mcp.json`。构建完成后：

1. 打开 Cursor Settings → MCP，确认 `mes-orders` 已启用
2. 若未出现：Reload Window，或检查 `mes-orders-mcp/dist/index.js` 是否存在
3. 在 Agent 对话里说：「用 list_orders 查一下江苏厂的订单」

配置要点（已写入项目 `.cursor/mcp.json`）：

```json
{
  "mcpServers": {
    "mes-orders": {
      "command": "node",
      "args": ["${workspaceFolder}/mes-orders-mcp/dist/index.js"],
      "env": {
        "MES_BASE_URL": "http://127.0.0.1:8004"
      }
    }
  }
}
```

## 在 Codex 里用

在 Codex 的 MCP 配置中增加同名 **stdio** server（字段名以你本机 Codex 版本文档为准），核心是：

- `command`: `node`
- `args`: 指向本机 `mes-orders-mcp/dist/index.js` 的绝对路径
- `env.MES_BASE_URL`: `http://127.0.0.1:8004`

配置后重启 Codex，在对话中让它调用 `list_orders`。

## 本地冒烟（不经过 AI）

```bash
# MES 需已启动
node --input-type=module -e "
import { fetchOrders, filterOrders } from './dist/orders.js';
const all = await fetchOrders();
console.error('count', filterOrders(all, { factory: '江苏厂' }).length);
"
```

注意：MCP 进程的业务日志只能打 **stderr**；stdout 留给协议。

## 常见问题

- **工具报错无法连接**：先 `curl http://127.0.0.1:8004/api/v1/orders`，确认 MES 在跑
- **Cursor 看不到工具**：先 `npm run build`，再重载窗口
- **改了 TypeScript**：重新 `npm run build` 后再试
