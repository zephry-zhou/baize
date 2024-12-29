package internal

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

type Logger struct {
	l *slog.Logger
	f *os.File
}

func NewStreamLogger(level slog.Level) *Logger {
	h := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				a.Value = slog.StringValue(a.Value.Time().Format("2006-01-02 15:04:05"))
			} else if a.Key == slog.SourceKey {
				source := a.Value.Any().(*slog.Source)
				file := filepath.Base(source.File)
				line := source.Line
				a.Value = slog.GroupValue(
					slog.String("file", file),
					slog.Int("line", line),
				)
			}
			return a
		},
	}))
	return &Logger{l: h.WithGroup("baize")}
}

func NewFileLogger(level slog.Level, path string) *Logger {

	file, _ := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	h := slog.New(slog.NewJSONHandler(file, &slog.HandlerOptions{
		AddSource: true,
		Level:     level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				a.Value = slog.StringValue(a.Value.Time().Format("2006-01-02 15:04:05"))
			} else if a.Key == slog.SourceKey {
				source := a.Value.Any().(*slog.Source)
				file := filepath.Base(source.File)
				line := source.Line
				a.Value = slog.GroupValue(
					slog.String("file", file),
					slog.Int("line", line),
				)
			}
			return a
		},
	}))
	return &Logger{l: h}
}

func (l *Logger) Debug(msg string, args ...any) {
	l.Log(context.Background(), slog.LevelDebug, msg, args...)
}

func (l *Logger) Info(msg string, args ...any) {
	l.Log(context.Background(), slog.LevelInfo, msg, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	l.Log(context.Background(), slog.LevelWarn, msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	l.Log(context.Background(), slog.LevelError, msg, args...)
}

func (l *Logger) Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	l.log(ctx, level, msg, args...)
}

func (l *Logger) Close() error {
	return l.f.Close()
}

func (l *Logger) log(ctx context.Context, level slog.Level, msg string, args ...any) {
	if !l.l.Enabled(ctx, level) {
		return
	}
	var pc uintptr

	var pcs [1]uintptr
	// skip [runtime.Callers, this function, this function's caller]
	runtime.Callers(4, pcs[:])
	pc = pcs[0]
	r := slog.NewRecord(time.Now(), level, msg, pc)
	r.Add(args...)
	if ctx == nil {
		ctx = context.Background()
	}
	_ = l.l.Handler().Handle(ctx, r)
}
