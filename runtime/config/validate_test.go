package config

import (
	"testing"
	"time"
)

func TestValidate_DefaultConfigIsValid(t *testing.T) {
	cfg := Default()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("default config should be valid, got: %v", err)
	}
}

func TestValidate_ThresholdAbove100(t *testing.T) {
	cfg := Default()
	cfg.CPUAlertThreshold = 150.0

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for threshold > 100")
	}
}

func TestValidate_ThresholdZero(t *testing.T) {
	cfg := Default()
	cfg.RAMAlertThreshold = 0

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for threshold = 0")
	}
}

func TestValidate_ThresholdNegative(t *testing.T) {
	cfg := Default()
	cfg.DiskAlertThreshold = -10

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for negative threshold")
	}
}

func TestValidate_IntervalTooShort(t *testing.T) {
	cfg := Default()
	cfg.Interval = 1 * time.Second

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for interval < 5s")
	}
}

func TestValidate_IntervalTooLong(t *testing.T) {
	cfg := Default()
	cfg.Interval = 2 * time.Hour

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for interval > 1h")
	}
}

func TestValidate_InvalidPort(t *testing.T) {
	cfg := Default()
	cfg.WebPort = 80

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for port < 1024")
	}
}

func TestValidate_EmptyModel(t *testing.T) {
	cfg := Default()
	cfg.Model = ""

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for empty model")
	}
}

func TestValidate_EmptyLogFile(t *testing.T) {
	cfg := Default()
	cfg.LogFile = ""

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for empty log file")
	}
}

func TestValidate_MultipleErrors(t *testing.T) {
	cfg := Default()
	cfg.CPUAlertThreshold = 150
	cfg.RAMAlertThreshold = -5
	cfg.WebPort = 80

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected multiple errors")
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatal("expected *ValidationError type")
	}
	if len(ve.Errors) < 3 {
		t.Errorf("expected at least 3 errors, got %d : %v", len(ve.Errors), ve.Errors)
	}
}

func TestValidate_AlertMinCycles(t *testing.T) {
	cfg := Default()
	cfg.AlertMinCycles = 0

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for AlertMinCycles = 0")
	}
}

func TestFromEnv_OverridesModel(t *testing.T) {
	t.Setenv("JARVINX_MODEL", "qwen2.5:7b")

	cfg := Default()
	cfg.FromEnv()

	if cfg.Model != "qwen2.5:7b" {
		t.Errorf("expected model 'qwen2.5:7b', got '%s'", cfg.Model)
	}
}

func TestFromEnv_OverridesThresholds(t *testing.T) {
	t.Setenv("JARVINX_CPU_THRESHOLD", "75")
	t.Setenv("JARVINX_RAM_THRESHOLD", "85")
	t.Setenv("JARVINX_DISK_THRESHOLD", "80")

	cfg := Default()
	cfg.FromEnv()

	if cfg.CPUAlertThreshold != 75.0 {
		t.Errorf("expected CPU 75.0, got %.1f", cfg.CPUAlertThreshold)
	}
	if cfg.RAMAlertThreshold != 85.0 {
		t.Errorf("expected RAM 85.0, got %.1f", cfg.RAMAlertThreshold)
	}
	if cfg.DiskAlertThreshold != 80.0 {
		t.Errorf("expected Disk 80.0, got %.1f", cfg.DiskAlertThreshold)
	}
}

func TestFromEnv_OverridesInterval(t *testing.T) {
	t.Setenv("JARVINX_INTERVAL", "30s")

	cfg := Default()
	cfg.FromEnv()

	if cfg.Interval != 30*time.Second {
		t.Errorf("expected 30s, got %v", cfg.Interval)
	}
}

func TestFromEnv_InvalidThresholdIgnored(t *testing.T) {
	t.Setenv("JARVINX_CPU_THRESHOLD", "not-a-number")

	cfg := Default()
	original := cfg.CPUAlertThreshold
	cfg.FromEnv()

	// Valeur invalide → valeur par défaut conservée
	if cfg.CPUAlertThreshold != original {
		t.Errorf("invalid value should be ignored, expected %.1f, got %.1f",
			original, cfg.CPUAlertThreshold)
	}
}

func TestFromEnv_InvalidIntervalIgnored(t *testing.T) {
	t.Setenv("JARVINX_INTERVAL", "invalid")

	cfg := Default()
	original := cfg.Interval
	cfg.FromEnv()

	if cfg.Interval != original {
		t.Errorf("invalid interval should be ignored, expected %v, got %v",
			original, cfg.Interval)
	}
}

func TestFromEnv_OverridesPort(t *testing.T) {
	t.Setenv("JARVINX_PORT", "9090")

	cfg := Default()
	cfg.FromEnv()

	if cfg.WebPort != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.WebPort)
	}
}

func TestFromEnv_EmptyVarsKeepDefaults(t *testing.T) {
	// Aucune variable définie → tout reste par défaut
	cfg := Default()
	defaults := Default()
	cfg.FromEnv()

	if cfg.Model != defaults.Model {
		t.Errorf("empty env should keep default model")
	}
	if cfg.Interval != defaults.Interval {
		t.Errorf("empty env should keep default interval")
	}
}
