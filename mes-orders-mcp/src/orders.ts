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

export interface OrderFilters {
  keyword?: string;
  factory?: string;
}

interface ApiEnvelope {
  code: number;
  message: string;
  data?: { orders?: Order[] };
}

export function mesBaseUrl(): string {
  return (process.env.MES_BASE_URL ?? "http://127.0.0.1:8004").replace(
    /\/$/,
    "",
  );
}

export async function fetchOrders(baseUrl = mesBaseUrl()): Promise<Order[]> {
  const url = `${baseUrl}/api/v1/orders`;
  let response: Response;
  try {
    response = await fetch(url);
  } catch (cause) {
    const detail = cause instanceof Error ? cause.message : String(cause);
    throw new Error(
      `无法连接 MES（${url}）：${detail}。请先启动 MES（例如 make demo 或 .bin/mes），并检查 MES_BASE_URL。`,
    );
  }
  let body: ApiEnvelope;
  try {
    body = (await response.json()) as ApiEnvelope;
  } catch {
    throw new Error(
      `MES 返回非 JSON（HTTP ${response.status}）：${url}`,
    );
  }
  if (!response.ok || body.code !== 0) {
    throw new Error(
      body.message ||
        `读取订单失败（HTTP ${response.status}, code=${body.code}）`,
    );
  }
  return body.data?.orders ?? [];
}

export function filterOrders(
  orders: Order[],
  filters: OrderFilters = {},
): Order[] {
  const keyword = filters.keyword?.trim().toLowerCase() ?? "";
  const factory = filters.factory?.trim() ?? "";
  return orders.filter((row) => {
    if (factory && row.factoryName !== factory) return false;
    if (!keyword) return true;
    return [row.id, row.productName, row.productCode].some((value) =>
      value.toLowerCase().includes(keyword),
    );
  });
}
