package app

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sort"
	"strings"
)

const fatalLogLevel = slog.Level(12)

// SetupLogger configures slog with the given log level.
func SetupLogger(level string) *slog.Logger {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	case "fatal":
		lvl = fatalLogLevel
	default:
		lvl = slog.LevelInfo
	}
	h := &colorHandler{
		level: lvl,
		out:   os.Stderr,
	}
	return slog.New(h)
}

type colorHandler struct {
	level slog.Level
	out   io.Writer
}

func (h *colorHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *colorHandler) Handle(_ context.Context, record slog.Record) error {
	if !h.Enabled(context.Background(), record.Level) {
		return nil
	}

	levelColor, levelLabel := levelStyle(record.Level)
	message := colorize(ansiBrightWhite, record.Message)
	attrs := make([]string, 0, record.NumAttrs())
	record.Attrs(func(attr slog.Attr) bool {
		attrs = append(attrs, formatAttr(attr))
		return true
	})
	sort.Strings(attrs)

	line := fmt.Sprintf(
		"%s %s %s",
		colorize(ansiDim, record.Time.Format("15:04:05")),
		colorize(levelColor, padLevel(levelLabel)),
		message,
	)
	if len(attrs) > 0 {
		line += " " + strings.Join(attrs, " ")
	}
	_, err := fmt.Fprintln(h.out, line)
	return err
}

func (h *colorHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &withAttrsHandler{
		next:  h,
		attrs: append([]slog.Attr(nil), attrs...),
	}
}

func (h *colorHandler) WithGroup(name string) slog.Handler {
	return &withGroupHandler{
		next:  h,
		group: name,
	}
}

type withAttrsHandler struct {
	next  *colorHandler
	attrs []slog.Attr
}

func (h *withAttrsHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *withAttrsHandler) Handle(ctx context.Context, record slog.Record) error {
	clone := slog.NewRecord(record.Time, record.Level, record.Message, record.PC)
	for _, attr := range h.attrs {
		clone.AddAttrs(attr)
	}
	record.Attrs(func(attr slog.Attr) bool {
		clone.AddAttrs(attr)
		return true
	})
	return h.next.Handle(ctx, clone)
}

func (h *withAttrsHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	combined := append(append([]slog.Attr(nil), h.attrs...), attrs...)
	return &withAttrsHandler{next: h.next, attrs: combined}
}

func (h *withAttrsHandler) WithGroup(name string) slog.Handler {
	return &groupedHandler{base: h, group: name}
}

type withGroupHandler struct {
	next  *colorHandler
	group string
}

func (h *withGroupHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *withGroupHandler) Handle(ctx context.Context, record slog.Record) error {
	clone := slog.NewRecord(record.Time, record.Level, record.Message, record.PC)
	record.Attrs(func(attr slog.Attr) bool {
		clone.AddAttrs(slog.Group(h.group, attr))
		return true
	})
	return h.next.Handle(ctx, clone)
}

func (h *withGroupHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	grouped := make([]slog.Attr, 0, len(attrs))
	for _, attr := range attrs {
		grouped = append(grouped, slog.Group(h.group, attr))
	}
	return &withAttrsHandler{next: h.next, attrs: grouped}
}

func (h *withGroupHandler) WithGroup(name string) slog.Handler {
	return &withGroupHandler{next: h.next, group: h.group + "." + name}
}

type groupedHandler struct {
	base  slog.Handler
	group string
}

func (h *groupedHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.base.Enabled(ctx, level)
}

func (h *groupedHandler) Handle(ctx context.Context, record slog.Record) error {
	clone := slog.NewRecord(record.Time, record.Level, record.Message, record.PC)
	record.Attrs(func(attr slog.Attr) bool {
		clone.AddAttrs(slog.Group(h.group, attr))
		return true
	})
	return h.base.Handle(ctx, clone)
}

func (h *groupedHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	grouped := make([]slog.Attr, 0, len(attrs))
	for _, attr := range attrs {
		grouped = append(grouped, slog.Group(h.group, attr))
	}
	return h.base.WithAttrs(grouped)
}

func (h *groupedHandler) WithGroup(name string) slog.Handler {
	return &groupedHandler{base: h.base, group: h.group + "." + name}
}

func formatAttr(attr slog.Attr) string {
	value := attr.Value.Resolve()
	if value.Kind() == slog.KindGroup {
		parts := make([]string, 0, len(value.Group()))
		for _, item := range value.Group() {
			parts = append(parts, formatAttr(item))
		}
		return colorize(ansiCyan, attr.Key) + "={" + strings.Join(parts, " ") + "}"
	}

	return colorize(ansiCyan, attr.Key) + "=" + colorize(ansiWhite, fmt.Sprint(value.Any()))
}

func levelStyle(level slog.Level) (string, string) {
	switch {
	case level <= slog.LevelDebug:
		return ansiBlue, "DEBUG"
	case level <= slog.LevelInfo:
		return ansiGreen, "INFO"
	case level <= slog.LevelWarn:
		return ansiYellow, "WARN"
	default:
		return ansiRed, "ERROR"
	}
}

func padLevel(level string) string {
	return fmt.Sprintf("%-5s", level)
}

func colorize(code, text string) string {
	return code + text + ansiReset
}

const (
	ansiReset       = "\033[0m"
	ansiDim         = "\033[2m"
	ansiBlue        = "\033[34m"
	ansiGreen       = "\033[32m"
	ansiYellow      = "\033[33m"
	ansiRed         = "\033[31m"
	ansiCyan        = "\033[36m"
	ansiWhite       = "\033[37m"
	ansiBrightWhite = "\033[97m"
)
