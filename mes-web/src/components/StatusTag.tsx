import { Tag } from "antd";
const states: Record<string, { label: string; color?: string }> = {
  queued: { label: "已排队", color: "processing" },
  consumed: { label: "已消费", color: "success" },
  enqueue_failed: { label: "入队失败", color: "error" },
  scm_created: { label: "历史记录" },
  已开工: { label: "已开工", color: "success" },
  变更中: { label: "变更中", color: "warning" },
  已完工: { label: "已完工" },
};
export function StatusTag({ status }: { status: string }) {
  const state = states[status] ?? { label: status };
  return <Tag color={state.color}>{state.label}</Tag>;
}
