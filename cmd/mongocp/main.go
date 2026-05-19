package main

import (
	"log/slog"
	"os"

	"github.com/kstieger/mongocp/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
		logger.Error("mongocp failed", "err", err)
		os.Exit(1)
	}
}
