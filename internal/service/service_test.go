package service

import (
	"context"
	"testing"
	"time"

	"github.com/servekit/go-common/lifecycle"

	pb "github.com/servekit/api/gen/go/gid/v1"
	"github.com/servekit/gid-service/pkg/config"
)

var _ lifecycle.Service = (*Service)(nil)

func TestNewRejectsInvalidConfig(t *testing.T) {
	cfg := testConfig()
	cfg.Snowflake.MachineID = 0

	if _, err := New(cfg); err == nil {
		t.Fatal("New() error = nil, want error for machineID=0")
	}
}

func TestFacadeNextIDDelegates(t *testing.T) {
	svc, err := New(testConfig())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	resp, err := svc.NextID(context.Background(), &pb.NextIDRequest{})
	if err != nil {
		t.Fatalf("NextID() error = %v", err)
	}
	if resp.GetId() <= 0 {
		t.Fatalf("NextID() id = %d, want positive", resp.GetId())
	}
}

func TestFacadeStartStopNoOp(t *testing.T) {
	svc, err := New(testConfig())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := svc.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if err := svc.Stop(); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
}

func testConfig() *config.Config {
	return &config.Config{
		Server: &config.ServerConfig{
			GRPCAddr: ":9000",
		},
		Snowflake: &config.SnowflakeConfig{
			MachineID: 1,
			StartTime: time.Now().Add(-time.Hour),
		},
	}
}
