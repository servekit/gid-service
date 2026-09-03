package gidservice

import (
	"context"

	pb "github.com/servekit/gid-service/gen/gid/v1"
	"github.com/servekit/gid-service/pkg/xcodes"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Client is a gRPC client for gid-service shaped like *Handler: it implements
// the generated pb.GidServiceServer interface (unary methods without
// grpc.CallOption), so a consumer can hold either backend behind that one
// generated interface — module mode passes the *Handler, grpc mode passes the
// *Client — with no per-consumer adapter.
//
// The UnimplementedGidServiceServer embed satisfies the interface's
// mustEmbed guard; every RPC below shadows it with a real delegation. When a
// new RPC is added to the proto, add its delegation here — until then grpc
// mode returns codes.Unimplemented for it.
type Client struct {
	pb.UnimplementedGidServiceServer

	conn *grpc.ClientConn
	cli  pb.GidServiceClient
}

// Compile-time assertion: *Client and *Handler expose the same interface.
var _ pb.GidServiceServer = (*Client)(nil)

// NewClient creates a new gRPC client.
func NewClient(target string, opts ...grpc.DialOption) (*Client, error) {
	if len(opts) == 0 {
		opts = []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	}
	conn, err := grpc.NewClient(target, opts...)
	if err != nil {
		return nil, xcodes.ErrDialService.Wrapf(err, "dial %s", target)
	}
	return &Client{conn: conn, cli: pb.NewGidServiceClient(conn)}, nil
}

// Close closes the gRPC connection.
func (c *Client) Close() error {
	return c.conn.Close()
}

// Ping delegates to the remote gid-service.
func (c *Client) Ping(ctx context.Context, in *emptypb.Empty) (*pb.Pong, error) {
	return c.cli.Ping(ctx, in)
}

// NextID delegates to the remote gid-service.
func (c *Client) NextID(ctx context.Context, in *pb.NextIDRequest) (*pb.NextIDResponse, error) {
	return c.cli.NextID(ctx, in)
}

// BatchNextID delegates to the remote gid-service.
func (c *Client) BatchNextID(ctx context.Context, in *pb.BatchNextIDRequest) (*pb.BatchNextIDResponse, error) {
	return c.cli.BatchNextID(ctx, in)
}

// Decompose delegates to the remote gid-service.
func (c *Client) Decompose(ctx context.Context, in *pb.DecomposeRequest) (*pb.DecomposeResponse, error) {
	return c.cli.Decompose(ctx, in)
}
