package llm

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// ParseResult contient le résultat du parsing avec des métadonnées
type ParseResult struct {
	Decision Decision
	Raw      string
	Attempts int
	Cleaned  string
}

// Decision est le schema attendu — on valide chaque champ
type Decision struct {
	Analysis string `json:"analysis"`
	Action   string `json:"action"`
	Command  string `json:"command,omitempty"`
	Reason   string `json:"reason"`
}

// Actions valides — whitelist stricte
var validActions = map[string]bool{
	"log":     true,
	"alert":   true,
	"suggest": true,
	"execute": true,
}

// ErrInvalidDecision représente un échec de parsing non récupérable
type ErrInvalidDecision struct {
	Raw    string
	Reason string
}

func (e ErrInvalidDecision) Error() string {
	return fmt.Sprintf("invalid decision: %s | raw: %.100s", e.Reason, e.Raw)
}

// ParseDecision tente d'extraire une Decision valide depuis une string LLM
func ParseDecision(raw string) (Decision, error) {
	// Étape 1 — nettoyage basique
	cleaned := cleanLLMOutput(raw)

	// Étape 2 — tentative directe
	if d, err := parseAndValidate(cleaned); err == nil {
		return d, nil
	}

	// Étape 3 — extraction par regex (LLM a mis du texte autour du JSON)
	extracted, err := extractJSON(cleaned)
	if err != nil {
		return fallbackDecision(raw), ErrInvalidDecision{
			Raw:    raw,
			Reason: "no valid JSON found",
		}
	}

	d, err := parseAndValidate(extracted)
	if err != nil {
		return fallbackDecision(raw), ErrInvalidDecision{
			Raw:    raw,
			Reason: err.Error(),
		}
	}

	return d, nil
}

// cleanLLMOutput supprime les artefacts courants des LLM
func cleanLLMOutput(s string) string {
	s = strings.TrimSpace(s)

	// Strip blocs markdown
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")

	// Strip texte avant le premier {
	if idx := strings.Index(s, "{"); idx > 0 {
		s = s[idx:]
	}

	// Strip texte après le dernier }
	if idx := strings.LastIndex(s, "}"); idx >= 0 && idx < len(s)-1 {
		s = s[:idx+1]
	}

	return strings.TrimSpace(s)
}

// extractJSON cherche un bloc JSON valide avec une regex
func extractJSON(s string) (string, error) {
	// Cherche tout ce qui ressemble à un objet JSON
	re := regexp.MustCompile(`\{[^{}]*\}`)
	matches := re.FindAllString(s, -1)

	for _, match := range matches {
		var test map[string]any
		if err := json.Unmarshal([]byte(match), &test); err == nil {
			return match, nil
		}
	}

	return "", fmt.Errorf("no JSON object found in: %.100s", s)
}

// parseAndValidate parse le JSON et valide le schema
func parseAndValidate(s string) (Decision, error) {
	var d Decision
	if err := json.Unmarshal([]byte(s), &d); err != nil {
		return Decision{}, fmt.Errorf("unmarshal: %w", err)
	}

	// Normaliser l'action AVANT validation
	d.Action = strings.ToLower(strings.TrimSpace(d.Action))

	if err := validateDecision(d); err != nil {
		return Decision{}, err
	}

	return d, nil
}

// validateDecision vérifie que la décision est cohérente
func validateDecision(d Decision) error {
	if strings.TrimSpace(d.Analysis) == "" {
		return fmt.Errorf("missing analysis field")
	}

	if !validActions[d.Action] {
		return fmt.Errorf("invalid action '%s' — must be: log, alert, suggest, execute", d.Action)
	}

	return nil
}

// fallbackDecision retourne une décision sûre quand tout échoue
func fallbackDecision(raw string) Decision {
	return Decision{
		Analysis: "Impossible d'analyser la réponse LLM",
		Action:   "log",
		Reason:   fmt.Sprintf("Fallback — réponse brute : %.80s", raw),
	}
}

func (d Decision) Display() {
	fmt.Printf("[ AGENT ] Action   : %s\n", d.Action)
	fmt.Printf("[ AGENT ] Analyse  : %s\n", d.Analysis)

	reason := d.Reason
	if reason == "" {
		reason = "—"
	}
	fmt.Printf("[ AGENT ] Raison   : %s\n", reason)

	if d.Command != "" {
		fmt.Printf("[ AGENT ] Commande : %s\n", d.Command)
	}
}
