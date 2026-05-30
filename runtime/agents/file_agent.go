package agents

import (
	"context"
	"fmt"
	"time"

	jxlog "github.com/Inkedi9/jarvinx/jxlog"
	"github.com/Inkedi9/jarvinx/tools"
)

type FileAgent struct {
	BaseAgent
	watchPaths []string
	maxSizeMB  int64
	dryRun     bool
	prevSizes  map[string]float64 // taille précédente par path en MB
}

func NewFileAgent(watchPaths []string, maxSizeMB int64, dryRun bool) *FileAgent {
	return &FileAgent{
		BaseAgent:  NewBaseAgent("file", 5*time.Minute),
		watchPaths: watchPaths,
		maxSizeMB:  maxSizeMB,
		dryRun:     dryRun,
		prevSizes:  make(map[string]float64),
	}
}

func (a *FileAgent) Run(ctx context.Context, actx AgentContext) error {
	if len(a.watchPaths) == 0 {
		jxlog.Debug("FILE AGENT", "Aucun dossier configuré — cycle ignoré")
		a.recordSuccess()
		return nil
	}

	hasAlerts := false

	for _, path := range a.watchPaths {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		stats := tools.ScanDirectory(path, a.maxSizeMB)

		if stats.Error != "" {
			jxlog.Warn("FILE AGENT", fmt.Sprintf("Scan '%s' : %s", path, stats.Error))
			continue
		}

		jxlog.Debug("FILE AGENT", fmt.Sprintf(
			"'%s' → %.1f MB | %d fichiers | %d gros fichiers",
			path, stats.TotalMB, stats.FileCount, len(stats.LargeFiles),
		))

		// Détecte les gros fichiers
		for _, f := range stats.LargeFiles {
			msg := fmt.Sprintf(
				"Fichier volumineux détecté : %s (%.1f MB)",
				f.Path, f.SizeMB,
			)
			if a.dryRun {
				jxlog.Info("DRY-RUN", fmt.Sprintf("File alert simulée : %s", msg))
			} else {
				jxlog.Warn("FILE AGENT", msg)
			}
			hasAlerts = true
		}

		// Détecte la croissance rapide du dossier
		if prev, ok := a.prevSizes[path]; ok {
			growth := stats.TotalMB - prev
			if growth > float64(a.maxSizeMB)/2 {
				msg := fmt.Sprintf(
					"Dossier '%s' a grandi de %.1f MB en un cycle (total : %.1f MB)",
					path, growth, stats.TotalMB,
				)
				if a.dryRun {
					jxlog.Info("DRY-RUN", fmt.Sprintf("File growth alert simulée : %s", msg))
				} else {
					jxlog.Warn("FILE AGENT", msg)
				}
				hasAlerts = true
			}
		}

		a.prevSizes[path] = stats.TotalMB
	}

	if hasAlerts {
		a.recordAlert()
	} else {
		a.recordSuccess()
	}

	return nil
}
