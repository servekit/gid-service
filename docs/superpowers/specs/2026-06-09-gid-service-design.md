# GID Service Design

Global ID (GID) Service — a lightweight Snowflake ID generation service.

## Goals

- Generate globally unique int64 IDs using Sonyflake algorithm
- Usable as standalone gRPC service OR embedded Go module (no network overhead)
- Simple, stateless, no database dependency

## Architecture

Two usage modes via `pkg/`:

- **Server** (`pkg/server.go`): gRPC + grpc-gateway HTTP, mirrors user-service pattern
- **Module** (`pkg/module.go`): in-process, no gRPC overhead, for embedding
- **Client** (`pkg/client.go`): thin gRPC client wrapper

## Directory Structure

```
gid_service/
├── api/proto/gid/v1/gid.proto
├── buf.gen.yaml / buf.yaml
├── cmd/server/main.go
├── config.yaml
├── internal/
│   ├── config/config.go
│   ├── snowflake/generator.go
│   └── service/gid_service.go
├── gen/gid/v1/
├── pkg/
│   ├── server.go
│   ├── client.go
│   └── module.go
├── go.mod / go.sum
├── Makefile
├── CLAUDE.md
└── .gitea/workflows/ci.yml
```

## Proto API

```protobuf
service GidService {
  rpc NextID(NextIDRequest) returns (NextIDResponse);
  rpc BatchNextID(BatchNextIDRequest) returns (BatchNextIDResponse);
  rpc Decompose(DecomposeRequest) returns (DecomposeResponse);
}
```

- `NextID`: generate a single ID
- `BatchNextID`: generate multiple IDs (count field, protovalidate max 1000)
- `Decompose`: parse an ID into time, sequence, machine_id, generated_at

## Snowflake Generator (`internal/snowflake/`)

- Based on `github.com/sony/sonyflake/v2`
- `StartTime`: configurable via YAML, hardcoded once deployed, never change
- `MachineID`: from config, manually assigned per node
- Methods: `NextID()`, `BatchNextID(count)`, `Decompose(id)`, `ToTime(id)`
- Sonyflake rate: up to 256 IDs per 10ms per node

## Configuration

```yaml
server:
  grpc:
    addr: ":9000"
  gateway:
    addr: ":8080"

snowflake:
  machine_id: 1
  start_time: "2026-06-01T00:00:00Z"
```

No database, no Redis. Pure in-memory generation.

## Key Decisions

| Decision | Choice | Why |
|----------|--------|-----|
| Algorithm | Sonyflake | User's existing choice, distributed-friendly |
| MachineID | Config file | Simple, matches existing snowflake.node convention |
| Database | None | Stateless ID generation |
| Redis | None | No caching/session needs |
| StartTime | Configurable YAML | Flexible but must not change after deployment |

## Dependencies

- `github.com/sony/sonyflake/v2` — ID generation
- `github.com/servekit/go-common` — grpcx, logging, xerr/xcodes (same as user-service)
- Standard gRPC stack: buf, protovalidate, grpc-gateway
