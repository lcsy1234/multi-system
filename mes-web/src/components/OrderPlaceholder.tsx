import { Button, Card, Result, Space, Typography } from "antd";

interface Props {
  mode: "detail" | "edit";
  orderId: string;
  onBack: () => void;
}

export function OrderPlaceholder({ mode, orderId, onBack }: Props) {
  const isDetail = mode === "detail";
  return (
    <Card className="data-card">
      <Result
        status="info"
        title={isDetail ? "订单详情" : "订单编辑"}
        subTitle={
          <Space direction="vertical" size={4}>
            <Typography.Text>
              订单编码：<span className="code">{orderId}</span>
            </Typography.Text>
            <Typography.Text type="secondary">
              {isDetail ? "详情页待完善" : "编辑页待完善"}
            </Typography.Text>
          </Space>
        }
        extra={
          <Button type="primary" onClick={onBack}>
            返回列表
          </Button>
        }
      />
    </Card>
  );
}
