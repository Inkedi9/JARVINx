package config

import (
	"bufio"
	"os"
	"strings"
)

func LoadEnv(filepath string) {
	file, err := os.Open(filepath)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Ignore commentaires et lignes vides
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		// Ne pas écraser une variable déjà définie dans l'environnement
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
}
