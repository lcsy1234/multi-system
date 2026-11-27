---
paths:
  - "app/template/manifest/protobuf/**/*.proto"
  - "app/template/api/**/*.pb.go"
  - "app/template/internal/cmd/grpc.go"
---

# gRPC 与 protobuf

## 契约位置

- proto 源文件：`app/template/manifest/protobuf/<模块>/<版本>/<模块>.proto`，版本目录用 `v1`、`v2`，与 HTTP 的 `api/<模块>/v1` 对齐。
- 生成物：`app/template/api/<模块>/<版本>/*.pb.go`（消息与 client/server 桩）、`internal/controller/<模块>/` 下的 gRPC 控制器骨架。
- 生成命令在 `app/template` 下执行 `make proto`（等价 `gf gen pb -p manifest/protobuf -a api -c internal/controller`）；插件安装用 `make proto.deps`，`protoc` 与 `gf` CLI 需预装。
- **生成文件提交进仓库**：只有契约变更时才需要重新生成，日常开发不依赖 protoc 工具链。

## 规则

- `*.pb.go` 与生成的 gRPC 控制器骨架不手改；手写实现放在同目录的非生成文件里（参考 `controller/demo/demo.go`）。
- **字段号不复用、不删除、不改类型**：废弃字段用 `reserved` 占位并保留编号；语义变化用新字段号。
- 破坏性变更开新版本包（`v2`），不改 `v1` 的既有字段语义；同一服务内 `v1`、`v2` 可共存。
- proto 文件是契约的唯一来源：不要在 Go 代码里手工构造与 proto 不一致的 JSON/结构体映射。
- 服务注册在 `internal/cmd/grpc.go`：`appgrpcx.ServerOptions(...)` → `grpcx.Server.New(...)` → 各模块 `Register(server)` → `server.Run()`。
- 新增模块必须同时改 `internal/cmd/grpc.go`，否则服务不存在。
- 控制器结构体名与 HTTP 的 `ControllerV1` 区分开（本仓库用 `Controller` + `v1.UnimplementedXxxServiceServer`），一个包里两套入口各自注册。

## 错误语义

- gRPC 用 `status.Error(codes.Xxx, msg)` 返回：参数问题 `codes.InvalidArgument`、未找到 `codes.NotFound`、无权限 `codes.PermissionDenied`、内部错误 `codes.Internal`。
- 内部错误对客户端只给稳定文案（`internal error`），细节进 `glog`；不要把 SQL、路径、堆栈放进 `status` 消息。
- 业务规则失败用 `codes.FailedPrecondition` 并要求调用方重试前修正状态，不与参数错误混用。

## 与 HTTP 的契约对齐

- 同一个业务操作在 HTTP 与 gRPC 暴露相同字段与语义：新增的 `biz_key`、`version`、
  列表的过滤条件与 `limit` 必须两边都有，不能只在一边实现。
- 参数校验一致：必填、范围、裁剪规则相同；参数问题统一 `codes.InvalidArgument`。
- 业务规则失败统一 `codes.FailedPrecondition`，消息与 HTTP 业务码文案一致；
  内部错误 `codes.Internal` + `internal error`，细节进 `glog`。
- 新增/修改 HTTP 端点且 gRPC 也需要时，proto、生成物与 controller 必须在同一提交里完成，
  并同步 README 的接口参考。

## 兼容性检查

- 改 proto 前确认现有调用方（其他服务、PDA、外部集成）与发布顺序；先发布兼容的服务端，再切换客户端。
- 新增字段必须可空或给默认值；不要给必填字段加 `required` 语义（proto3 无 required，靠业务校验）。
- 改完 proto 至少验证：`make proto` 成功、编译通过、`grpc` 子命令能起、老字段仍能读写。
