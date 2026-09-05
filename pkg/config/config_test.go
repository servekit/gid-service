package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/servekit/go-common/logging"

	"github.com/servekit/gid-service/internal/provider/snowflake"
)

func TestValidate(t *testing.T) {
	valid := func() *Config {
		return &Config{
			Server: &ServerConfig{
				GRPCAddr: ":19091",
			},
			Snowflake: &SnowflakeConfig{
				MachineID: 1,
				StartTime: time.Now().Add(-time.Hour),
			},
			Log: &logging.Config{Level: "info", Format: "json"},
		}
	}

	tests := []struct {
		name    string
		cfg     *Config
		mutate  func(*Config)
		wantErr bool
	}{
		{name: "valid", cfg: valid()},
		{name: "nil config", cfg: nil, wantErr: true},
		{name: "missing server", cfg: valid(), mutate: func(cfg *Config) {
			cfg.Server = nil
		}, wantErr: true},
		{name: "missing grpc addr", cfg: valid(), mutate: func(cfg *Config) {
			cfg.Server.GRPCAddr = ""
		}, wantErr: true},
		{name: "missing snowflake", cfg: valid(), mutate: func(cfg *Config) {
			cfg.Snowflake = nil
		}, wantErr: true},
		{name: "machine id too small", cfg: valid(), mutate: func(cfg *Config) {
			cfg.Snowflake.MachineID = 0
		}, wantErr: true},
		{name: "machine id too large", cfg: valid(), mutate: func(cfg *Config) {
			cfg.Snowflake.MachineID = snowflake.MaxMachineID + 1
		}, wantErr: true},
		{name: "missing start time", cfg: valid(), mutate: func(cfg *Config) {
			cfg.Snowflake.StartTime = time.Time{}
		}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mutate != nil {
				tt.mutate(tt.cfg)
			}
			err := tt.cfg.Validate()
			if tt.wantErr && err == nil {
				t.Fatal("Validate() error = nil, want error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
		})
	}
}

// loadEnvFile parses a dotenv file and sets each KEY=VALUE into the test
// environment. Mirrors docker-compose `env_file` semantics: blank lines and
// lines starting with '#' are skipped, inline comments are NOT supported (the
// value is everything after the first '='), and quotes are not stripped.
func loadEnvFile(t *testing.T, path string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read env file: %v", err)
	}
	for _, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.IndexByte(line, '=')
		if idx < 0 {
			t.Fatalf("malformed env line (no '='): %q", line)
		}
		key := strings.TrimSpace(line[:idx])
		if key == "" {
			t.Fatalf("empty key in env line: %q", line)
		}
		t.Setenv(key, line[idx+1:])
	}
}

// TestExampleConfigsAreLoadable guards that config.example.yaml + .env.example
// stay self-consistent: loads the YAML with every ${VAR} resolved from
// .env.example and asserts Load() + Validate() succeed with real (expanded)
// values. Catches drift in either direction — a ${VAR} in the YAML with no
// matching var in .env.example, or an env value that breaks parsing.
func TestExampleConfigsAreLoadable(t *testing.T) {
	root := filepath.Join("..", "..")
	loadEnvFile(t, filepath.Join(root, ".env.example"))
	t.Setenv("GID_SERVICE_CONFIG", filepath.Join(root, "config.example.yaml"))

	cfg, err := Load()
	if err != nil {
		t.Fatalf("config.example.yaml + .env.example must load and validate: %v", err)
	}

	// Spot-check that ${VAR} was actually expanded, not left literal.
	if cfg.Server.GRPCAddr != ":19091" {
		t.Errorf("Server.GRPCAddr = %q, want %q", cfg.Server.GRPCAddr, ":19091")
	}
	if cfg.Snowflake.MachineID != 1 {
		t.Errorf("Snowflake.MachineID = %d, want 1", cfg.Snowflake.MachineID)
	}
	wantStart := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	if !cfg.Snowflake.StartTime.Equal(wantStart) {
		t.Errorf("Snowflake.StartTime = %v, want %v", cfg.Snowflake.StartTime, wantStart)
	}
}
