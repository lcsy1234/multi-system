import { useState } from "react";
import {
  Button,
  Card,
  Empty,
  Form,
  Input,
  Select,
  Space,
  Table,
  Typography,
} from "antd";
import type { TableColumnsType } from "antd";
import type { Order } from "../types";

interface Filters {
  keyword?: string;
  factory?: string;
}

interface Props {
  orders: Order[];
  loading: boolean;
  onDetail: (order: Order) => void;
  onEdit: (order: Order) => void;
}

export function OrderList({ orders, loading, onDetail, onEdit }: Props) {
  const [filters, setFilters] = useState<Filters>({});
  const [filterForm] = Form.useForm<Filters>();
  const query = filters.keyword?.trim().toLowerCase() ?? "";
  const filtered = orders.filter(
    (row) =>
      (!filters.factory || row.factoryName === filters.factory) &&
      (!query ||
        [row.id, row.productName, row.productCode].some((value) =>
          value.toLowerCase().includes(query),
        )),
  );
  const columns: TableColumnsType<Order> = [
    {
      title: "订单编码",
      dataIndex: "id",
      width: 200,
      render: (id: string) => <span className="code">{id}</span>,
    },
    {
      title: "工厂",
      dataIndex: "factoryName",
      width: 140,
      render: (name: string, row) => (
        <>
          {name}
          <div className="secondary">{row.factoryCode}</div>
        </>
      ),
    },
    {
      title: "产线",
      dataIndex: "lineName",
      width: 170,
      render: (name: string, row) => (
        <>
          {name}
          <div className="secondary">{row.lineCode}</div>
        </>
      ),
    },
    {
      title: "生产产品项",
      dataIndex: "productName",
      width: 280,
      render: (name: string, row) => (
        <>
          {name}
          <div className="secondary">{row.productCode}</div>
        </>
      ),
    },
    { title: "计划开工日期", dataIndex: "planDate", width: 150 },
    {
      title: "计划生产数量",
      dataIndex: "planQuantity",
      width: 120,
      render: (value: number) => value.toLocaleString(),
    },
    {
      title: "操作",
      width: 120,
      fixed: "right",
      render: (_, row) => (
        <Space>
          <Button type="link" size="small" onClick={() => onDetail(row)}>
            详情
          </Button>
          <Button type="link" size="small" onClick={() => onEdit(row)}>
            编辑
          </Button>
        </Space>
      ),
    },
  ];
  return (
    <>
      <Card className="filter-card">
        <Form
          form={filterForm}
          layout="vertical"
          className="filter-form"
          onFinish={setFilters}
        >
          <Form.Item label="关键词" name="keyword">
            <Input placeholder="搜索订单编码或产品" allowClear />
          </Form.Item>
          <Form.Item label="工厂" name="factory">
            <Select
              allowClear
              placeholder="全部工厂"
              options={[...new Set(orders.map((row) => row.factoryName))].map(
                (factory) => ({ value: factory, label: factory }),
              )}
            />
          </Form.Item>
          <Space className="filter-actions">
            <Button type="primary" htmlType="submit">
              查询
            </Button>
            <Button
              onClick={() => {
                filterForm.resetFields();
                setFilters({});
              }}
            >
              重置
            </Button>
          </Space>
        </Form>
      </Card>
      <Card className="data-card">
        <div className="table-toolbar">
          <div />
          <div className="table-tools">
            <Typography.Text type="secondary">
              共 {filtered.length} 条
            </Typography.Text>
          </div>
        </div>
        <Table<Order>
          rowKey="id"
          columns={columns}
          dataSource={filtered}
          loading={loading}
          scroll={{ x: 1180 }}
          pagination={{ pageSize: 10, hideOnSinglePage: true }}
          locale={{
            emptyText: <Empty description="暂无符合条件的订单" />,
          }}
        />
        <div className="table-caption">演示订单数据 · 详情与编辑页待完善</div>
      </Card>
    </>
  );
}
