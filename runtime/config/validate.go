package config

import (
	"errors"
	"fmt"
	"strings"
)

// ValidationError agrège toutes les erreurs de config en une seule
type ValidationError struct {
	Errors []string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("configuration invalide :\n  - %s",
		strings.Join(e.Errors, "\n  - "))
}

func (e *ValidationError) Add(msg string) {
	e.Errors = append(e.Errors, msg)
}

func (e *ValidationError) HasErrors() bool {
	return len(e.Errors) > 0
}

// Validate vérifie toute la configuration et retourne toutes les erreurs d'un coup
func (c *Config) Validate() error {
	ve := &ValidationError{}

	// ── Intervalle ────────────────────────────────────────────────
	if c.Interval < minInterval {
		ve.Add(fmt.Sprintf("Interval trop court : %v (minimum %v)", c.Interval, minInterval))
	}
	if c.Interval > maxInterval {
		ve.Add(fmt.Sprintf("Interval trop long : %v (maximum %v)", c.Interval, maxInterval))
	}

	// ── Seuils d'alerte ───────────────────────────────────────────
	if err := validateThreshold("CPUAlertThreshold", c.CPUAlertThreshold); err != nil {
		ve.Add(err.Error())
	}
	if err := validateThreshold("RAMAlertThreshold", c.RAMAlertThreshold); err != nil {
		ve.Add(err.Error())
	}
	if err := validateThreshold("DiskAlertThreshold", c.DiskAlertThreshold); err != nil {
		ve.Add(err.Error())
	}

	// ── Comportement alertes ──────────────────────────────────────
	if c.AlertMinCycles < 1 {
		ve.Add(fmt.Sprintf("AlertMinCycles doit être >= 1, got %d", c.AlertMinCycles))
	}
	if c.AlertMinCycles > 20 {
		ve.Add(fmt.Sprintf("AlertMinCycles doit être <= 20, got %d", c.AlertMinCycles))
	}
	if c.AlertCooldown < 1 {
		ve.Add(fmt.Sprintf("AlertCooldown doit être >= 1, got %d", c.AlertCooldown))
	}

	// ── Réseau ────────────────────────────────────────────────────
	if c.WebPort < 1024 || c.WebPort > 65535 {
		ve.Add(fmt.Sprintf("WebPort invalide : %d (doit être entre 1024 et 65535)", c.WebPort))
	}

	if len(c.AllowedOrigins) == 0 {
		ve.Add("AllowedOrigins ne peut pas être vide")
	}

	for _, origin := range c.AllowedOrigins {
		if !strings.HasPrefix(origin, "http://") &&
			!strings.HasPrefix(origin, "https://") {
			ve.Add(fmt.Sprintf("AllowedOrigins: '%s' doit commencer par http:// ou https://", origin))
		}
	}

	// ── Fichiers ──────────────────────────────────────────────────
	if strings.TrimSpace(c.LogFile) == "" {
		ve.Add("LogFile ne peut pas être vide")
	}
	if strings.TrimSpace(c.StateFile) == "" {
		ve.Add("StateFile ne peut pas être vide")
	}
	if strings.TrimSpace(c.AlertFile) == "" {
		ve.Add("AlertFile ne peut pas être vide")
	}

	// ── LLM ──────────────────────────────────────────────────────
	if strings.TrimSpace(c.OllamaURL) == "" {
		ve.Add("OllamaURL ne peut pas être vide")
	}
	if strings.TrimSpace(c.Model) == "" {
		ve.Add("Model ne peut pas être vide")
	}

	if ve.HasErrors() {
		return ve
	}
	return nil
}

func validateThreshold(name string, value float64) error {
	if value <= 0 {
		return fmt.Errorf("%s doit être > 0, got %.1f", name, value)
	}
	if value > 100 {
		return fmt.Errorf("%s doit être <= 100, got %.1f", name, value)
	}
	return nil
}

// Erreurs sentinelles — utilisables pour des type assertions
var (
	ErrConfigInvalid = errors.New("configuration invalide")
)
