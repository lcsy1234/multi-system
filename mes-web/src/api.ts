import type { Catalog, Order, Requisition, Submission } from "./types";


async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const response = await fetch(path, options);
  const result: { code: number; message: string; data: T } =
    await response.json();
  // GoFrame business errors may also use HTTP 200.
  if (!response.ok || result.code !== 0)
    throw new Error(result.message || `请求失败（${response.status}）`);
  return result.data;
}
export const api = {
  catalog: (signal?: AbortSignal) =>
    request<Catalog>("/api/v1/catalog", { signal }),
  orders: (signal?: AbortSignal) =>
    request<{ orders: Order[] }>("/api/v1/orders", { signal }),
  records: (signal?: AbortSignal) =>
    request<{ records: Requisition[] }>("/api/v1/material-requisitions", {
      signal,
    }),
  submit: (submission: Submission) =>
    request<{ record: Requisition }>("/api/v1/material-requisitions", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(submission),
    }),
};
