package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/shirou/gopsutil/v3/disk"
)

// FileInfo représente un fichier ou dossier analysé
type FileInfo struct {
	Path      string
	SizeBytes int64
	SizeMB    float64
	IsDir     bool
	Children  int       // nombre de fichiers dans le dossier (si IsDir)
	ModTime   time.Time // last modification time
}

// DirStats represents the global stats of a monitored directory
type DirStats struct {
	Path          string
	TotalBytes    int64
	TotalMB       float64
	FileCount     int
	LargeFiles    []FileInfo // fichiers au-dessus du seuil
	LastModified  time.Time  // modtime du fichier le plus récemment modifié
	InodesUsed    uint64     // 0 on Windows (fail-silent)
	InodesPercent float64    // 0 on Windows (fail-silent)
	Error         string
}

// ScanDirectory analyse un dossier et retourne ses stats
func ScanDirectory(path string, maxSizeMB int64) DirStats {
	stats := DirStats{Path: path}

	info, err := os.Stat(path)
	if err != nil {
		stats.Error = fmt.Sprintf("inaccessible : %v", err)
		return stats
	}

	if !info.IsDir() {
		stats.Error = fmt.Sprintf("'%s' n'est pas un dossier", path)
		return stats
	}

	maxBytes := maxSizeMB * 1024 * 1024

	err = filepath.Walk(path, func(p string, fi os.FileInfo, err error) error {
		if err != nil {
			return nil // ignore les erreurs de permission
		}
		if fi.IsDir() {
			return nil
		}

		stats.TotalBytes += fi.Size()
		stats.FileCount++

		if fi.ModTime().After(stats.LastModified) {
			stats.LastModified = fi.ModTime()
		}

		if fi.Size() >= maxBytes {
			stats.LargeFiles = append(stats.LargeFiles, FileInfo{
				Path:      p,
				SizeBytes: fi.Size(),
				SizeMB:    float64(fi.Size()) / 1024 / 1024,
				ModTime:   fi.ModTime(),
			})
		}

		return nil
	})

	if err != nil {
		stats.Error = err.Error()
	}

	stats.TotalMB = float64(stats.TotalBytes) / 1024 / 1024

	// Inodes — fail-silent (returns 0 on Windows)
	if usage, inoErr := disk.Usage(path); inoErr == nil {
		stats.InodesUsed = usage.InodesUsed
		stats.InodesPercent = usage.InodesUsedPercent
	}

	// Trie les gros fichiers par taille décroissante
	sort.Slice(stats.LargeFiles, func(i, j int) bool {
		return stats.LargeFiles[i].SizeBytes > stats.LargeFiles[j].SizeBytes
	})

	// Garde max 10 gros fichiers
	if len(stats.LargeFiles) > 10 {
		stats.LargeFiles = stats.LargeFiles[:10]
	}

	return stats
}
