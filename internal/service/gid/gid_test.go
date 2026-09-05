package gid

import (
	"context"
	"testing"
	"time"

	pb "github.com/servekit/api/gen/go/gid/v1"
	"github.com/servekit/gid-service/internal/provider/snowflake"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	gen, err := snowflake.New(1, time.Now().Add(-time.Hour).Truncate(10*time.Millisecond))
	if err != nil {
		t.Fatalf("snowflake.New() error = %v", err)
	}
	return New(gen)
}

func TestNextID(t *testing.T) {
	svc := newTestService(t)

	resp, err := svc.NextID(context.Background(), &pb.NextIDRequest{})
	if err != nil {
		t.Fatalf("NextID() error = %v", err)
	}
	if resp.GetId() <= 0 {
		t.Fatalf("NextID() id = %d, want positive", resp.GetId())
	}
}

func TestBatchNextID(t *testing.T) {
	svc := newTestService(t)

	resp, err := svc.BatchNextID(context.Background(), &pb.BatchNextIDRequest{Count: 3})
	if err != nil {
		t.Fatalf("BatchNextID() error = %v", err)
	}
	if len(resp.GetIds()) != 3 {
		t.Fatalf("BatchNextID() len(ids) = %d, want 3", len(resp.GetIds()))
	}

	seen := make(map[int64]struct{}, len(resp.GetIds()))
	for _, id := range resp.GetIds() {
		if id <= 0 {
			t.Fatalf("BatchNextID() id = %d, want positive", id)
		}
		if _, ok := seen[id]; ok {
			t.Fatalf("BatchNextID() duplicate id = %d", id)
		}
		seen[id] = struct{}{}
	}
}

func TestBatchNextIDRejectsInvalidCount(t *testing.T) {
	svc := newTestService(t)

	if _, err := svc.BatchNextID(context.Background(), &pb.BatchNextIDRequest{Count: -1}); err == nil {
		t.Fatal("BatchNextID() error = nil, want error")
	}
}

func TestDecompose(t *testing.T) {
	gen, err := snowflake.New(42, time.Now().Add(-time.Hour).Truncate(10*time.Millisecond))
	if err != nil {
		t.Fatalf("snowflake.New() error = %v", err)
	}
	svc := New(gen)

	idResp, err := svc.NextID(context.Background(), &pb.NextIDRequest{})
	if err != nil {
		t.Fatalf("NextID() error = %v", err)
	}

	decResp, err := svc.Decompose(context.Background(), &pb.DecomposeRequest{Id: idResp.GetId()})
	if err != nil {
		t.Fatalf("Decompose() error = %v", err)
	}

	if decResp.GetMachineId() != 42 {
		t.Fatalf("Decompose() machineID = %d, want 42", decResp.GetMachineId())
	}
	if decResp.GetSequence() < 0 {
		t.Fatalf("Decompose() sequence = %d, want non-negative", decResp.GetSequence())
	}
	if decResp.GetTime() <= 0 {
		t.Fatalf("Decompose() time = %d, want positive unix timestamp", decResp.GetTime())
	}
	if decResp.GetGeneratedAt() == "" {
		t.Fatal("Decompose() generatedAt is empty")
	}
}
