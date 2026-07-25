package gidservice

import (
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

// NewModule creates a managed in-process service without starting gRPC.
// The snowflake generator is built from cfg.Snowflake; functional options
// may be passed to inject further resources.
//
// Returns the Handler — Handler IS the public capability and also satisfies
// signalx.Service (Start/Stop), so module users manage lifecycle via the
// same object they call RPC methods on:
//
//	hdl, err := gidservice.NewModule(cfg)
//	if err != nil { panic(err) }
//	defer hdl.Stop()
//	resp, err := hdl.NextID(ctx, &gidv1.NextIDRequest{})
func NewModule(cfg *config.Config, opts ...option.Option) (*Handler, error) {
	svc, err := service.New(cfg, opts...)
	if err != nil {
		return nil, err
	}
	return handler.New(svc), nil
}
