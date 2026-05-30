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
	filepath   string
	maxBytes   int64
	maxBackups int
	mu         sync.Mutex
}

func NewLogger(filepath string) *Logger {
	return NewLoggerWithRotation(filepath, 10*1024*1024, 3)
}

func NewLoggerWithRotation(filepath string, maxBytes int64, maxBackups int) *Logger {
	return &Logger{
		filepath:   filepath,
		maxBytes:   maxBytes,
		maxBackups: maxBackups,
	}
}

func (l *Logger) Write(entry LogEntry) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Vérifie si rotation nécessaire avant d'écrire
	if l.maxBytes > 0 {
		if err := l.rotateIfNeeded(); err != nil {
			return fmt.Errorf("rotate: %w", err)
		}
	}

	// 0644 → 0600
	file, err := os.OpenFile(l.filepath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(entry); err != nil {
		return fmt.Errorf("encode log entry: %w", err)
	}

	return nil
}

// rotateIfNeeded vérifie la taille et rotate si nécessaire
// Doit être appelé avec le mutex déjà acquis
func (l *Logger) rotateIfNeeded() error {
	info, err := os.Stat(l.filepath)
	if os.IsNotExist(err) {
		return nil // fichier pas encore créé
	}
	if err != nil {
		return fmt.Errorf("stat: %w", err)
	}

	if info.Size() < l.maxBytes {
		return nil // pas encore à la limite
	}

	return l.rotate()
}

// rotate déplace les fichiers existants et archive le courant
func (l *Logger) rotate() error {
	// Supprime le backup le plus vieux si on atteint la limite
	oldest := fmt.Sprintf("%s.%d", l.filepath, l.maxBackups)
	_ = os.Remove(oldest) // ignore l'erreur si n'existe pas

	// Décale les backups existants : .2 → .3, .1 → .2
	for i := l.maxBackups - 1; i >= 1; i-- {
		src := fmt.Sprintf("%s.%d", l.filepath, i)
		dst := fmt.Sprintf("%s.%d", l.filepath, i+1)
		_ = os.Rename(src, dst) // on ignore intentionnellement — fichier peut ne pas exister
	}

	// Archive le fichier courant : logs.jsonl → logs.jsonl.1
	if err := os.Rename(l.filepath, l.filepath+".1"); err != nil {
		return fmt.Errorf("archive current log: %w", err)
	}

	return nil
}

// Size retourne la taille actuelle du fichier de log en bytes
func (l *Logger) Size() int64 {
	l.mu.Lock()
	defer l.mu.Unlock()

	info, err := os.Stat(l.filepath)
	if err != nil {
		return 0
	}
	return info.Size()
}

func (l *Logger) Filepath() string {
	return l.filepath
}
