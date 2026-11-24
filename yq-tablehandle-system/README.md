# yq-tablehandle-system

表格处理系统的初始骨架。React 首页通过 `GET /api/health` 显示 GoFrame 后端的连接状态，暂不包含表格业务功能。

## 本地启动

需要 Go 1.24.13、Node.js 20.19+（或 22.12+）和 npm。从本子项目目录分别在两个终端运行：

```sh
cd backend && go run .
```

```sh
cd frontend && npm ci && npm run dev
```

访问 http://127.0.0.1:5175 。开发服务器将 `/api` 请求代理到 http://127.0.0.1:8191 。后端正常时返回 `{"status":"ok"}`，首页显示「已连接」。

前端构建：在 `frontend` 目录执行 `npm run build`。生产环境需另外配置 `/api` 反向代理。
