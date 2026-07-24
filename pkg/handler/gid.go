// Package handler implements gid.v1.GidServiceServer as a thin shim over
// internal/service. Each RPC method is a one-line delegation — service takes
// the proto request directly. Handler holds NO business logic.
//
// Handler also implements signalx.Service (Start/Stop) by delegating to the
// underlying *service.Service, so in-process module users manage lifecycle
// via the same object they call RPC methods on.
package handler

import (
	"context"

	"github.com/servekit/go-common/signalx"

	gidv1 "github.com/servekit/gid-service/gen/gid/v1"
	"github.com/servekit/gid-service/internal/service"
)

// Handler implements gid.v1.GidServiceServer. It holds no mutable state —
// the embedded *service.Service owns all business state and lifecycle.
type Handler struct {
	gidv1.UnimplementedGidServiceServer

	svc *service.Service
}

// New constructs a Handler wrapping svc.
func New(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

// Compile-time assertions: Handler satisfies both the gRPC server interface
// and signalx.Service (Start/Stop).
var (
	_ gidv1.GidServiceServer = (*Handler)(nil)
	_ signalx.Service        = (*Handler)(nil)
)

// Start starts service-internal components. Safe to call from in-process
// module users before invoking RPCs.
func (h *Handler) Start() error { return h.svc.Start() }

// Stop releases resources owned by the service. After Stop, the Handler
// must not be used.
func (h *Handler) Stop() error { return h.svc.Stop() }

// NextID delegates to service.NextID.
func (h *Handler) NextID(ctx context.Context, req *gidv1.NextIDRequest) (*gidv1.NextIDResponse, error) {
	return h.svc.NextID(ctx, req)
}

// BatchNextID delegates to service.BatchNextID.
func (h *Handler) BatchNextID(ctx context.Context, req *gidv1.BatchNextIDRequest) (*gidv1.BatchNextIDResponse, error) {
	return h.svc.BatchNextID(ctx, req)
}

// Decompose delegates to service.Decompose.
func (h *Handler) Decompose(ctx context.Context, req *gidv1.DecomposeRequest) (*gidv1.DecomposeResponse, error) {
	return h.svc.Decompose(ctx, req)
}
