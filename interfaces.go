package gitfresh

import (
	"context"
	"log/slog"
)

type AppLogger interface {
	Info(msg string, args ...any)
	Error(msg string, args ...any)
	Warn(msg string, args ...any)
	Debug(msg string, args ...any)
	LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr)
}
