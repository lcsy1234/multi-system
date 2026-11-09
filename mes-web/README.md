# MES Web

独立的 React + TypeScript + Vite 项目，使用 Ant Design 组件与中文语言包。

## 开发

需要 Node.js 20.19+（或 22.12+）。在根目录先启动后端：

```sh
make demo
```

在另一个终端进入本目录：

```sh
npm ci
npm run dev
```

访问 http://127.0.0.1:5173/ 。Vite 将 `/api`、`/health`、`/swagger` 代理到 MES 的 8004 端口，不需要开放 CORS。修改 React/CSS 文件支持热更新。

`make demo` 不启动 worker，先提交申请观察“已排队”，另开终端执行 `make worker` 才消费。列表默认手动刷新，可手动开启 5 秒自动刷新；离开申请列表或关闭开关会取消轮询。

## 构建

```sh
npm run build
```

该命令先做 TypeScript 检查，再生成 `dist/`。MES 在 http://127.0.0.1:8004/ 提供这些静态文件，不再使用 `go:embed`；只修改前端时，重新构建即可更新 8004 页面，不必重新编译 Go。

从 `mes/` 启动后端时，默认读取 `../mes-web/dist`，部署时可通过 `MES_WEB_DIR` 指定绝对路径。尚未构建时，8004 首页会提示先执行 `make frontend-build`；API、健康检查仍独立工作。

## 目录

```text
src/
├── main.tsx                       # 中文配置与主题
├── App.tsx                        # 后台布局、菜单、产程/申请列表
├── api.ts                         # GoFrame API 响应处理
├── types.ts                       # 接口类型
├── useMES.ts                      # 基础数据与申请加载
├── styles.css                     # 布局与响应式样式
└── components/
    ├── MaterialsEditor.tsx         # 物料明细编辑
    ├── RequisitionModal.tsx        # 领料申请弹窗
    ├── RequisitionDetail.tsx       # 详情与时间线
    └── StatusTag.tsx              # 状态展示
```

当前只演示 MES 保存申请、入队与消费日志，不包含 SCM 调用和实际库存扣减。数量使用十进制字符串与 Ant Design InputNumber 的 stringMode，提交时由前后端分别校验。

使用方案参考 [Ant Design 官方 Vite 文档](https://ant.design/docs/react/use-with-vite/)。
