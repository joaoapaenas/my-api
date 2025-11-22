package logger

import (
	"log/slog"
	"os"
)

// Init configures the global logger.
// env: "development" (text logs) or "production" (json logs)
func Init(env string) {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo, // Change to slog.LevelDebug for more verbosity
	}

	if env == "production" {
		// JSON is machine-readable (required for AWS CloudWatch)
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		// Text is human-readable (nice for local dev)
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
}
