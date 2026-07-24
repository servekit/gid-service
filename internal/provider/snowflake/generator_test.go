package snowflake

import (
	"errors"
	"testing"
	"time"

	"github.com/servekit/go-common/xerr"
)

func TestGeneratorDecomposeReturnsMachineID(t *testing.T) {
	startTime := testStartTime()
	gen, err := New(42, startTime)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	id, err := gen.NextID()
	if err != nil {
		t.Fatalf("NextID() error = %v", err)
	}

	d := gen.Decompose(id)
	if d.MachineID != 42 {
		t.Fatalf("machineID = %d, want 42", d.MachineID)
	}
	if d.Sequence < 0 {
		t.Fatalf("sequence = %d, want non-negative", d.Sequence)
	}
	if d.Time.Before(startTime) {
		t.Fatalf("generatedAt = %v, want after %v", d.Time, startTime)
	}
}

func TestNewValidatesMachineIDRange(t *testing.T) {
	startTime := testStartTime()

	tests := []struct {
		name      string
		machineID int64
		wantErr   bool
	}{
		{name: "zero", machineID: 0, wantErr: true},
		{name: "negative", machineID: -1, wantErr: true},
		{name: "max", machineID: MaxMachineID},
		{name: "over max", machineID: MaxMachineID + 1, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.machineID, startTime)
			if tt.wantErr {
				if err == nil {
					t.Fatal("New() error = nil, want error")
				}
				var xerrErr *xerr.Error
				if !errors.As(err, &xerrErr) {
					t.Fatalf("New() error type = %T, want *xerr.Error", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}
		})
	}
}

func TestBatchNextIDValidatesCount(t *testing.T) {
	startTime := testStartTime()
	gen, err := New(1, startTime)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	tests := []struct {
		name    string
		count   int
		wantLen int
		wantErr bool
	}{
		{name: "zero", count: 0, wantErr: true},
		{name: "negative", count: -1, wantErr: true},
		{name: "one", count: 1, wantLen: 1},
		{name: "max", count: MaxBatchCount, wantLen: MaxBatchCount},
		{name: "over max", count: MaxBatchCount + 1, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ids, err := gen.BatchNextID(tt.count)
			if tt.wantErr {
				if err == nil {
					t.Fatal("BatchNextID() error = nil, want error")
				}
				var xerrErr *xerr.Error
				if !errors.As(err, &xerrErr) {
					t.Fatalf("BatchNextID() error type = %T, want *xerr.Error", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("BatchNextID() error = %v", err)
			}
			if len(ids) != tt.wantLen {
				t.Fatalf("len(ids) = %d, want %d", len(ids), tt.wantLen)
			}
		})
	}
}

func testStartTime() time.Time {
	return time.Now().Add(-time.Hour).Truncate(10 * time.Millisecond)
}
