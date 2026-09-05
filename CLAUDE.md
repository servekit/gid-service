# CLAUDE.md — gid-service

## 项目定位

全局 ID 生成服务。基于 Sonyflake 雪花算法，生成全局唯一 int64 ID。
可独立部署为 gRPC 服务，也可作为 Go 模块 in-process 使用。**纯 gRPC 服务**：
不监听 HTTP，proto 不带 `google.api.http` 注解；对外 HTTP 面由网关
（当前为 testkit-service）提供。

## 技术栈约定

### gRPC / Proto

- Proto 定义在契约仓库 `../api/gid/v1/`（三分文件：service/enums/message/request_response；本仓库不持有 proto）
- 使用 `buf` 生成代码到 `gen/` 目录（protoc-gen-go + protoc-gen-go-grpc），由 `make proto` 产出
- gRPC server 监听 `:19091`

### 错误处理

- 返回的错误统一使用 `go-common/xerr`，不用裸 `fmt.Errorf` 或 `errors.New`
- 通用错误码直接用 `xcodes` 包：`xcodes.ErrInternal` 等
- 禁止 panic，所有错误通过返回值传递

### 日志

- 使用标准库 `log/slog`，禁止 `fmt.Println`、`log.Println` 等非结构化输出
- 库代码不直接打日志，通过返回 error 交给调用方
- 只有 `cmd/server/` 入口层可以打日志

### 依赖

- `go-common` — gRPC server（`grpcx`）、日志（`logging`）、错误码（`xerr`）
- `github.com/sony/sonyflake/v2` — ID 生成

### 通用

- 配置：YAML，用 `github.com/spf13/viper`
- 遵循 golang-development skill 的编码规范

### 文件内函数排列

- 导出的类型、构造函数、方法放在文件上部
- 未导出的辅助函数放在文件底部，用 `// --- internal helpers ---` 分隔

### 错误处理

- 禁止擅自添加 `//nolint` 注释，必须显式处理每个 error
- 即使是辅助操作，也要显式处理 error，不允许用 `_ =` 忽略
- 唯一例外：资源清理（`Close()` 等）可以用 `_ =`

## 架构分层

遵循 golang-service-development skill：handler / service / 领域子包三层严格分离。

- `pkg/handler/` — proto service 的薄壳，每个 RPC 一行委托到 `service.Service`；同时实现 `signalx.Service`（Start/Stop 转发）以满足 in-process module 用法
- `internal/service/service.go` — Service 本体 + `New` + `Start`/`Stop` + facade 方法（一行委托到子包）；`internal/service/` 根目录**只有** `service.go`，禁止单文件领域
- `internal/service/<domain>/` — 领域子包，业务实现 + `xxxToProto` 转换；资源通过 `New` 注入，不持有父 `*Service` 引用
- `pkg/option/` — functional options，未来可注入资源（DB/Redis 等）；当前为空

gid-service 当前只有一个 `gid` 领域，子包为 `internal/service/gid/`。加新 RPC 时：proto 加方法 → `service.go` 加 facade 一行 → `gid/gid.go` 加业务方法 → `pkg/handler/gid.go` 加委托。

## 目录结构

```
gid-service/
├── cmd/server/             # 启动入口
├── gen/                    # buf 生成代码
├── internal/
│   ├── jobs/               # cron scheduler（jobs.Scheduler，脚手架代码，当前未启用）
│   ├── provider/
│   │   └── snowflake/      # 雪花 ID 生成器（基础设施，非业务领域）
│   └── service/
│       ├── service.go      # Service 本体 + New + Start/Stop + setupJobs + facade
│       └── gid/            # gid 领域子包（业务实现）
├── pkg/                    # 可被外部 import
│   ├── client.go           # gRPC 客户端
│   ├── module.go           # in-process 入口（返回 *handler.Handler）
│   ├── server.go           # gRPC server（纯 gRPC，无 gateway）
│   ├── config/             # 配置加载（Server/Snowflake/Cron/Log）
│   ├── handler/            # proto service 薄壳
│   ├── option/             # functional options
│   └── xcodes/             # 错误码
├── CLAUDE.md
├── Makefile
├── config.yaml
├── go.mod
└── go.sum
```

## 运行模式

两种运行形态：

1. **独立 gRPC 服务** — `cmd/server/main.go` 启动，gRPC `:19091`
2. **in-process module** — 通过 `pkg.NewModule` 嵌入到父进程，无网络开销，不启动 gRPC

基础设施归 `internal/provider/`（snowflake 这种"能力提供者"不属于业务领域）；
周期任务统一走 `internal/jobs.Scheduler`（即使当前无 cron job 也保留框架）。

## 开发命令

```bash
make fmt vet lint test   # 格式化、静态检查、测试
make run                 # 本地启动（gRPC）
```

## 关联

- 架构规则：`ai-kit-studio/skills/golang-service-development`
- Go 风格：`ai-kit-studio/skills/golang-development`
- 基础库：`go-common/skills/go-common-usage`
