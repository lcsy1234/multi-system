# tablehandle

工厂填报、Excel 模板与 AI 脚本转换服务。`api/admin/conversion` 和 `api/admin/script` 各自定义模块接口，其 `v1/` 存放请求与响应契约。这里的 `admin` 是源码分组，不改变 HTTP 路径或权限，`boot/http` 是 HTTP 入口，`internal/controller` 提供 `/api` 路由，`internal/logic/report` 实现填报。

在本目录运行 `make run`、`make test` 或 `make build`。配置示例在 `manifest/config/env.example`，本地密钥仍从子项目根 `.env` 加载，详见[启动说明](../../../README.md)。
