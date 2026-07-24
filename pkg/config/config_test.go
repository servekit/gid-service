package config

import (
	"testing"
	"time"

	"github.com/servekit/go-common/logging"

	"github.com/servekit/gid-service/internal/provider/snowflake"
)

func TestValidate(t *testing.T) {
	valid := func() *Config {
		return &Config{
			Server: &ServerConfig{
				GRPCAddr: ":9000",
				HTTPAddr: ":8080",
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
		{name: "missing http addr", cfg: valid(), mutate: func(cfg *Config) {
			cfg.Server.HTTPAddr = ""
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
