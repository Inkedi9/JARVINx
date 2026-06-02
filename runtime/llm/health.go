package llm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type HealthStatus struct {
	Online  bool
	Version string
	Models  []string
	Error   string
}

func CheckOllama(baseURL string, model string) HealthStatus {
	client := &http.Client{Timeout: 5 * time.Second}

	// Ping — est-ce qu'Ollama répond ?
	resp, err := client.Get(baseURL + "/api/tags")
	if err != nil {
		return HealthStatus{
			Online: false,
			Error:  fmt.Sprintf("Ollama inaccessible sur %s : %v", baseURL, err),
		}
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return HealthStatus{
			Online: false,
			Error:  fmt.Sprintf("Ollama répond %d sur %s", resp.StatusCode, baseURL),
		}
	}

	// Parse la liste des modèles disponibles
	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return HealthStatus{
			Online: false,
			Error:  fmt.Sprintf("réponse Ollama invalide : %v", err),
		}
	}

	models := make([]string, 0, len(result.Models))
	for _, m := range result.Models {
		models = append(models, m.Name)
	}

	// Vérifie que le modèle configuré est disponible
	modelFound := false
	for _, m := range models {
		if m == model {
			modelFound = true
			break
		}
	}

	status := HealthStatus{
		Online: true,
		Models: models,
	}

	if !modelFound {
		status.Error = fmt.Sprintf(
			"modèle '%s' non trouvé — modèles disponibles : %v",
			model, models,
		)
	}

	return status
}

func (h HealthStatus) Display(model string) {
	if !h.Online {
		fmt.Printf("\033[31m[ OLLAMA ]\033[0m ✗ Hors ligne — %s\n", h.Error)
		return
	}

	if h.Error != "" {
		// Online mais modèle manquant
		fmt.Printf("\033[33m[ OLLAMA ]\033[0m ⚠ En ligne mais %s\n", h.Error)
		return
	}

	fmt.Printf("\033[32m[ OLLAMA ]\033[0m ✓ En ligne · modèle '%s' disponible · %d modèle(s) installé(s)\n",
		model, len(h.Models))
}
