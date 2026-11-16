# MES Orders MCP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship a TypeScript stdio MCP server that exposes `list_orders` against live MES `/api/v1/orders`, usable from Cursor and Codex.

**Architecture:** Independent package `mes-orders-mcp` using `@modelcontextprotocol/sdk` + zod. Tool fetches orders over HTTP, filters by keyword/factory, returns JSON text. Project `.cursor/mcp.json` registers the server via `${workspaceFolder}`.

**Tech Stack:** Node 20+, TypeScript, `@modelcontextprotocol/sdk`, zod

## Global Constraints

- Only tool: `list_orders` (no detail/edit)
- Call real MES API only (no seed fallback)
- Default `MES_BASE_URL=http://127.0.0.1:8004`
- Never log to stdout (use stderr)
- Chinese README for learning

---

### Task 1: Scaffold package + fetch/filter + MCP server

**Files:**
- Create: `mes-orders-mcp/package.json`
- Create: `mes-orders-mcp/tsconfig.json`
- Create: `mes-orders-mcp/src/orders.ts`
- Create: `mes-orders-mcp/src/index.ts`
- Create: `mes-orders-mcp/README.md`
- Create: `.cursor/mcp.json`

- [ ] **Step 1: Create package.json and tsconfig**

`package.json`: name `mes-orders-mcp`, type module, scripts build/start, deps `@modelcontextprotocol/sdk` + `zod`, devDeps `typescript` `@types/node`.

`tsconfig.json`: target ES2022, module Node16, outDir dist, rootDir src, strict true.

- [ ] **Step 2: Implement orders.ts**

Export `Order` type matching MES JSON. `fetchOrders(baseUrl)` hits `/api/v1/orders`, unwraps `{ code, message, data: { orders } }`. `filterOrders(orders, { keyword?, factory? })`.

- [ ] **Step 3: Implement index.ts MCP server**

`McpServer` name `mes-orders`, register `list_orders` with zod optional keyword/factory, connect `StdioServerTransport`. On error return `isError: true`.

- [ ] **Step 4: npm install && npm run build**

- [ ] **Step 5: README + .cursor/mcp.json**

README: what MCP is, how to build, start MES, Cursor/Codex config. `.cursor/mcp.json` with command node and args `${workspaceFolder}/mes-orders-mcp/dist/index.js`.

- [ ] **Step 6: Smoke test**

With MES up: use a tiny node script or mcp inspector / echo via calling filter path. At minimum: `node -e` importing built orders module against live API, or run server and verify build + curl MES separately.

- [ ] **Step 7: Commit**

```bash
git add mes-orders-mcp .cursor/mcp.json docs/superpowers/specs/2026-11-16-mes-orders-mcp-design.md docs/superpowers/plans/2026-11-16-mes-orders-mcp.md
git commit -m "feat: add mes-orders MCP server for list_orders"
```
