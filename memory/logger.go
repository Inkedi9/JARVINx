package memory

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"time"
)

type LogEntry struct {
	Timestamp   time.Time `json:"timestamp"`
	CPUPercent  float64   `json:"cpu_percent"`
	MemUsed     uint64    `json:"mem_used_mb"`
	MemTotal    uint64    `json:"mem_total_mb"`
	MemPercent  float64   `json:"mem_percent"`
	DiskUsed    uint64    `json:"disk_used_gb"`
	DiskTotal   uint64    `json:"disk_total_gb"`
	DiskPercent float64   `json:"disk_percent"`
}

type Logger struct {
	filepath string
}

func round(v float64, decimals int) float64 {
	pow := math.Pow(10, float64(decimals))
	return math.Round(v*pow) / pow
}

func NewLogger(filepath string) *Logger {
	return &Logger{filepath: filepath}
}

func (l *Logger) Write(entry LogEntry) error {
	// On arrondit avant d'encoder
	entry.CPUPercent = round(entry.CPUPercent, 1)
	entry.MemPercent = round(entry.MemPercent, 1)
	entry.DiskPercent = round(entry.DiskPercent, 1)
	// Ouvre le fichier en mode append, le crée s'il n'existe pas
	file, err := os.OpenFile(l.filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	defer file.Close()

	// Encode l'entrée en JSON sur une seule ligne
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(entry); err != nil {
		return fmt.Errorf("encode log entry: %w", err)
	}

	return nil
}
