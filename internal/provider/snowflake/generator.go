// Package snowflake wraps sonyflake for global unique ID generation.
package snowflake

import (
	"time"

	"github.com/servekit/gid-service/pkg/xcodes"

	"github.com/sony/sonyflake/v2"
)

const (
	// MaxMachineID is the largest machine ID supported by Sonyflake's default 16-bit machine field.
	MaxMachineID int64 = 1<<16 - 1
	// MaxBatchCount matches the public API validation limit.
	MaxBatchCount = 1000
)

// Generator generates globally unique IDs using Sonyflake.
type Generator struct {
	sf *sonyflake.Sonyflake
}

// New creates a new Generator.
// machineID identifies this node (must be unique across instances).
// startTime is the epoch for ID timestamps — do not change after deployment.
func New(machineID int64, startTime time.Time) (*Generator, error) {
	if machineID < 1 || machineID > MaxMachineID {
		return nil, xcodes.ErrMachineIDInvalid
	}
	if startTime.After(time.Now()) {
		return nil, xcodes.ErrStartTimeFuture
	}

	mid := int(machineID)
	sf, err := sonyflake.New(sonyflake.Settings{
		StartTime: startTime,
		MachineID: func() (int, error) {
			return mid, nil
		},
	})
	if err != nil {
		return nil, xcodes.ErrCreateSonyflake.Wrap(err)
	}

	return &Generator{sf: sf}, nil
}

// NextID generates a single unique ID.
func (g *Generator) NextID() (int64, error) {
	id, err := g.sf.NextID()
	if err != nil {
		return 0, xcodes.ErrGenerateID.Wrap(err)
	}
	return id, nil
}

// BatchNextID generates count unique IDs.
func (g *Generator) BatchNextID(count int) ([]int64, error) {
	if count < 1 || count > MaxBatchCount {
		return nil, xcodes.ErrBatchCountInvalid
	}

	ids := make([]int64, count)
	for i := range count {
		id, err := g.sf.NextID()
		if err != nil {
			return nil, xcodes.ErrGenerateID.Wrapf(err, "generate id %d/%d", i+1, count)
		}
		ids[i] = id
	}
	return ids, nil
}

// DecomposedID holds the parsed components of a Sonyflake ID.
type DecomposedID struct {
	Time      time.Time
	Sequence  int64
	MachineID int64
}

// Decompose parses an ID into its components.
func (g *Generator) Decompose(id int64) *DecomposedID {
	parts := g.sf.Decompose(id)
	return &DecomposedID{
		Time:      g.sf.ToTime(id),
		Sequence:  parts["sequence"],
		MachineID: parts["machine"],
	}
}
