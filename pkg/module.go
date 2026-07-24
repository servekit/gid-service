package gidservice

import (
	"time"

	gidv1 "github.com/servekit/gid-service/gen/gid/v1"
	"github.com/servekit/gid-service/internal/service"
	"github.com/servekit/gid-service/pkg/config"
	"github.com/servekit/gid-service/pkg/handler"
	"github.com/servekit/gid-service/pkg/option"
)

// Handler is the in-process entry point. Callers invoke proto-typed RPC
// methods directly on it — no serialization, no network. Aliased to
// *handler.Handler so external code references it as gidservice.Handler.
type Handler = handler.Handler

// Compile-time assertion: *Handler satisfies the gRPC server interface.
var _ gidv1.GidServiceServer = (*Handler)(nil)

// NewModule creates a managed in-process service from snowflake parameters.
//
// machineID identifies this node (must be unique across instances).
// startTime is the epoch for ID timestamps — do not change after deployment.
//
// Returns the Handler — Handler IS the public capability and also satisfies
// signalx.Service (Start/Stop), so module users manage lifecycle via the
// same object they call RPC methods on:
//
//	hdl, err := gidservice.NewModule(1, startTime)
//	if err != nil { panic(err) }
//	defer hdl.Stop()
//	resp, err := hdl.NextID(ctx, &gidv1.NextIDRequest{})
func NewModule(machineID int64, startTime time.Time) (*Handler, error) {
	return NewModuleFromConfig(&config.Config{
		Snowflake: &config.SnowflakeConfig{
			MachineID: machineID,
			StartTime: startTime,
		},
	})
}

// NewModuleFromConfig creates a managed in-process service without starting
// gRPC. Caller may pass functional options to inject resources; by default
// the snowflake generator is built from cfg.Snowflake.
func NewModuleFromConfig(cfg *config.Config, opts ...option.Option) (*Handler, error) {
	svc, err := service.New(cfg, opts...)
	if err != nil {
		return nil, err
	}
	return handler.New(svc), nil
}
