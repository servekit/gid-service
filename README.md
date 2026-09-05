# gid-service

全局 ID 生成服务。基于 Sonyflake 雪花算法，生成全局唯一 int64 ID。
可独立部署为 gRPC 服务，也可作为 Go 模块 in-process 使用。**纯 gRPC 服务**，
不监听 HTTP；对外 HTTP 面由网关（当前为 testkit-service）提供。

## 部署

```bash
make run
```

默认监听：

- gRPC：`:19091`

地址及其它运行参数见 `config.yaml`，支持环境变量覆盖。

## 接口调用

### gRPC

```bash
grpcurl -plaintext -d '{}' localhost:19091 gid.v1.GidService/NextID
grpcurl -plaintext -d '{"count":3}' localhost:19091 gid.v1.GidService/BatchNextID
```

## 作为 Go 模块（in-process）

被其他 Go 服务 import，无网络开销，直接调用 proto 类型方法：

```go
import (
    "time"

    gidv1 "github.com/servekit/gid-service/gen/gid/v1"
    gidservice "github.com/servekit/gid-service/pkg"
    "github.com/servekit/gid-service/pkg/config"
)

hdl, err := gidservice.NewModule(&config.Config{
    Snowflake: &config.SnowflakeConfig{
        MachineID: 1,
        StartTime: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
    },
})
if err != nil {
    panic(err)
}
defer hdl.Stop()

resp, err := hdl.NextID(ctx, &gidv1.NextIDRequest{})
```

`Handler` 同时实现 `signalx.Service`（Start/Stop），生命周期管理与 RPC 调用共用同一对象。
模块模式下无需配置 Server（gRPC 不启动）。

## 注意事项

- **MachineID 必须全局唯一**：多实例部署时，每个实例的 MachineID 不得重复，否则会产生重复 ID。
- **StartTime 部署后不可更改**：StartTime 是 ID 时间戳的纪元，一旦上线生成过 ID 就不能修改，否则可能导致 ID 重复或时间回退。
- **ID 寿命**：ID 为 int64，时间位以 StartTime 起算约 174 年后耗尽，规划StartTime 时留意。
- **gRPC / module 双后端**：消费方可按部署形态在 gRPC 客户端与 in-process 模块间切换，参考 demo-service 的 `internal/thirdcall/gid_service/`。
