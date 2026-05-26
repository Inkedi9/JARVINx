package jxlog

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestInit_SetsGlobalLogger(t *testing.T) {
	Init(false)
	if Log == nil {
		t.Error("Init should set global Log")
	}
}

func TestInit_DebugMode(t *testing.T) {
	Init(true)
	if Log == nil {
		t.Error("Init with debug=true should set global Log")
	}
}

func TestJXHandler_InfoVisible(t *testing.T) {
	var buf bytes.Buffer
	h := NewJXHandler(&buf, slog.LevelInfo)
	logger := slog.New(h)

	logger.Info("hello jarvinx")

	if !strings.Contains(buf.String(), "hello jarvinx") {
		t.Errorf("expected message in output, got: %s", buf.String())
	}
}

func TestJXHandler_DebugFilteredAtInfoLevel(t *testing.T) {
	var buf bytes.Buffer
	h := NewJXHandler(&buf, slog.LevelInfo)
	logger := slog.New(h)

	logger.Debug("should not appear")

	if buf.Len() > 0 {
		t.Errorf("debug should be filtered at Info level, got: %s", buf.String())
	}
}

func TestJXHandler_DebugVisibleAtDebugLevel(t *testing.T) {
	var buf bytes.Buffer
	h := NewJXHandler(&buf, slog.LevelDebug)
	logger := slog.New(h)

	logger.Debug("debug visible")

	if !strings.Contains(buf.String(), "debug visible") {
		t.Errorf("debug should be visible at Debug level, got: %s", buf.String())
	}
}

func TestJXHandler_ErrorLevel(t *testing.T) {
	var buf bytes.Buffer
	h := NewJXHandler(&buf, slog.LevelInfo)
	logger := slog.New(h)

	logger.Error("something broke")

	output := buf.String()
	if !strings.Contains(output, "something broke") {
		t.Errorf("expected error in output, got: %s", output)
	}
	if !strings.Contains(output, "ERROR") {
		t.Errorf("expected ERROR label, got: %s", output)
	}
}

func TestColorTag_KnownTags(t *testing.T) {
	tags := []string{"REGISTRY", "AGENT", "ALERT", "EXEC", "JARVINX"}
	for _, tag := range tags {
		result := ColorTag(tag)
		if !strings.Contains(result, tag) {
			t.Errorf("ColorTag(%s) should contain tag name, got: %s", tag, result)
		}
	}
}

func TestColorTag_UnknownTag(t *testing.T) {
	result := ColorTag("MYSTERY_TAG")
	if !strings.Contains(result, "MYSTERY_TAG") {
		t.Error("unknown tag should still be included in output")
	}
}

func TestHelpers_NilLogNoPanic(t *testing.T) {
	Log = nil
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("nil Log should not panic, got: %v", r)
		}
		Init(false)
	}()

	Info("TEST", "message")
	Warn("TEST", "message")
	Error("TEST", "message")
	Debug("TEST", "message")
}

func TestCaptureOutput(t *testing.T) {
	Init(false)
	output := CaptureOutput(func() {
		Info("TEST", "captured message")
	})
	if !strings.Contains(output, "captured message") {
		t.Errorf("CaptureOutput should capture log output, got: %s", output)
	}
}
