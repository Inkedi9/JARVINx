package memory

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
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
	mu       sync.Mutex
}

func NewLogger(filepath string) *Logger {
	return &Logger{filepath: filepath}
}

func (l *Logger) Write(entry LogEntry) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 0644 → 0600
	file, err := os.OpenFile(l.filepath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(entry); err != nil {
		return fmt.Errorf("encode log entry: %w", err)
	}

	return nil
}
