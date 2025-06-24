package logger

import (
	"log/slog"
	"os"

	"pg-bulk-flow/internal/config"
)

func SetupDefault(cfg config.Log) {
	if cfg.PlainText {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: cfg.Level})))
	} else {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: cfg.Level})))
	}
}
