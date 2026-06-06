package log

import (
	"log/slog"
	"os"

	"github.com/ThreeDotsLabs/humanslog"
)

func Init(level slog.Level) {
	handlerOpts := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	if os.Getenv("APP_ENV") == "production" {
		handler = slog.NewJSONHandler(os.Stderr, handlerOpts)
	} else {
		handler = humanslog.NewHandler(os.Stderr, &humanslog.Options{
			HandlerOptions: handlerOpts,
			TimeFormat:     "[15:04:05.000]",
		})
	}

	slog.SetDefault(slog.New(handler))
}
