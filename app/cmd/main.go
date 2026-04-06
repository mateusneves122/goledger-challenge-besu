package main

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"

	"github.com/neves144/goledger-challenge-besu/app/internal/config"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))

	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	if err := run(cfg); err != nil {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}
}
