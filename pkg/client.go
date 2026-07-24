package gidservice

import (
	pb "github.com/servekit/gid-service/gen/gid/v1"
	"github.com/servekit/gid-service/pkg/xcodes"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client wraps the generated gRPC client for gid-service.
// Embeds pb.GidServiceClient so all RPC methods are directly available.
type Client struct {
	conn *grpc.ClientConn
	pb.GidServiceClient
}

// NewClient creates a new gRPC client.
func NewClient(target string, opts ...grpc.DialOption) (*Client, error) {
	if len(opts) == 0 {
		opts = []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	}
	conn, err := grpc.NewClient(target, opts...)
	if err != nil {
		return nil, xcodes.ErrDialService.Wrapf(err, "dial %s", target)
	}
	return &Client{conn: conn, GidServiceClient: pb.NewGidServiceClient(conn)}, nil
}

// Close closes the gRPC connection.
func (c *Client) Close() error {
	return c.conn.Close()
}
