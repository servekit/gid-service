# gid-service

全局 ID 生成服务。基于 Sonyflake 雪花算法，生成全局唯一 int64 ID。
可独立部署为 gRPC 服务，也可作为 Go 模块 in-process 使用。

## 架构

遵循 ai-kit-studio 的 `golang-service-development` skill 三层分层：

- `pkg/handler.Handler` — proto service 薄壳，每个 RPC 一行委托
- `internal/service.Service` — Service 本体 + facade + 资源管理（`lifecycle.Manager`）+ `setupJobs`
- `internal/service/gid.Service` — gid 领域子包，业务实现

基础设施归 `internal/provider/`（snowflake 这种"能力提供者"不属于业务领域）；
周期任务统一走 `internal/jobs.Scheduler`（即使当前无 cron job 也保留框架）。

三种运行模式：
1. **独立 gRPC 服务** — `cmd/server/main.go` 启动，gRPC `:19091` + HTTP gateway `:18081`
2. **HTTP gateway** — 自动注册到 gRPC server，REST 客户端可直接调用
3. **in-process module** — 通过 `pkg.NewModule` 嵌入到父进程，无网络开销

## 用法

### 独立部署

```bash
make run
```

gRPC server 监听 `:19091`，HTTP gateway 监听 `:18081`。

### gRPC 调用

```bash
grpcurl -plaintext -d '{}' localhost:19091 gid.v1.GidService/NextID
grpcurl -plaintext -d '{"count":3}' localhost:19091 gid.v1.GidService/BatchNextID
```

### HTTP gateway 调用

```bash
curl http://localhost:18081/v1/gid/next
curl -X POST http://localhost:18081/v1/gid/batch -d '{"count":3}'
curl http://localhost:18081/v1/gid/decompose/24804801279688705
```

### in-process module（被其他 Go 服务 import）

```go
import (
    gidservice "gid-service/pkg"
    "gid-service/pkg/option"
    gidv1 "gid-service/gen/gid/v1"
)

hdl, err := gidservice.NewModule(1, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
if err != nil { panic(err) }
defer hdl.Stop()

resp, err := hdl.NextID(ctx, &gidv1.NextIDRequest{})
```

参考 demo-service 的 `internal/thirdcall/gidservice/` 看 grpc / module 双后端切换。

## 开发

```bash
make fmt vet lint test
make proto  # 重生成 gen/
```

## 关联

- 架构规则：`ai-kit-studio/skills/golang-service-development`
- Go 风格：`ai-kit-studio/skills/golang-development`
- 基础库：`go-common/skills/go-common-usage`
