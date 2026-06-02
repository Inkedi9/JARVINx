package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// LoadEnv charge un fichier .env dans les variables d'environnement
// Les variables déjà définies dans l'environnement ont priorité
func LoadEnv(filepath string) {
	file, err := os.Open(filepath)
	if err != nil {
		return
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		if os.Getenv(key) == "" {
			_ = os.Setenv(key, val)
		}
	}
}

// FromEnv surcharge les valeurs de Config avec les variables d'environnement
// Toute variable non définie garde sa valeur par défaut
func (c *Config) FromEnv() {
	// LLM
	if v := os.Getenv("JARVINX_MODEL"); v != "" {
		c.Model = v
	}
	if v := os.Getenv("JARVINX_OLLAMA_URL"); v != "" {
		c.OllamaURL = v
	}

	// Intervalle
	if v := os.Getenv("JARVINX_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			c.Interval = d
		} else {
			fmt.Fprintf(os.Stderr, "[ WARN ] JARVINX_INTERVAL invalide '%s' — valeur ignorée\n", v)
		}
	}

	// Seuils d'alerte
	if v := os.Getenv("JARVINX_CPU_THRESHOLD"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			c.CPUAlertThreshold = f
		} else {
			fmt.Fprintf(os.Stderr, "[ WARN ] JARVINX_CPU_THRESHOLD invalide '%s' — valeur ignorée\n", v)
		}
	}
	if v := os.Getenv("JARVINX_RAM_THRESHOLD"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			c.RAMAlertThreshold = f
		} else {
			fmt.Fprintf(os.Stderr, "[ WARN ] JARVINX_RAM_THRESHOLD invalide '%s' — valeur ignorée\n", v)
		}
	}
	if v := os.Getenv("JARVINX_DISK_THRESHOLD"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			c.DiskAlertThreshold = f
		} else {
			fmt.Fprintf(os.Stderr, "[ WARN ] JARVINX_DISK_THRESHOLD invalide '%s' — valeur ignorée\n", v)
		}
	}

	// Comportement alertes
	if v := os.Getenv("JARVINX_ALERT_COOLDOWN"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			c.AlertCooldown = i
		} else {
			fmt.Fprintf(os.Stderr, "[ WARN ] JARVINX_ALERT_COOLDOWN invalide '%s' — valeur ignorée\n", v)
		}
	}
	if v := os.Getenv("JARVINX_ALERT_MIN_CYCLES"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			c.AlertMinCycles = i
		} else {
			fmt.Fprintf(os.Stderr, "[ WARN ] JARVINX_ALERT_MIN_CYCLES invalide '%s' — valeur ignorée\n", v)
		}
	}

	// Rotation des logs
	if v := os.Getenv("JARVINX_LOG_MAX_MB"); v != "" {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			c.LogMaxSizeBytes = i * 1024 * 1024
		}
	}
	if v := os.Getenv("JARVINX_LOG_MAX_BACKUPS"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			c.LogMaxBackups = i
		}
	}

	// Dry run
	if v := os.Getenv("JARVINX_DRY_RUN"); v == "true" {
		c.DryRun = true
	}

	// Web
	if v := os.Getenv("JARVINX_PORT"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			c.WebPort = i
		} else {
			fmt.Fprintf(os.Stderr, "[ WARN ] JARVINX_PORT invalide '%s' — valeur ignorée\n", v)
		}
	}

	// Notifications
	if v := os.Getenv("DISCORD_WEBHOOK"); v != "" {
		c.DiscordWebhook = v
	}
	if v := os.Getenv("SLACK_WEBHOOK"); v != "" {
		c.SlackWebhook = v
	}
	if v := os.Getenv("NTFY_URL"); v != "" {
		c.NtfyURL = v
	}
	if v := os.Getenv("NTFY_TOPIC"); v != "" {
		c.NtfyTopic = v
	}
	if v := os.Getenv("GOTIFY_URL"); v != "" {
		c.GotifyURL = v
	}
	if v := os.Getenv("GOTIFY_TOKEN"); v != "" {
		c.GotifyToken = v
	}

	// Rapport quotidien
	if v := os.Getenv("JARVINX_DAILY_REPORT"); v == "true" {
		c.DailyReportEnabled = true
	}
	if v := os.Getenv("JARVINX_REPORT_HOUR"); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i >= 0 && i <= 23 {
			c.DailyReportHour = i
		}
	}
	if v := os.Getenv("JARVINX_REPORT_MINUTE"); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i >= 0 && i <= 59 {
			c.DailyReportMinute = i
		}
	}

	// CORS origins supplémentaires
	if v := os.Getenv("JARVINX_ALLOWED_ORIGINS"); v != "" {
		for _, o := range strings.Split(v, ",") {
			o = strings.TrimSpace(o)
			if o != "" {
				c.AllowedOrigins = append(c.AllowedOrigins, o)
			}
		}
	}

	// Docker
	if v := os.Getenv("JARVINX_DOCKER_ENABLED"); v == "false" {
		c.DockerEnabled = false
	}
	if v := os.Getenv("JARVINX_DOCKER_WATCH"); v != "" {
		c.DockerWatchList = strings.Split(v, ",")
	}

	// Files
	if v := os.Getenv("JARVINX_FILE_WATCH"); v != "" {
		c.FileWatchPaths = strings.Split(v, ",")
	}
	if v := os.Getenv("JARVINX_FILE_MAX_MB"); v != "" {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			c.FileMaxSizeMB = i
		}
	}
	if v := os.Getenv("JARVINX_FILE_ENABLED"); v == "false" {
		c.FileEnabled = false
	}

	// SQLite history store
	if v := os.Getenv("JARVINX_SQLITE_PATH"); v != "" {
		c.SQLitePath = v
	}

	// Execute guard
	if v := os.Getenv("JARVINX_EXEC_COOLDOWN"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			c.ExecCooldown = d
		} else {
			fmt.Fprintf(os.Stderr, "[ WARN ] JARVINX_EXEC_COOLDOWN invalide '%s' — valeur ignorée\n", v)
		}
	}
}
