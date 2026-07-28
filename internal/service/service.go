// Package service contains gid-service business logic.
//
// Layering contract (see golang-service-development skill §2):
//   - This is the SERVICE ROOT. It holds the Service struct + New + Start/Stop
//   - resource resolve helpers + one-line facade methods (one per RPC).
//   - Business logic lives in SUBPACKAGES (internal/service/<domain>/). This
//     file does NOT contain RPC implementations — only delegations.
//   - handler calls service.X; service.X is a one-line facade that calls
//     s.<domain>.X in the subpackage. handler never imports the subpackage.
//   - Service methods take proto types DIRECTLY and return proto types.
//   - Resources are constructed here from cfg (or injected via option in
//     future) and passed to subpackage constructors.
//
// Lifecycle:
//   - Service holds a *lifecycle.Manager. gid-service currently owns no
//     closable resources (snowflake generator is pure compute), so the
//     manager stays empty — Start/Stop are well-defined no-ops. Future
//     resources (Redis, cron, etc.) register Stoppers here.
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/servekit/go-common/lifecycle"

	gidv1 "github.com/servekit/gid-service/gen/gid/v1"
	"github.com/servekit/gid-service/internal/jobs"
	"github.com/servekit/gid-service/internal/provider/snowflake"
	"github.com/servekit/gid-service/internal/service/gid"
	"github.com/servekit/gid-service/internal/version"
	"github.com/servekit/gid-service/pkg/config"
	"github.com/servekit/gid-service/pkg/option"
)

// Service holds gid-service business state.
//
// gen is the snowflake generator kept on the root Service for resolveXxx
// helpers — it's the same instance injected into the gid subpackage. The gid
// subpackage does NOT reference this struct.
type Service struct {
	cfg *config.Config
	mgr *lifecycle.Manager

	gen *snowflake.Generator

	// One field per domain subpackage.
	gid *gid.Service

	// startedAt is set once in New; Ping returns it for uptime.
	startedAt int64
}

// New constructs a Service from config and functional options.
//
// The snowflake generator is built from cfg.Snowflake. It has no Close (pure
// compute), so it is NOT registered with mgr. On partial failure, already-
// registered components (currently none) are stopped via mgr.Stop() before
// returning the error.
func New(cfg *config.Config, opts ...option.Option) (*Service, error) {
	// Reserved for future injectable resources; currently Options is empty.
	option.Apply(opts...)

	if err := cfg.ValidateSnowflake(); err != nil {
		return nil, err
	}

	mgr := lifecycle.NewManager()

	gen, err := snowflake.New(cfg.Snowflake.MachineID, cfg.Snowflake.StartTime)
	if err != nil {
		if cerr := mgr.Stop(); cerr != nil {
			err = errors.Join(err, cerr)
		}
		return nil, err
	}

	svc := &Service{
		cfg: cfg,
		mgr: mgr,
		gen: gen,
		gid: gid.New(gen),
		startedAt: time.Now().UnixMilli(),
	}

	// Cron is scaffold-only — gid-service has no periodic jobs yet, so
	// setupJobs is intentionally NOT called. To enable:
	//   1. Add a `cron:` section to config.yaml (populates cfg.Cron).
	//   2. Uncomment the svc.setupJobs() call below.
	//   3. Register jobs via scheduler.AddFunc inside setupJobs.
	//
	// if err := svc.setupJobs(); err != nil {
	//     if cerr := mgr.Stop(); cerr != nil {
	//         err = errors.Join(err, fmt.Errorf("rollback: %w", cerr))
	//     }
	//     return nil, err
	// }

	return svc, nil
}

// Start starts all owned components concurrently. No-op when nothing is
// registered with mgr (current state — no jobs, no closable resources).
func (s *Service) Start() error { return s.mgr.Start() }

// Stop stops all owned components in reverse registration order.
func (s *Service) Stop() error { return s.mgr.Stop() }

// Ping is a health-check RPC. Returns only public, non-sensitive info.
func (s *Service) Ping(ctx context.Context) (*gidv1.Pong, error) {
	v := version.Get()
	return &gidv1.Pong{
		Service:   "gid-service",
		Version:   v.Version,
		GitCommit: v.GitCommit,
		GitBranch: v.GitBranch,
		BuildTime: v.BuildTime,
		GoVersion: v.GoVersion,
		Status:    "SERVING",
		Now:       time.Now().UnixMilli(),
		StartedAt: s.startedAt,
	}, nil
}

// --- facade methods (one per RPC, delegate to subpackage) ---

// NextID delegates to the gid subpackage.
func (s *Service) NextID(ctx context.Context, req *gidv1.NextIDRequest) (*gidv1.NextIDResponse, error) {
	return s.gid.NextID(ctx, req)
}

// BatchNextID delegates to the gid subpackage.
func (s *Service) BatchNextID(ctx context.Context, req *gidv1.BatchNextIDRequest) (*gidv1.BatchNextIDResponse, error) {
	return s.gid.BatchNextID(ctx, req)
}

// Decompose delegates to the gid subpackage.
func (s *Service) Decompose(ctx context.Context, req *gidv1.DecomposeRequest) (*gidv1.DecomposeResponse, error) {
	return s.gid.Decompose(ctx, req)
}

// --- internal helpers ---

// setupJobs builds the jobs.Scheduler, registers it on s.mgr (so its
// lifecycle is managed alongside other owned resources), and wires periodic
// jobs. Scaffold-only — New does NOT call this; cron stays out of the runtime
// path entirely. See the comment in New for the activation steps.
func (s *Service) setupJobs() error {
	scheduler, err := jobs.New(&jobs.Deps{
		Config: s.cfg.Cron,
	})
	if err != nil {
		return fmt.Errorf("init jobs: %w", err)
	}
	s.mgr.Add("jobs", scheduler)

	// Register periodic jobs here, e.g.:
	//
	//   if err := scheduler.AddFunc("*/5 * * * *", func() {
	//       ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	//       defer cancel()
	//       if err := s.gid.SomePeriodicOp(ctx); err != nil {
	//           slog.Error("gid periodic op", "error", err)
	//       }
	//   }); err != nil {
	//       return fmt.Errorf("register gid periodic op: %w", err)
	//   }
	return nil
}

// Keep setupJobs live (compiled, IDE-navigable, lint-checked) even though
// New does not call it yet. Drop this line once the first cron job lands.
var _ = (*Service).setupJobs
