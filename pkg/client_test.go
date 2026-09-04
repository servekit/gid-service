package gidservice

import (
	"context"
	"net"
	"testing"
	"time"

	gidv1 "github.com/servekit/gid-service/gen/gid/v1"
	"github.com/servekit/gid-service/pkg/config"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// TestClient_GRPCRoundTrip runs a real in-process gRPC server backed by the
// module-mode Handler and drives it through the server-shaped *Client. This
// is the wire-level smoke for every delegation: a self-recursive or
// mis-routed delegation (the class of bug unit tests miss, because they never
// exercise *Client) fails here.
func TestClient_GRPCRoundTrip(t *testing.T) {
	hdl, err := NewModule(&config.Config{
		Snowflake: &config.SnowflakeConfig{MachineID: 7, StartTime: time.Now().Add(-time.Hour)},
	})
	require.NoError(t, err)

	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	gs := grpc.NewServer()
	gidv1.RegisterGidServiceServer(gs, hdl)
	go func() { _ = gs.Serve(lis) }()
	t.Cleanup(gs.Stop)

	c, err := NewClient(lis.Addr().String())
	require.NoError(t, err)
	t.Cleanup(func() { _ = c.Close() })

	ctx := context.Background()

	pong, err := c.Ping(ctx, &emptypb.Empty{})
	require.NoError(t, err)
	require.Equal(t, "SERVING", pong.GetStatus())

	id1, err := NextID(ctx, c)
	require.NoError(t, err)
	require.NotZero(t, id1)

	id2, err := NextID(ctx, c)
	require.NoError(t, err)
	require.NotEqual(t, id1, id2, "snowflake IDs must be unique")

	ids, err := c.BatchNextID(ctx, &gidv1.BatchNextIDRequest{Count: 3})
	require.NoError(t, err)
	require.Len(t, ids.GetIds(), 3)

	dec, err := c.Decompose(ctx, &gidv1.DecomposeRequest{Id: id1})
	require.NoError(t, err)
	require.Equal(t, int64(7), dec.GetMachineId())
}
