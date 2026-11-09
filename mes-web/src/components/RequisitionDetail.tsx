import { Alert, Descriptions, Drawer, Table, Timeline, Typography } from "antd";
import type { Catalog, Requisition } from "../types";
import { StatusTag } from "./StatusTag";

export function RequisitionDetail({
  record,
  catalog,
  onClose,
}: {
  record: Requisition | null;
  catalog: Catalog | null;
  onClose: () => void;
}) {
  return (
    <Drawer title="申请进展" size="large" open={!!record} onClose={onClose}>
      {record && (
        <>
          <Descriptions
            column={1}
            size="small"
            items={[
              {
                key: "number",
                label: "申请单号",
                children: (
                  <Typography.Text copyable className="break-code">
                    {record.number}
                  </Typography.Text>
                ),
              },
              { key: "process", label: "产程编号", children: record.processId },
              {
                key: "status",
                label: "任务状态",
                children: <StatusTag status={record.status} />,
              },
            ]}
          />
          {record.lastError && (
            <Alert
              title={record.lastError}
              type="error"
              showIcon
              className="form-alert"
            />
          )}
          <h3 className="section-title detail-heading">申请物料</h3>
          <Table
            size="small"
            rowKey="materialCode"
            pagination={false}
            scroll={{ x: 420 }}
            dataSource={record.items}
            columns={[
              {
                title: "物料",
                dataIndex: "materialCode",
                render: (code: string) =>
                  catalog?.materials.find((material) => material.code === code)
                    ?.name ?? code,
              },
              { title: "申请数量", dataIndex: "quantity" },
              {
                title: "单位",
                dataIndex: "materialCode",
                key: "unit",
                render: (code: string) =>
                  catalog?.materials.find((material) => material.code === code)
                    ?.unit ?? "—",
              },
              {
                title: "备注",
                dataIndex: "remark",
                render: (value: string) => value || "—",
              },
            ]}
          />
          <h3 className="section-title detail-heading">处理时间线</h3>
          <Timeline
            items={record.events.map((event, index) => ({
              key: index,
              title: new Date(event.time).toLocaleString("zh-CN", {
                timeZone: "Asia/Shanghai",
                hour12: false,
              }),
              content: event.message,
            }))}
          />
          <Alert
            type="info"
            title="本阶段仅消费任务并打印日志，不调用 SCM，也不扣减仓库库存。"
          />
        </>
      )}
    </Drawer>
  );
}
