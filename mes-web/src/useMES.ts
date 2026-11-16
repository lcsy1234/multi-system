import { useCallback, useEffect, useRef, useState } from "react";
import { api } from "./api";
import type { Catalog, Order, Requisition } from "./types";

export function useMES() {
  const [catalog, setCatalog] = useState<Catalog | null>(null);
  const [orders, setOrders] = useState<Order[]>([]);
  const [records, setRecords] = useState<Requisition[]>([]);
  const [loading, setLoading] = useState(false);
  const [ordersLoading, setOrdersLoading] = useState(false);
  const [catalogError, setCatalogError] = useState("");
  const [ordersError, setOrdersError] = useState("");
  const [recordsError, setRecordsError] = useState("");
  const busy = useRef(false);
  const refresh = useCallback(async (signal?: AbortSignal) => {
    if (busy.current) return;
    busy.current = true;
    setLoading(true);
    try {
      const result = await api.records(signal);
      if (!signal?.aborted) {
        setRecords(result.records ?? []);
        setRecordsError("");
      }
    } catch (cause) {
      if (!signal?.aborted)
        setRecordsError(
          cause instanceof Error ? cause.message : "读取申请失败",
        );
    } finally {
      busy.current = false;
      if (!signal?.aborted) setLoading(false);
    }
  }, []);
  const loadCatalog = useCallback(async (signal?: AbortSignal) => {
    try {
      const result = await api.catalog(signal);
      if (!signal?.aborted) {
        setCatalog(result);
        setCatalogError("");
      }
    } catch (cause) {
      if (!signal?.aborted)
        setCatalogError(
          cause instanceof Error ? cause.message : "加载基础数据失败",
        );
    }
  }, []);
  const loadOrders = useCallback(async (signal?: AbortSignal) => {
    setOrdersLoading(true);
    try {
      const result = await api.orders(signal);
      if (!signal?.aborted) {
        setOrders(result.orders ?? []);
        setOrdersError("");
      }
    } catch (cause) {
      if (!signal?.aborted)
        setOrdersError(
          cause instanceof Error ? cause.message : "读取订单失败",
        );
    } finally {
      if (!signal?.aborted) setOrdersLoading(false);
    }
  }, []);
  useEffect(() => {
    const controller = new AbortController();
    void loadCatalog(controller.signal);
    void loadOrders(controller.signal);
    void refresh(controller.signal);
    return () => controller.abort();
  }, [loadCatalog, loadOrders, refresh]);
  return {
    catalog,
    orders,
    ordersLoading,
    records,
    loading,
    error: catalogError || ordersError || recordsError,
    refresh,
    reload: async () => {
      await loadCatalog();
      await loadOrders();
      await refresh();
    },
  };
}
