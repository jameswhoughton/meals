package meals

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"slices"
)

type CustomHandler struct {
	h            slog.Handler
	sourceLevels []slog.Level
}

func (c *CustomHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return c.h.Enabled(ctx, level)
}

func (c *CustomHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &CustomHandler{
		h:            c.h.WithAttrs(attrs),
		sourceLevels: c.sourceLevels,
	}
}

func (c *CustomHandler) WithGroup(name string) slog.Handler {
	return &CustomHandler{
		h:            c.h.WithGroup(name),
		sourceLevels: c.sourceLevels,
	}
}

func (c *CustomHandler) Handle(ctx context.Context, r slog.Record) error {
	if slices.Contains(c.sourceLevels, r.Level) {
		if _, file, line, ok := runtime.Caller(3); ok {
			r.AddAttrs(slog.String("source", fmt.Sprintf("%s:%d", file, line)))
		}
	}

	return c.h.Handle(ctx, r)
}

func NewApplicationLogger(logFile *os.File) *slog.Logger {
	handler := CustomHandler{
		h:            slog.NewJSONHandler(logFile, nil),
		sourceLevels: []slog.Level{slog.LevelDebug, slog.LevelError},
	}

	return slog.New(&handler)
}
