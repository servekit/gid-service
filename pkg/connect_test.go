package gidservice

import (
	"context"
	"testing"
	"time"

	"github.com/servekit/gid-service/pkg/config"
	"github.com/servekit/go-common/lifecycle"

	"github.com/stretchr/testify/require"
)

func testConnectConfig() ConnectConfig {
	return ConnectConfig{Config: &config.Config{
		Snowflake: &config.SnowflakeConfig{MachineID: 3, StartTime: time.Now().Add(-time.Hour)},
	}}
}

// TestConnect_ModuleDuplicateFails verifies the one-live-module-per-process
// invariant: a second module-mode Connect while the first instance is still
// active errors with the sharing remedy, and stopping the first frees the
// slot for a rebuild.
func TestConnect_ModuleDuplicateFails(t *testing.T) {
	mgr1 := lifecycle.NewManager()
	svc1, raw1, err := Connect(testConnectConfig(), mgr1)
	require.NoError(t, err)
	require.NotNil(t, raw1)

	id, err := NextID(context.Background(), svc1)
	require.NoError(t, err)
	require.NotZero(t, id)

	// Second build in the same process while the first is live → error.
	mgr2 := lifecycle.NewManager()
	_, _, err = Connect(testConnectConfig(), mgr2)
	require.ErrorContains(t, err, "already active")
	require.ErrorContains(t, err, "With*Handler")

	// Stopping the first releases the claim; a rebuild is legal.
	require.NoError(t, mgr1.Stop())
	mgr3 := lifecycle.NewManager()
	svc3, raw3, err := Connect(testConnectConfig(), mgr3)
	require.NoError(t, err)
	require.NotNil(t, raw3)

	id3, err := NextID(context.Background(), svc3)
	require.NoError(t, err)
	require.NotZero(t, id3)
	require.NoError(t, mgr3.Stop())
}
