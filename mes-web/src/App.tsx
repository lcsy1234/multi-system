import { useEffect, useState } from "react";
import {
  Alert,
  Avatar,
  Badge,
  Breadcrumb,
  Button,
  Card,
  Drawer,
  Empty,
  Form,
  Grid,
  Input,
  Layout,
  Menu,
  Select,
  Space,
  Switch,
  Table,
  Tabs,
  Tag,
  Typography,
} from "antd";
import type { TableColumnsType } from "antd";
import {
  ApiOutlined,
  AppstoreOutlined,
  MenuOutlined,
  ReloadOutlined,
  TeamOutlined,
} from "@ant-design/icons";
import { useMES } from "./useMES";
import type { Page, ProductionProcess, Requisition } from "./types";
import { RequisitionModal } from "./components/RequisitionModal";
import { RequisitionDetail } from "./components/RequisitionDetail";
import { StatusTag } from "./components/StatusTag";

interface Filters {
  keyword?: string;
  factory?: string;
}
export default function App() {
  const { catalog, records, loading, error, refresh, reload } = useMES();
  const [page, setPage] = useState<Page>("processes");
  const [tab, setTab] = useState("已开工");
  const [filters, setFilters] = useState<Filters>({});
  const [process, setProcess] = useState<ProductionProcess | null>(null);
  const [detailId, setDetailId] = useState<string | null>(null);
  const [autoRefresh, setAutoRefresh] = useState(false);
  const [navigationOpen, setNavigationOpen] = useState(false);
  const [filterForm] = Form.useForm<Filters>();
  const screens = Grid.useBreakpoint();
  const title = page === "processes" ? "生产产程" : "领料申请";
  const navigate = (next: Page) => {
    setPage(next);
    setTab(next === "processes" ? "已开工" : "全部申请");
    setNavigationOpen(false);
  };
  useEffect(() => {
    document.title = `MES · ${title}`;
  }, [title]);
  useEffect(() => {
    if (!autoRefresh || page !== "requests") return;
    const controller = new AbortController();
    const timer = window.setInterval(() => {
      if (!document.hidden) void refresh(controller.signal);
    }, 5000);
    return () => {
      window.clearInterval(timer);
      controller.abort();
    };
  }, [autoRefresh, page, refresh]);
  const viewRequests = (row: ProductionProcess) => {
    filterForm.setFieldsValue({ keyword: row.id, factory: undefined });
    setFilters({ keyword: row.id });
    navigate("requests");
  };
  const processColumns: TableColumnsType<ProductionProcess> = [
    {
      title: "产程编号 / 关联订单",
      dataIndex: "id",
      width: 205,
      render: (id: string, row) => (
        <>
          <span className="code">{id}</span>
          <div className="secondary">{row.order}</div>
        </>
      ),
    },
    {
      title: "工厂 / 产线",
      dataIndex: "factory",
      width: 170,
      render: (factory: string, row) => (
        <>
          {factory}
          <div className="secondary">{row.line}</div>
        </>
      ),
    },
    { title: "生产品项", dataIndex: "product", width: 250 },
    { title: "计划开工日期", dataIndex: "planDate", width: 135 },
    {
      title: "产程状态",
      dataIndex: "status",
      width: 115,
      render: (value: string) => <StatusTag status={value} />,
    },
    {
      title: "操作",
      width: 160,
      fixed: "right",
      render: (_, row) => (
        <Space>
          <Button
            type="link"
            size="small"
            disabled={row.status === "已完工"}
            onClick={() => setProcess(row)}
          >
            领料
          </Button>
          <Button type="link" size="small" onClick={() => viewRequests(row)}>
            查看申请
          </Button>
        </Space>
      ),
    },
  ];
  const requestColumns: TableColumnsType<Requisition> = [
    {
      title: "申请单号 / 产程",
      dataIndex: "number",
      width: 280,
      render: (number: string, row) => (
        <>
          <Typography.Text
            className="code request-number"
            ellipsis={{ tooltip: number }}
          >
            {number}
          </Typography.Text>
          <div className="secondary">{row.processId}</div>
        </>
      ),
    },
    { title: "申请日期", dataIndex: "date", width: 130 },
    {
      title: "物料种类",
      width: 100,
      render: (_, row) => `${row.items.length} 种`,
    },
    {
      title: "任务状态",
      dataIndex: "status",
      width: 125,
      render: (value: string) => <StatusTag status={value} />,
    },
    {
      title: "最近结果",
      width: 220,
      render: (_, row) => (
        <Typography.Text type={row.lastError ? "danger" : "secondary"}>
          {row.lastError ||
            (row.status === "consumed"
              ? "worker 已打印消费日志"
              : row.status === "queued"
                ? "等待 worker 消费"
                : "历史演示数据")}
        </Typography.Text>
      ),
    },
    {
      title: "操作",
      width: 80,
      fixed: "right",
      render: (_, row) => (
        <Button type="link" size="small" onClick={() => setDetailId(row.id)}>
          详情
        </Button>
      ),
    },
  ];
  const query = filters.keyword?.trim().toLowerCase() ?? "";
  const processes = (catalog?.processes ?? []).filter(
    (row) =>
      (tab === "全部数据" || row.status === tab) &&
      (!filters.factory || row.factory === filters.factory) &&
      (!query ||
        [row.id, row.order, row.product].some((value) =>
          value.toLowerCase().includes(query),
        )),
  );
  const requests = records.filter((row) => {
    const factory = catalog?.processes.find(
      (item) => item.id === row.processId,
    )?.factory;
    const matches =
      tab === "全部申请" ||
      (tab === "已消费"
        ? row.status === "consumed"
        : tab === "失败"
          ? row.status === "enqueue_failed"
          : row.status === "queued");
    return (
      matches &&
      (!filters.factory || factory === filters.factory) &&
      (!query ||
        [row.number, row.processId].some((value) =>
          value.toLowerCase().includes(query),
        ))
    );
  });
  const navigation = (
    <>
      <div className="brand">
        <span className="brand-icon">M</span>
        <div>
          <strong>MES</strong>
          <small>生产执行系统</small>
        </div>
      </div>
      <div className="workspace-label">
        制造协同工作台 <Tag>DEMO</Tag>
      </div>
      <Menu
        mode="inline"
        selectedKeys={[page]}
        defaultOpenKeys={["production"]}
        items={[
          {
            key: "production",
            icon: <AppstoreOutlined />,
            label: "生产管理",
            children: [
              { key: "processes", label: "生产产程" },
              {
                key: "requests",
                label: (
                  <span className="menu-label">
                    领料申请{" "}
                    <Badge count={records.length} showZero color="#dbe7ff" />
                  </span>
                ),
              },
            ],
          },
          {
            key: "docs",
            icon: <ApiOutlined />,
            label: (
              <a href="/swagger" target="_blank" rel="noreferrer">
                接口文档
              </a>
            ),
          },
        ]}
        onClick={({ key }) => {
          if (key === "processes" || key === "requests") navigate(key);
        }}
      />
      <div className="sidebar-footer">
        <Badge status="success" text="本地演示环境" />
        <small>MES 异步任务入门</small>
      </div>
    </>
  );
  return (
    <Layout className="app-layout">
      {screens.lg && (
        <Layout.Sider width={224} theme="light" className="sidebar">
          {navigation}
        </Layout.Sider>
      )}
      <Drawer
        title="导航菜单"
        placement="left"
        size={250}
        open={navigationOpen}
        onClose={() => setNavigationOpen(false)}
        styles={{ body: { padding: 0 } }}
      >
        {navigation}
      </Drawer>
      <Layout className="main-layout">
        <Layout.Header className="topbar">
          <Space>
            {!screens.lg && (
              <Button
                type="text"
                icon={<MenuOutlined />}
                aria-label="打开导航菜单"
                onClick={() => setNavigationOpen(true)}
              />
            )}
            <Breadcrumb items={[{ title: "生产管理" }, { title }]} />
          </Space>
          <Space className="user-info">
            <Tag>演示工作区</Tag>
            <Avatar size="small" icon={<TeamOutlined />} />
            <span>生产人员</span>
          </Space>
        </Layout.Header>
        <Layout.Content className="page-content">
          <div className="page-heading">
            <div>
              <div className="eyebrow">PRODUCTION MANAGEMENT</div>
              <h1>{title}</h1>
              <Typography.Text type="secondary">
                {page === "processes"
                  ? "查看生产进度，按产程发起物料领用申请。"
                  : "观察领料任务从排队到 worker 消费的过程。"}
              </Typography.Text>
            </div>
            <Button
              icon={<ReloadOutlined />}
              loading={loading}
              onClick={() => void reload()}
            >
              刷新数据
            </Button>
          </div>
          {error && (
            <Alert
              type="error"
              showIcon
              title={error}
              className="page-alert"
              action={
                <Button size="small" onClick={() => void reload()}>
                  重试
                </Button>
              }
            />
          )}
          <Card className="filter-card">
            <Form
              form={filterForm}
              layout="vertical"
              className="filter-form"
              onFinish={setFilters}
            >
              <Form.Item label="关键词" name="keyword">
                <Input placeholder="搜索产程编号、订单或产品" allowClear />
              </Form.Item>
              <Form.Item label="工厂" name="factory">
                <Select
                  allowClear
                  placeholder="全部工厂"
                  options={[
                    ...new Set(catalog?.processes.map((row) => row.factory)),
                  ].map((factory) => ({ value: factory, label: factory }))}
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
              <Tabs
                activeKey={tab}
                onChange={setTab}
                items={(page === "processes"
                  ? ["已开工", "变更中", "已完工", "全部数据"]
                  : ["全部申请", "待消费", "已消费", "失败"]
                ).map((label) => ({ key: label, label }))}
              />
              <div className="table-tools">
                {page === "requests" && (
                  <Space>
                    <Switch
                      size="small"
                      checked={autoRefresh}
                      onChange={setAutoRefresh}
                      aria-label="自动刷新申请列表"
                    />
                    <Typography.Text type="secondary">
                      自动刷新（5 秒）
                    </Typography.Text>
                  </Space>
                )}
                <Typography.Text type="secondary">
                  共 {page === "processes" ? processes.length : requests.length}{" "}
                  条
                </Typography.Text>
              </div>
            </div>
            {page === "processes" ? (
              <Table<ProductionProcess>
                rowKey="id"
                columns={processColumns}
                dataSource={processes}
                loading={!catalog}
                scroll={{ x: 1040 }}
                pagination={{ pageSize: 10, hideOnSinglePage: true }}
                locale={{
                  emptyText: <Empty description="暂无符合条件的产程" />,
                }}
              />
            ) : (
              <Table<Requisition>
                rowKey="id"
                columns={requestColumns}
                dataSource={requests}
                loading={loading}
                scroll={{ x: 960 }}
                pagination={{ pageSize: 10, hideOnSinglePage: true }}
                locale={{
                  emptyText: (
                    <Empty description="暂无领料申请，请到生产产程发起领料" />
                  ),
                }}
              />
            )}
            <div className="table-caption">
              {page === "processes"
                ? "演示产程数据 · 选择产程后点击「领料」"
                : "MySQL 最近 100 条申请 · 默认手动刷新"}
            </div>
          </Card>
          <Alert
            type="info"
            showIcon
            title="提交申请 → Redis 排队 → worker 消费"
            description="本阶段仅保存申请、消费任务并打印日志，不调用 SCM，不扣减实际库存。"
            className="flow-alert"
          />
        </Layout.Content>
        <Layout.Footer className="app-footer">
          制造协同演示 · MES 生产执行系统
        </Layout.Footer>
      </Layout>
      {process && catalog && (
        <RequisitionModal
          process={process}
          catalog={catalog}
          onClose={() => setProcess(null)}
          onSubmitted={() => {
            setProcess(null);
            filterForm.resetFields();
            setFilters({});
            navigate("requests");
            void refresh();
          }}
        />
      )}
      <RequisitionDetail
        record={records.find((record) => record.id === detailId) ?? null}
        catalog={catalog}
        onClose={() => setDetailId(null)}
      />
    </Layout>
  );
}
