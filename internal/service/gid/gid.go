// Package gid contains the gid domain business logic.
//
// Layering contract (see golang-service-development skill §2):
//   - This is a SUBPACKAGE under internal/service/. gid methods are NOT on
//     the outer *service.Service; the outer service exposes them via one-line
//     facade methods (see ../service.go).
//   - Methods take proto types DIRECTLY and return proto types — no
//     intermediate Go structs.
//   - Resources (snowflake generator) are injected via New(); the subpackage
//     does NOT hold a reference to the parent *service.Service (avoids import
//     cycle).
//   - The subpackage does NOT manage resource lifecycle — that's the parent
//     service.Service's job via lifecycle.Manager.
package gid

import (
	"context"
	"time"

	gidv1 "github.com/servekit/gid-service/gen/gid/v1"
	"github.com/servekit/gid-service/internal/provider/snowflake"
)

// Service is the gid domain service. The snowflake generator is injected at
// construction; the subpackage does not manage its lifecycle.
type Service struct {
	gen *snowflake.Generator
}

// New constructs a gid domain service with the snowflake generator injected.
func New(gen *snowflake.Generator) *Service {
	return &Service{gen: gen}
}

// NextID generates a single unique ID.
func (s *Service) NextID(_ context.Context, _ *gidv1.NextIDRequest) (*gidv1.NextIDResponse, error) {
	id, err := s.gen.NextID()
	if err != nil {
		return nil, err
	}
	return &gidv1.NextIDResponse{Id: id}, nil
}

// BatchNextID generates count unique IDs.
func (s *Service) BatchNextID(_ context.Context, req *gidv1.BatchNextIDRequest) (*gidv1.BatchNextIDResponse, error) {
	ids, err := s.gen.BatchNextID(int(req.GetCount()))
	if err != nil {
		return nil, err
	}
	return &gidv1.BatchNextIDResponse{Ids: ids}, nil
}

// Decompose parses an ID into its components.
func (s *Service) Decompose(_ context.Context, req *gidv1.DecomposeRequest) (*gidv1.DecomposeResponse, error) {
	d := s.gen.Decompose(req.GetId())
	return &gidv1.DecomposeResponse{
		Time:        d.Time.Unix(),
		Sequence:    d.Sequence,
		MachineId:   d.MachineID,
		GeneratedAt: d.Time.Format(time.RFC3339),
	}, nil
}
