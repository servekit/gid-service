// Package gidservice provides client, server, and in-process module access
// to gid-service.
//
// Server usage:
//
//	srv, _ := gidservice.NewServer(cfg)
//	if err := signalx.RunWithForceQuit(srv); err != nil {
//	    // handle error
//	}
//
// Client usage:
//
//	c, _ := gidservice.NewClient("localhost:19091")
//	defer c.Close()
//	resp, err := c.NextID(ctx, &pb.NextIDRequest{})
//
// Module usage:
//
//	hdl, _ := gidservice.NewModule(cfg)
//	defer hdl.Stop()
//	resp, err := hdl.NextID(ctx, &pb.NextIDRequest{})
package gidservice

import (
	"errors"

	"buf.build/go/protovalidate"
	protovalidate_middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"google.golang.org/grpc"

	"github.com/servekit/go-common/grpcx"
	"github.com/servekit/go-common/signalx"

	pb "github.com/servekit/api/gen/go/gid/v1"
	"github.com/servekit/gid-service/internal/service"
	"github.com/servekit/gid-service/pkg/config"
	"github.com/servekit/gid-service/pkg/handler"
)

// Compile-time assertion: *Server satisfies signalx.Service.
var _ signalx.Service = (*Server)(nil)

// Server wraps a gRPC server for gid-service.
//
// Holds grpcSrv (the gRPC transport) and hdl (the Handler, which itself
// wraps the underlying *service.Service and exposes Start/Stop).
// There's no separate svc field — Handler is the single handle for both
// RPC dispatch and lifecycle.
//
// gid-service is gRPC-only: no HTTP gateway runs in-process. The
// client-facing HTTP surface lives in the gateway (testkit today).
type Server struct {
	grpcSrv *grpcx.Server
	hdl     *handler.Handler
}

// NewServer creates a Server with all dependencies wired.
//
// The gRPC server runs with:
//   - grpcx.ErrorInterceptor: maps xerr-wrapped service errors to gRPC status
//     codes.
//   - protovalidate.UnaryServerInterceptor: enforces (buf.validate.field)
//     rules declared in gid.proto.
func NewServer(cfg *config.Config) (*Server, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	svc, err := service.New(cfg)
	if err != nil {
		return nil, err
	}
	hdl := handler.New(svc)

	validator, err := protovalidate.New()
	if err != nil {
		return nil, err
	}

	grpcSrv := grpcx.New(
		&grpcx.ServerConfig{GRPCAddr: cfg.Server.GRPCAddr},
		func(s *grpc.Server) { pb.RegisterGidServiceServer(s, hdl) },
		nil, // no HTTP gateway — gRPC-only service
		grpcx.ErrorInterceptor,
		grpcx.TrustedActorUnary(),
		protovalidate_middleware.UnaryServerInterceptor(validator),
	)

	return &Server{grpcSrv: grpcSrv, hdl: hdl}, nil
}

// Start starts service internals and the gRPC server without blocking.
// On partial failure, started components are rolled back via Stop.
func (s *Server) Start() error {
	if err := s.hdl.Start(); err != nil {
		return err
	}
	if err := s.grpcSrv.Start(); err != nil {
		return errors.Join(err, s.hdl.Stop())
	}
	return nil
}

// Stop gracefully stops the gRPC server and service internals.
// Errors from each component are aggregated via errors.Join.
func (s *Server) Stop() error {
	return errors.Join(s.grpcSrv.Stop(), s.hdl.Stop())
}
