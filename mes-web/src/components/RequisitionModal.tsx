import { useState } from "react";
import {
  Alert,
  App,
  Button,
  Col,
  DatePicker,
  Form,
  Input,
  Modal,
  Row,
  Select,
  Space,
  Typography,
} from "antd";
import dayjs from "dayjs";
import type { Dayjs } from "dayjs";
import { api } from "../api";
import type {
  Catalog,
  DraftLine,
  ProductionProcess,
  Requisition,
} from "../types";
import { MaterialsEditor } from "./MaterialsEditor";

interface Props {
  process: ProductionProcess;
  catalog: Catalog;
  onClose: () => void;
  onSubmitted: (record: Requisition) => void;
}
interface Values {
  warehouseId: string;
  date: Dayjs;
  remark?: string;
}
export function RequisitionModal({
  process,
  catalog,
  onClose,
  onSubmitted,
}: Props) {
  const [form] = Form.useForm<Values>();
  const { message } = App.useApp();
  const [requestId] = useState(() => crypto.randomUUID());
  const [lines, setLines] = useState<DraftLine[]>(() =>
    catalog.materials.map((material) => ({
      ...material,
      quantity: material.suggested,
      remark: "",
    })),
  );
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");
  const submit = async (values: Values) => {
    if (submitting) return;
    if (
      !lines.length ||
      lines.some(
        (line) =>
          !/^(0|[1-9][0-9]{0,8})(\.[0-9]{1,4})?$/.test(line.quantity) ||
          Number(line.quantity) <= 0,
      )
    ) {
      setError("请至少添加一种物料，数量必须大于零且最多四位小数。");
      return;
    }
    setError("");
    setSubmitting(true);
    try {
      const { record } = await api.submit({
        clientRequestId: requestId,
        processId: process.id,
        warehouseId: values.warehouseId,
        date: values.date.format("YYYY-MM-DD"),
        remark: values.remark ?? "",
        items: lines.map((line) => ({
          materialCode: line.code,
          quantity: line.quantity,
          remark: line.remark,
        })),
      });
      if (record.status === "enqueue_failed")
        void message.warning("申请已保存，但任务入队失败，请查看详情。");
      else void message.success("领料申请已提交，任务已入队。");
      onSubmitted(record);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "提交失败，请重试");
    } finally {
      setSubmitting(false);
    }
  };
  const today = new Intl.DateTimeFormat("sv-SE", {
    timeZone: "Asia/Shanghai",
  }).format(new Date());
  return (
    <Modal
      open
      title="领料申请"
      width={1120}
      onCancel={onClose}
      closable={!submitting}
      keyboard={!submitting}
      mask={{ closable: !submitting }}
      className="requisition-modal"
      footer={
        <div className="modal-footer">
          <Typography.Text type="secondary">
            本次申请：<strong>{lines.length}</strong> 种物料
          </Typography.Text>
          <Space>
            <Button disabled={submitting} onClick={onClose}>
              取消
            </Button>
            <Button
              type="primary"
              loading={submitting}
              onClick={() => form.submit()}
            >
              提交领料申请
            </Button>
          </Space>
        </div>
      }
    >
      <p className="modal-description">
        根据生产产程填写本次物料需求，提交后进入异步任务队列。
      </p>
      <Form
        form={form}
        layout="vertical"
        requiredMark
        initialValues={{ date: dayjs(today), remark: "" }}
        onFinish={submit}
        disabled={submitting}
      >
        <h3 className="section-title">基本信息</h3>
        <Row gutter={[24, 0]}>
          <Col xs={24} md={8}>
            <Form.Item label="产程编号">
              <Input aria-label="产程编号" readOnly value={process.id} />
            </Form.Item>
          </Col>
          <Col xs={24} md={8}>
            <Form.Item label="生产品项">
              <Input aria-label="生产品项" readOnly value={process.product} />
            </Form.Item>
          </Col>
          <Col xs={24} md={8}>
            <Form.Item label="工厂">
              <Input aria-label="工厂" readOnly value={process.factory} />
            </Form.Item>
          </Col>
          <Col xs={24} md={8}>
            <Form.Item label="产线">
              <Input aria-label="产线" readOnly value={process.line} />
            </Form.Item>
          </Col>
          <Col xs={24} md={8}>
            <Form.Item
              label="领料仓库"
              name="warehouseId"
              rules={[{ required: true, message: "请选择领料仓库" }]}
            >
              <Select
                placeholder="请选择仓库"
                options={catalog.warehouses.map((warehouse) => ({
                  value: warehouse.id,
                  label: warehouse.name,
                }))}
              />
            </Form.Item>
          </Col>
          <Col xs={24} md={8}>
            <Form.Item
              label="申请日期"
              name="date"
              rules={[{ required: true, message: "请选择申请日期" }]}
            >
              <DatePicker className="full-width" allowClear={false} />
            </Form.Item>
          </Col>
          <Col span={24}>
            <Form.Item label="备注" name="remark">
              <Input
                maxLength={300}
                showCount
                placeholder="填写领料用途或其他说明（选填）"
              />
            </Form.Item>
          </Col>
        </Row>
      </Form>
      <fieldset className="materials-fieldset" disabled={submitting}>
        <MaterialsEditor
          disabled={submitting}
          materials={catalog.materials}
          lines={lines}
          onChange={setLines}
        />
      </fieldset>
      {error && (
        <Alert title={error} type="error" showIcon className="form-alert" />
      )}
    </Modal>
  );
}
