package logger

import (
	"log/slog"
	"os"
)

func Init() {
	// Use JSONHandler for production (CloudWatch friendly)
	// Use TextHandler for local dev if preferred, but JSON is safer for consistency.
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)
}
