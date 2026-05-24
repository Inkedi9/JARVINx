package llm

import (
	"testing"
)

func TestParseDecision_ValidJSON(t *testing.T) {
	raw := `{"analysis": "système stable", "action": "log", "reason": "rien à signaler"}`

	d, err := ParseDecision(raw)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if d.Action != "log" {
		t.Errorf("expected action 'log', got '%s'", d.Action)
	}
	if d.Analysis == "" {
		t.Error("expected non-empty analysis")
	}
}

func TestParseDecision_WithMarkdown(t *testing.T) {
	raw := "```json\n{\"analysis\": \"stable\", \"action\": \"log\", \"reason\": \"ok\"}\n```"

	d, err := ParseDecision(raw)

	if err != nil {
		t.Fatalf("expected no error with markdown, got: %v", err)
	}
	if d.Action != "log" {
		t.Errorf("expected 'log', got '%s'", d.Action)
	}
}

func TestParseDecision_TextAroundJSON(t *testing.T) {
	raw := `Voici mon analyse :
{"analysis": "CPU élevé", "action": "alert", "reason": "seuil dépassé"}
Fin de l'analyse.`

	d, err := ParseDecision(raw)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if d.Action != "alert" {
		t.Errorf("expected 'alert', got '%s'", d.Action)
	}
}

func TestParseDecision_InvalidAction(t *testing.T) {
	raw := `{"analysis": "test", "action": "reboot", "reason": "test"}`

	_, err := ParseDecision(raw)

	if err == nil {
		t.Fatal("expected error for invalid action, got nil")
	}
}

func TestParseDecision_MissingAnalysis(t *testing.T) {
	raw := `{"action": "log", "reason": "ok"}`

	_, err := ParseDecision(raw)

	if err == nil {
		t.Fatal("expected error for missing analysis, got nil")
	}
}

func TestParseDecision_MalformedJSON(t *testing.T) {
	raw := `{"analysis": "test", "action": "log" "reason": "virgule manquante"}`

	d, err := ParseDecision(raw)

	// On doit avoir une erreur MAIS aussi un fallback valide
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
	if d.Action != "log" {
		t.Errorf("fallback should have action 'log', got '%s'", d.Action)
	}
}

func TestParseDecision_UppercaseAction(t *testing.T) {
	raw := `{"analysis": "stable", "action": "LOG", "reason": "ok"}`

	d, err := ParseDecision(raw)

	if err != nil {
		t.Fatalf("expected no error for uppercase action, got: %v", err)
	}
	if d.Action != "log" {
		t.Errorf("expected normalized 'log', got '%s'", d.Action)
	}
}

func TestFallbackDecision(t *testing.T) {
	d := fallbackDecision("raw llm output")

	if d.Action != "log" {
		t.Errorf("fallback should default to 'log', got '%s'", d.Action)
	}
	if d.Analysis == "" {
		t.Error("fallback analysis should not be empty")
	}
}
