package log

import (
	"log/slog"
	"os"
)

func Init(level slog.Level) {
	opts := &slog.HandlerOptions{
		Level: level,
	}
	handler := slog.NewTextHandler(os.Stderr, opts)

	slog.SetDefault(slog.New(handler))
}
