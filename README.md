# multi-system

生产领料场景下的四系统通信练习工程。四个项目均使用 GoFrame v2.10.2，拥有独立 Go module、配置和进程；根目录通过 `go.work` 组织本地开发。

| 项目 | 职责 | 地址 | 接口文档 |
| --- | --- | --- | --- |
| `erp/` | 金蝶 ERP 本地模拟，接收业务单据同步 | http://127.0.0.1:8001 | http://127.0.0.1:8001/swagger |
| `wms/` | 管理仓库库存、拣货和发料出库 | http://127.0.0.1:8002 | http://127.0.0.1:8002/swagger |
| `scm/` | 调拨、生产订单履约、协调领料 | http://127.0.0.1:8003 | http://127.0.0.1:8003/swagger |
| `mes/` | 生产执行，生产人员提交领料申请 | http://127.0.0.1:8004 | http://127.0.0.1:8004/swagger |

## 启动

需要 Go 1.24+，统一启动脚本还需要 Python 3 和 make。运行不依赖 `gf` CLI，也不依赖数据库、消息队列或 Docker。

```sh
make deps
make dev
```

脚本先编译四个服务，再检查端口、启动进程并验证健康检查、OpenAPI 和 Swagger。日志保存在 `.logs/`，Ctrl+C 停止本次启动的全部服务。端口被占用时会报错，不会结束其他进程。

也可以在不同终端单独启动：

```sh
cd mes
go run .
```

每个项目从自己的目录读取 `manifest/config/config.yaml`。默认仅监听本机地址，修改该文件可调整端口；统一启动脚本的端口映射需要一起调整。

## 验证

```sh
make fmt
make test
make vet
make smoke
```

`make test` 在关闭 workspace 的条件下分别检查四个独立模块；目前没有业务逻辑和业务单元测试。`make smoke` 会真正启动四个二进制进程，验证健康响应中的服务名称、OpenAPI 路由与 Swagger 页面，结束后清理进程。请先停止正在运行的 `make dev` 再执行烟测。

```sh
curl http://127.0.0.1:8004/health
```

响应使用 GoFrame 标准封装，例如：

```json
{"code":0,"message":"","data":{"service":"mes","status":"ok","role":"Manufacturing execution and material requisitions"}}
```

`/health` 仅表示当前进程可服务，不表示下游依赖或业务链路可用。

## 当前范围

已经完成项目骨架、健康接口、独立配置、OpenAPI 和统一开发命令。
领料业务、持久化、系统间调用、库存扣减、消息队列和真实金蝶对接尚未实现；当前 ERP 是练习服务，不是金蝶产品的实现或官方接口兼容服务。

业务边界、领料流程、建议接口和后续练习见 [领料链路设计](docs/material-issue-flow.md)。真实金蝶的产品版本、单据类型、字段映射、保存/提交/审核流程和认证方式需在接入时确认。

## 项目结构

```text
multi-system/
├── go.work
├── Makefile
├── scripts/services.py
├── docs/material-issue-flow.md
├── erp/                  # 独立 go.mod、main.go、配置
├── wms/
├── scm/
└── mes/
```

四个服务之间不通过 Go 包互相调用；之后通过 HTTP 或消息传输接口契约。每个服务的 `api/` 定义接口，`internal/controller/` 接收请求，`internal/logic/` 放业务逻辑，`internal/client/` 放其他系统的适配客户端。数据库选型后再填充 `internal/model/` 和 `internal/dao/`。

工程布局与路由方式参考 [GoFrame 官方项目初始化](https://goframe.org/en/2.9.x/quick/scaffold-boost) 和 [控制器生成](https://goframe.org/en/docs/cli/gen-ctrl)。本项目使用本机 `gf init` 的官方模板，再固定依赖版本、替换示例接口并移除尚未采用的部署模板。
