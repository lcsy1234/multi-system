export interface MaterialLine {
  materialCode: string;
  quantity: string;
  remark: string;
}
export interface Submission {
  clientRequestId: string;
  processId: string;
  warehouseId: string;
  date: string;
  remark: string;
  items: MaterialLine[];
}
export interface Requisition extends Submission {
  id: string;
  number: string;
  status: string;
  lastError: string;
  attempts: number;
  createdAt: string;
  updatedAt: string;
  events: { time: string; message: string }[];
}
export interface ProductionProcess {
  id: string;
  order: string;
  factory: string;
  line: string;
  product: string;
  planDate: string;
  status: string;
}
export interface Material {
  code: string;
  name: string;
  spec: string;
  unit: string;
  category: string;
  suggested: string;
}
export interface Catalog {
  processes: ProductionProcess[];
  materials: Material[];
  warehouses: { id: string; name: string }[];
}
export interface DraftLine extends Material {
  quantity: string;
  remark: string;
}
export type Page = "processes" | "requests";
