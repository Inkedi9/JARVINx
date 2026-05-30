package config

import "time"

const (
	minInterval = 5 * time.Second
	maxInterval = 1 * time.Hour
)

type Config struct {
	// Runtime
	Interval time.Duration
	DryRun   bool

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

	// Alerts
	AlertFile      string
	AlertCooldown  int
	AlertMinCycles int

	// Notifications
	DiscordWebhook string
	SlackWebhook   string
	NtfyURL        string
	NtfyTopic      string
	GotifyURL      string
	GotifyToken    string

	AllowedOrigins []string

	// Rotation des logs
	LogMaxSizeBytes   int64
	LogMaxBackups     int
	AlertMaxSizeBytes int64
	AlertMaxBackups   int

	// Docker
	DockerEnabled   bool
	DockerWatchList []string

	// Files
	FileEnabled    bool
	FileWatchPaths []string
	FileMaxSizeMB  int64
}

func Default() *Config {
	return &Config{
		Interval:           15 * time.Second,
		OllamaURL:          "http://localhost:11434",
		Model:              "llama3.1:8b",
		LogFile:            "logs.jsonl",
		StateFile:          "state.json",
		AlertFile:          "alerts.jsonl",
		WebPort:            8080,
		CPUAlertThreshold:  85.0,
		RAMAlertThreshold:  90.0,
		DiskAlertThreshold: 85.0,
		AlertCooldown:      5,
		AlertMinCycles:     2,
		DiscordWebhook:     "",
		AllowedOrigins: []string{
			"http://localhost:3000", // Next.js dev
			"http://localhost:8080", // dashboard servi par Go
		},
		LogMaxSizeBytes:   10 * 1024 * 1024, // 10 MB
		LogMaxBackups:     3,
		AlertMaxSizeBytes: 5 * 1024 * 1024, // 5 MB
		AlertMaxBackups:   3,
		DryRun:            false,
		DockerEnabled:     true,
		DockerWatchList:   []string{},
		FileEnabled:       true,
		FileWatchPaths:    []string{}, // vide = désactivé jusqu'à config
		FileMaxSizeMB:     500,        // alerte si fichier > 500 MB
		NtfyURL:           "https://ntfy.sh",
		NtfyTopic:         "jarvinx",
	}
}
