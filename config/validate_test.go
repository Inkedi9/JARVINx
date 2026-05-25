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
