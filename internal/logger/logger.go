package logger

import (
	"log/slog"
	"os"
)

func init() {
	var logger *slog.Logger
	switch os.Getenv("APP_ENV") {
	case "prod":
		logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	case "dev":
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	default:
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
	slog.SetDefault(logger)
}
