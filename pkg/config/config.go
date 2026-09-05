// Package config provides service configuration loading from YAML files.
package config

import (
	"strings"
	"time"

	"github.com/servekit/go-common/configx"
	"github.com/servekit/go-common/cronx"
	"github.com/servekit/go-common/logging"

	"github.com/servekit/gid-service/internal/provider/snowflake"
	"github.com/servekit/gid-service/pkg/xcodes"
)

// serviceName identifies this binary in config file lookup (/etc/<name>) and
// the <NAME>_CONFIG env var. envPrefix scopes all env overrides under GID_SERVICE_*.
const (
	serviceName = "gid-service"
	envPrefix   = "GID_SERVICE"
)

// Config holds all service configuration.
type Config struct {
	Server    *ServerConfig
	Snowflake *SnowflakeConfig
	Cron      *cronx.Config
	Log       *logging.Config
}

// ServerConfig holds the gRPC server address.
type ServerConfig struct {
	GRPCAddr string `default:":19091"`
}

// SnowflakeConfig holds snowflake ID generator settings.
type SnowflakeConfig struct {
	MachineID int64     `default:"1"`
	StartTime time.Time `default:"2026-01-01T00:00:00Z"`
}

// Load reads configuration from file and environment, expands ${VAR}
// references in the file against the environment, applies defaults, then
// validates and returns a Config.
//
// Env expansion lets config.yaml reference deploy-time values by name
// (e.g. snowflake.machine_id: ${SNOWFLAKE_MACHINE_ID}) instead of holding
// the literal. Unset vars expand to "" (os.ExpandEnv semantics), which
// Validate then surfaces as a missing-required-field error.
func Load() (*Config, error) {
	var cfg Config
	if err := configx.Load(&cfg,
		configx.WithServiceName(serviceName),
		configx.WithEnvPrefix(envPrefix),
		configx.WithExpandEnv(),
	); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate checks that all required configuration values are present and usable.
func (cfg *Config) Validate() error {
	if cfg == nil {
		return xcodes.ErrConfigRequired
	}
	if cfg.Server == nil {
		return xcodes.ErrServerConfigRequired
	}
	if strings.TrimSpace(cfg.Server.GRPCAddr) == "" {
		return xcodes.ErrServerGRPCAddrRequired
	}
	if cfg.Snowflake == nil {
		return cfg.ValidateSnowflake()
	}
	if err := cfg.ValidateSnowflake(); err != nil {
		return err
	}
	return nil
}

// ValidateSnowflake checks the generator configuration required by module usage.
func (cfg *Config) ValidateSnowflake() error {
	if cfg == nil {
		return xcodes.ErrConfigRequired
	}
	if cfg.Snowflake == nil {
		return xcodes.ErrSnowflakeConfigRequired
	}
	if cfg.Snowflake.MachineID < 1 || cfg.Snowflake.MachineID > snowflake.MaxMachineID {
		return xcodes.ErrSnowflakeMachineIDInvalid
	}
	if cfg.Snowflake.StartTime.IsZero() {
		return xcodes.ErrSnowflakeStartTimeRequired
	}
	if cfg.Snowflake.StartTime.After(time.Now()) {
		return xcodes.ErrSnowflakeStartTimeFuture
	}
	return nil
}
