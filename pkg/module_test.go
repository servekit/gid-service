package gidservice

import (
	"context"
	"testing"
	"time"

	"github.com/servekit/go-common/lifecycle"
	"github.com/servekit/go-common/signalx"

	pb "github.com/servekit/gid-service/gen/gid/v1"
	"github.com/servekit/gid-service/pkg/config"
)

var (
	_ lifecycle.Service = (*Handler)(nil)
	_ signalx.Service   = (*Handler)(nil)
)

func TestModuleBatchNextIDRejectsInvalidCount(t *testing.T) {
	hdl, err := NewModule(&config.Config{
		Snowflake: &config.SnowflakeConfig{
			MachineID: 1,
			StartTime: time.Now().Add(-time.Hour),
		},
	})
	if err != nil {
		t.Fatalf("NewModule() error = %v", err)
	}

	_, err = hdl.BatchNextID(context.Background(), &pb.BatchNextIDRequest{Count: -1})
	if err == nil {
		t.Fatal("BatchNextID() error = nil, want error")
	}
}

func TestNewModuleDoesNotRequireServerConfig(t *testing.T) {
	hdl, err := NewModule(&config.Config{
		Snowflake: &config.SnowflakeConfig{
			MachineID: 1,
			StartTime: time.Now().Add(-time.Hour),
		},
	})
	if err != nil {
		t.Fatalf("NewModule() error = %v", err)
	}

	if err := hdl.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if err := hdl.Stop(); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
}

func TestModuleNextIDRoundTrip(t *testing.T) {
	hdl, err := NewModule(&config.Config{
		Snowflake: &config.SnowflakeConfig{
			MachineID: 1,
			StartTime: time.Now().Add(-time.Hour),
		},
	})
	if err != nil {
		t.Fatalf("NewModule() error = %v", err)
	}

	resp, err := hdl.NextID(context.Background(), &pb.NextIDRequest{})
	if err != nil {
		t.Fatalf("NextID() error = %v", err)
	}
	if resp.GetId() <= 0 {
		t.Fatalf("NextID() id = %d, want positive", resp.GetId())
	}
}
