package jxlog

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"
)

var Log *slog.Logger

func Init(debug bool) {
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}
	handler := NewJXHandler(os.Stdout, level)
	Log = slog.New(handler)
	slog.SetDefault(Log)
}

func Info(tag, msg string, args ...any)  { emit(slog.LevelInfo, tag, msg, args...) }
func Warn(tag, msg string, args ...any)  { emit(slog.LevelWarn, tag, msg, args...) }
func Error(tag, msg string, args ...any) { emit(slog.LevelError, tag, msg, args...) }
func Debug(tag, msg string, args ...any) { emit(slog.LevelDebug, tag, msg, args...) }

func emit(level slog.Level, tag, msg string, args ...any) {
	if Log == nil {
		return
	}
	attrs := make([]any, 0, 1+len(args))
	attrs = append(attrs, slog.String("tag", ColorTag(tag)))
	attrs = append(attrs, args...)
	Log.Log(context.Background(), level, msg, attrs...)
}

// JXHandler — handler slog custom style JARVINx
type JXHandler struct {
	out   io.Writer
	level slog.Level
}

func NewJXHandler(out io.Writer, level slog.Level) *JXHandler {
	return &JXHandler{out: out, level: level}
}

func (h *JXHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *JXHandler) Handle(_ context.Context, r slog.Record) error {
	levelStr := h.formatLevel(r.Level)
	timestamp := r.Time.Format("15:04:05")

	tag := ""
	var extras []string

	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "tag" {
			tag = a.Value.String()
		} else {
			extras = append(extras, fmt.Sprintf("%s=%v", a.Key, a.Value))
		}
		return true
	})

	var line string
	if tag != "" {
		line = fmt.Sprintf("\033[90m%s\033[0m %s %s %s",
			timestamp, levelStr, tag, r.Message)
	} else {
		line = fmt.Sprintf("\033[90m%s\033[0m %s %s",
			timestamp, levelStr, r.Message)
	}

	if len(extras) > 0 {
		line += " \033[90m" + strings.Join(extras, " ") + "\033[0m"
	}

	_, _ = fmt.Fprintln(h.out, line)
	return nil
}

func (h *JXHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *JXHandler) WithGroup(_ string) slog.Handler {
	return h
}

func (h *JXHandler) formatLevel(level slog.Level) string {
	switch {
	case level >= slog.LevelError:
		return "\033[31m[ ERROR ]\033[0m"
	case level >= slog.LevelWarn:
		return "\033[33m[ WARN  ]\033[0m"
	case level >= slog.LevelInfo:
		return "\033[97m[ INFO  ]\033[0m"
	default:
		return "\033[90m[ DEBUG ]\033[0m"
	}
}

func ColorTag(tag string) string {
	colors := map[string]string{
		"ORCHESTRATOR": "\033[34m",
		"SCHEDULER":    "\033[36m",
		"REGISTRY":     "\033[36m",
		"WEB":          "\033[36m",
		"CLI":          "\033[34m",
		"SYSTEM AGENT": "\033[35m",
		"AGENT":        "\033[35m",
		"ALERT":        "\033[31m",
		"EXEC":         "\033[33m",
		"STATE":        "\033[90m",
		"OK":           "\033[32m",
		"WARN":         "\033[33m",
		"ERROR":        "\033[31m",
		"OLLAMA":       "\033[32m",
		"JARVINX":      "\033[36m",
	}
	color, ok := colors[tag]
	if !ok {
		color = "\033[97m"
	}
	return fmt.Sprintf("%s[ %s ]\033[0m", color, tag)
}

func FormatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	}
	return fmt.Sprintf("%dm %ds", m, s)
}

// CaptureOutput — utilitaire pour les tests
func CaptureOutput(fn func()) string {
	var buf bytes.Buffer
	old := Log
	Log = slog.New(NewJXHandler(&buf, slog.LevelDebug))
	fn()
	Log = old
	return buf.String()
}
