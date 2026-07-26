// Command server is the gid-service gRPC + HTTP entry point.
package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/servekit/go-common/logging"
	"github.com/servekit/go-common/signalx"

	pkg "github.com/servekit/gid-service/pkg"
	"github.com/servekit/gid-service/pkg/config"
)

func main() {
	// Load .env when present so local binary runs pick up the same values
	// docker-compose injects. Missing .env (docker/prod) is not an error.
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, "warning: failed to load .env:", err)
	}

	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}
	logging.Setup(cfg.Log)

	srv, err := pkg.NewServer(cfg)
	if err != nil {
		slog.Error("init server", "error", err)
		os.Exit(1)
	}

	if err := signalx.RunWithForceQuit(srv); err != nil {
		slog.Error("run server", "error", err)
		os.Exit(1)
	}
}
