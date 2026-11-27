# 元气森林工厂数据智能填报 Agent

员工上传 Excel，选择工厂、日期、产线和 SKU，即可生成沿用原模板格式的填报文件。工作台展示实际工具调用、数据预览、来源表和缺失提醒。所有业务数据均为演示数据。

## 启动

需要 Go 1.24.13+、Node.js 20.19+ 或 22.12+、npm。以下命令从仓库根目录执行。

**终端一，启动 GoFrame 后端**

```sh
cd yq-tablehandle-system
cp -n backend/app/tablehandle/manifest/config/env.example .env
cd backend
make run
```

后端采用 `template-gf` 的 `ai.hy` 分支 HTTP 结构和 GoFrame v2.10.3。在 `backend` 下也可执行 `go run ./app/tablehandle/boot/http`。默认监听 `127.0.0.1:8191`，可用 `TABLEHANDLE_HTTP_ADDR` 覆盖；该应用不需要模板中的数据库、Redis 或 worker。

**终端二，启动网页**

前端为盘古插件 `tablehandle`。安装前需有云效 npm 制品库的读取权限；项目 `.npmrc` 仅指定 `@pangu` 源，认证请按云效仓库指南配置在个人 `~/.npmrc`，不要把 Token 写进仓库。CI 使用只读 Secret `PANGU_NPM_TOKEN`。

```sh
cd yq-tablehandle-system/frontend
npm ci
npm run dev
```

打开 [工厂智能填报工作台](http://127.0.0.1:5175)。Vite 将同源的 `/api` 代理到本机 8191 端口。两个进程都需保持运行。

**盘古 OP 联调**：运行 `npm start`，在宿主 `/pd` 为 `tablehandle` 配置终端打印的本地插件地址。在菜单管理中按下表登记「文件名」，链接和授权由宿主管理；正式插件不使用独立预览的 hash 路由。

| 父菜单 | 子菜单 | 文件名 |
| --- | --- | --- |
| 智能填报 | 工厂数据智能填报 | `plugin://tablehandle/report` |
| AI 脚本 | AI 优化脚本 | `plugin://tablehandle/ai-script` |
| AI 脚本 | 脚本转 JSON | `plugin://tablehandle/script-json` |
| AI 脚本 | 元气小虾米转 JSON | `plugin://tablehandle/shrimp-json` |

宿主联调还需要配置业务网关：SDK 的 JSON 请求走宿主 baseURL，应将 `src/config.url.tablehandle.ts` 中的业务接口转发给 GoFrame，并把 GoFrame 的原始 JSON 包装成宿主的 `{ code, data, message }` 契约。SSE `/api/run`、模板 `/api/sample` 和结果 `/api/results/:id` 使用宿主同源的鉴权转发，透传事件流和文件，关闭 SSE 缓冲；Vite 代理不作用于宿主页面。脚本试跑需要宿主 CSP 允许 Blob Worker 和 WebAssembly。

**选择运行模式**

在 **脚本转 JSON** 粘贴配置（最大 64 KiB）。内置解析支持 XML、带注释的 JSON、键值对和 INI 分组。转换失败或结果需要调整时，点 **带此样本优化脚本**，说明这种格式该怎么处理。点击 **生成并试跑脚本** 后，后端将样本、要求、当前结果及已匹配的扩展脚本发送给配置的 DeepSeek 服务。请先移除敏感数据。核对试跑结果后点击 **应用到脚本转 JSON**。扩展脚本保存在当前浏览器，后续按从新到旧的顺序本地匹配；可以在 **AI 优化脚本** 移除。

- 默认留空密钥即可体验。本地演示模式按固定顺序执行四个 Eino 工具，支持自动匹配已知列名和“只填实际产量和良品率，汇总行保留公式”一类的指标选择。
- 使用大模型时，在子项目的 `.env` 填入 `DEEPSEEK_API_KEY` 后重启后端，右上角会显示 DeepSeek。默认模型为 `deepseek-flash`，可通过 `DEEPSEEK_MODEL` 和 `DEEPSEEK_BASE_URL` 指定支持工具调用的模型及地址。已导出的环境变量优先。在线模式会向配置的模型服务发送模板解析信息、筛选说明和工具返回的 Mock 数据。
- 在线模式使用 Eino ReAct 和 DeepSeek ChatModel。模型决定表头与字段的对应关系和需要写入的指标。服务端固定工厂与日期范围，并从真实查询结果取值。调用顺序、字段、目标行和工作表均经过校验。模型或工具报错时网页显示失败原因，可修改条件后重新发起任务。

## 体验示例

1. 点击“下载示例模板”，再将文件拖入上传区。下载接口按上海日期生成《工厂生产日报》，包含合并标题、近 14 天的日期、百分比样式以及四个 SUM 汇总公式。
2. 保留默认工厂和日期范围，点击“开始智能填报”。示例将返回 14 行按天汇总的数据。选择单条产线或 SKU 可缩小汇总范围。
3. 再试一次，在补充说明输入“只填实际产量和良品率，汇总行保留公式”。下载后检查 D、H 列，以及第 18 行的汇总公式。Excel 打开文件时计算公式。
4. 选择全部位于数据窗口之前的日期，结果应为零行，说明中会列出无数据日期。模板原有日期和公式仍可下载。

## 模板与数据口径

每次任务处理一本工作簿中的一个工作表，使用一张来源表。其他工作表随结果文件交付。表头应在同一行，至少两列，并含“日期”“工厂名称”或可识别的业务字段。支持预填维度、标准 Excel 日期和示例中的中文列名。上传上限为 8 MB，目标工作表上限为 5000 行、80 列。

在表头下预留明细行，可填日期或设置行样式。匹配时优先使用已有维度，空白行用于剩余记录。数据量超过可填区域时报告剩余条数。汇总区域以下的行保留原样。表头、合并区域、已有值和公式均受保护。百分率按数值写入，显示精度继承模板格式。

| 来源表 | 明细粒度 | 业务口径 |
| --- | --- | --- |
| dws_factory_production_daily | 日期、工厂、产线、SKU | 计划、实际、合格及损耗按箱求和。达成率为实际合计除以计划合计，良品率为合格合计除以实际合计。 |
| dwd_equipment_oee_daily | 日期、工厂、产线 | 运行和停机时长按小时求和，能耗按度求和，OEE 按运行时长加权。 |
| db_inventory_inbound | 入库日期、工厂、SKU | 入库箱数求和，质检状态按组展示。 |

每次运行生成含当天的连续 14 天数据，覆盖滁州、肇庆、天津三座工厂，每厂两条产线和三个 SKU。查询结果按模板包含的维度分组。设备数据没有 SKU 维度，入库数据没有产线维度，对这些来源使用不适用的筛选将得到缺失提醒。

## 服务结构

- `backend/app/tablehandle/boot/http` 是模板式 HTTP 入口，后端根目录通过 Makefile 选择服务。`internal/cmd` 启动服务，`internal/controller/tablehandle` 提供上传、模板下载、SSE 和结果下载。任务独立打开原始文件并以随机文件名另存。文件位于本机临时目录，重启服务后需重新上传。
- `backend/app/tablehandle/internal/logic/report/agent.go` 定义 System Prompt 和四个 Eino 工具。查询使用只读结构化条件，拒绝 SQL 文本。回填工具接收目标 Sheet 与查询结果中的行号，逐格检查空值和公式，生成报告。
- `backend/app/tablehandle/internal/logic/report/workbook.go` 负责示例生成、模板解析与坐标识别。`backend/app/tablehandle/internal/logic/report/warehouse.go` 定义内存表、筛选和聚合。
- `frontend/src/manifest.ts` 声明页面，`src/pages/layout.sidebar.tsx` 使用宿主菜单树渲染侧栏，业务样式限定在插件命名空间下。

| 方法 | 路径 | 用途 |
| --- | --- | --- |
| GET | /api/health、/api/options | 服务健康、模式、筛选项与默认日期 |
| GET | /api/sample | 下载当天生成的示例模板 |
| POST | /api/uploads | multipart 的 file 字段，返回上传 ID 和 Sheet 列表 |
| POST | /api/run | JSON 包含 id、sheet、filters，返回 SSE 的 start、step、result 或 failure 事件 |
| GET | /api/results/:id | 下载任务输出的 Excel |

## 验证与构建

```sh
cd yq-tablehandle-system/backend
go test ./...
go build -o /tmp/factory-report-agent .
cd ../frontend
npm ci
npm run build
```

`ENV_YQSL=dev npm run build` 先类型检查，再生成 `dist/remoteEntry.js`，不上传 OSS。独立站使用 `npm run build:preview` 生成 `dist-preview/`，将 `/api` 转发到 GoFrame；`npm run build:offline` 保留离线转换工具。

## 果糖接收、使用记录表

可直接上传原版《果糖接收、使用记录表.xlsx》。系统识别三层合并表头，将备注栏上方的 17 个空白行作为明细区域。默认按日期填入近 14 天的模拟接收记录、两个储罐温度、三组线别与品项的白夜班用量，并填写日期范围和对应组名。底部签名栏供实际人员签署。

演示数据源为 `dws_fructose_receipt_usage_daily`，每天每厂一条记录。模板中的“结存量”暂按期初库存解释，“日结存”为期初加接收量减去六个班次用量，单位为吨。按产线或 SKU 筛选时，只展示匹配组的用量，库存和接收量仍是全厂口径。

接真实数仓时，需要只读连接或查询 API、实际表结构和字段口径。将返回值映射成日期、工厂及上述记录字段，替换后端 `seedFructose` 的数据提供逻辑即可复用模板写入流程。三组线别与品项当前使用演示配置，接入时也需对应现场配置。
