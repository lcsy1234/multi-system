#!/usr/bin/env node
import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { z } from "zod";
import { fetchOrders, filterOrders, mesBaseUrl } from "./orders.js";

const server = new McpServer({
  name: "mes-orders",
  version: "0.1.0",
});

server.registerTool(
  "list_orders",
  {
    title: "List MES orders",
    description:
      "从 MES 拉取生产订单列表。可选 keyword（订单编码/产品名/产品编码）与 factory（工厂名称，如江苏厂）过滤。",
    inputSchema: {
      keyword: z
        .string()
        .optional()
        .describe("搜索订单编码、产品名称或产品编码"),
      factory: z
        .string()
        .optional()
        .describe("按工厂名称精确过滤，例如：江苏厂"),
    },
  },
  async ({ keyword, factory }) => {
    try {
      const all = await fetchOrders();
      const orders = filterOrders(all, { keyword, factory });
      const payload = {
        baseUrl: mesBaseUrl(),
        count: orders.length,
        orders,
      };
      return {
        content: [
          {
            type: "text" as const,
            text: JSON.stringify(payload, null, 2),
          },
        ],
      };
    } catch (cause) {
      const message = cause instanceof Error ? cause.message : String(cause);
      return {
        isError: true,
        content: [{ type: "text" as const, text: message }],
      };
    }
  },
);

async function main() {
  const transport = new StdioServerTransport();
  await server.connect(transport);
}

main().catch((error) => {
  console.error("mes-orders-mcp failed to start:", error);
  process.exit(1);
});
