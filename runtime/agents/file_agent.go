package agents

import (
	"context"
	"fmt"
	"time"

	jxlog "github.com/Inkedi9/jarvinx/jxlog"
	"github.com/Inkedi9/jarvinx/tools"
)

const unknownAge = "unknown"

type FileAgent struct {
	BaseAgent
	watchPaths []string
	maxSizeMB  int64
	dryRun     bool
	prevSizes  map[string]float64   // taille précédente par path en MB
	prevTimes  map[string]time.Time // timestamp du dernier scan par path
}

func NewFileAgent(watchPaths []string, maxSizeMB int64, dryRun bool) *FileAgent {
	return &FileAgent{
		BaseAgent:  NewBaseAgent("file", 5*time.Minute),
		watchPaths: watchPaths,
		maxSizeMB:  maxSizeMB,
		dryRun:     dryRun,
		prevSizes:  make(map[string]float64),
		prevTimes:  make(map[string]time.Time),
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

		lastMod := unknownAge
		if !stats.LastModified.IsZero() {
			lastMod = time.Since(stats.LastModified).Round(time.Minute).String() + " ago"
		}
		inodesInfo := ""
		if stats.InodesUsed > 0 {
			inodesInfo = fmt.Sprintf(" | inodes %.1f%%", stats.InodesPercent)
		}
		jxlog.Debug("FILE AGENT", fmt.Sprintf(
			"'%s' → %.1f MB | %d fichiers | %d gros fichiers | last-mod: %s%s",
			path, stats.TotalMB, stats.FileCount, len(stats.LargeFiles), lastMod, inodesInfo,
		))

		// Détecte les gros fichiers
		for _, f := range stats.LargeFiles {
			msg := fmt.Sprintf(
				"Fichier volumineux détecté : %s (%.1f MB, modifié %s)",
				f.Path, f.SizeMB, f.ModTime.Format("2006-01-02 15:04"),
			)
			if a.dryRun {
				jxlog.Info("DRY-RUN", fmt.Sprintf("File alert simulée : %s", msg))
			} else {
				jxlog.Warn("FILE AGENT", msg)
			}
			hasAlerts = true
		}

		// Détecte la croissance rapide du dossier (avec taux MB/min)
		if prev, ok := a.prevSizes[path]; ok {
			growth := stats.TotalMB - prev
			if growth > float64(a.maxSizeMB)/2 {
				rate := 0.0
				if prevT, hasPrevT := a.prevTimes[path]; hasPrevT {
					if elapsed := time.Since(prevT).Minutes(); elapsed > 0 {
						rate = growth / elapsed
					}
				}
				msg := fmt.Sprintf(
					"Dossier '%s' a grandi de %.1f MB (%.2f MB/min) — total : %.1f MB",
					path, growth, rate, stats.TotalMB,
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
		a.prevTimes[path] = time.Now()
	}

	if hasAlerts {
		a.recordAlert()
	} else {
		a.recordSuccess()
	}

	return nil
}
