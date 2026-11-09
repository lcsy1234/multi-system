import { useState } from "react";
import {
  Button,
  Input,
  InputNumber,
  Select,
  Space,
  Table,
  Tabs,
  Typography,
} from "antd";
import type { TableColumnsType } from "antd";
import { DeleteOutlined, PlusOutlined } from "@ant-design/icons";
import type { DraftLine, Material } from "../types";

interface Props {
  materials: Material[];
  lines: DraftLine[];
  onChange: (lines: DraftLine[]) => void;
  disabled?: boolean;
}
export function MaterialsEditor({
  materials,
  lines,
  onChange,
  disabled = false,
}: Props) {
  const [category, setCategory] = useState("全部");
  const [selected, setSelected] = useState<React.Key[]>([]);
  const update = (code: string, patch: Partial<DraftLine>) =>
    onChange(
      lines.map((line) => (line.code === code ? { ...line, ...patch } : line)),
    );
  const remove = (codes: React.Key[]) => {
    onChange(lines.filter((line) => !codes.includes(line.code)));
    setSelected([]);
  };
  const columns: TableColumnsType<DraftLine> = [
    {
      title: "物料",
      dataIndex: "name",
      width: 160,
      render: (name: string, row) => (
        <>
          <div>{name}</div>
          <Typography.Text type="secondary" className="small-text">
            {row.code}
          </Typography.Text>
        </>
      ),
    },
    { title: "规格", dataIndex: "spec", width: 180 },
    { title: "单位", dataIndex: "unit", width: 60 },
    {
      title: "申请数量",
      width: 150,
      render: (_, row) => (
        <InputNumber<string>
          aria-label={`${row.name}申请数量`}
          disabled={disabled}
          stringMode
          value={row.quantity || null}
          min="0.0001"
          max="999999999.9999"
          step="0.0001"
          precision={4}
          onChange={(value) => update(row.code, { quantity: value ?? "" })}
          className="quantity-input"
        />
      ),
    },
    {
      title: "备注",
      width: 160,
      render: (_, row) => (
        <Input
          disabled={disabled}
          aria-label={`${row.name}备注`}
          value={row.remark}
          maxLength={100}
          placeholder="选填"
          onChange={(event) => update(row.code, { remark: event.target.value })}
        />
      ),
    },
    {
      title: "操作",
      width: 68,
      render: (_, row) => (
        <Button
          type="link"
          danger
          disabled={disabled}
          size="small"
          onClick={() => remove([row.code])}
        >
          删除
        </Button>
      ),
    },
  ];
  const missing = materials.filter(
    (material) => !lines.some((line) => line.code === material.code),
  );
  return (
    <section>
      <div className="section-heading">
        <h3>物料明细</h3>
        <Space wrap>
          <Select
            aria-label="添加物料"
            className="add-material"
            value={undefined}
            placeholder={
              <>
                <PlusOutlined /> 添加物料
              </>
            }
            disabled={disabled || !missing.length}
            options={missing.map((material) => ({
              value: material.code,
              label: material.name,
            }))}
            onChange={(code) => {
              const material = materials.find((item) => item.code === code);
              if (material)
                onChange([
                  ...lines,
                  { ...material, quantity: material.suggested, remark: "" },
                ]);
            }}
          />
          <Button
            icon={<DeleteOutlined />}
            danger
            disabled={disabled || !selected.length}
            onClick={() => remove(selected)}
          >
            删除所选
          </Button>
        </Space>
      </div>
      <Tabs
        activeKey={category}
        onChange={setCategory}
        items={["全部", "生产原料", "生产包材", "品保用料"].map((label) => ({
          key: label,
          label,
        }))}
      />
      <Table<DraftLine>
        size="small"
        rowKey="code"
        columns={columns}
        dataSource={lines.filter(
          (line) => category === "全部" || line.category === category,
        )}
        pagination={false}
        scroll={{ x: 800, y: 280 }}
        locale={{ emptyText: "当前类别没有物料，可添加物料后继续" }}
        rowSelection={{
          selectedRowKeys: selected,
          preserveSelectedRowKeys: true,
          onChange: setSelected,
          getCheckboxProps: () => ({ disabled }),
        }}
      />
      <p className="field-hint">
        基础物料为演示数据，未连接仓库库存。数量按各自单位填写。
      </p>
    </section>
  );
}
