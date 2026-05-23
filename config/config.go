package config

import "time"

type Config struct {
	// Runtime
	Interval time.Duration

	// LLM
	OllamaURL string
	Model     string

	// Memory
	LogFile   string
	StateFile string

	// Seuils d'alerte
	CPUAlertThreshold  float64
	RAMAlertThreshold  float64
	DiskAlertThreshold float64

	// Ports
	WebPort int
}

func Default() *Config {
	return &Config{
		Interval:           15 * time.Second,
		OllamaURL:          "http://localhost:11434",
		Model:              "llama3.1:8b",
		LogFile:            "logs.jsonl",
		StateFile:          "state.json",
		WebPort:            8080,
		CPUAlertThreshold:  85.0,
		RAMAlertThreshold:  90.0,
		DiskAlertThreshold: 90.0,
	}
}
