package tools

import (
	"fmt"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

type SystemState struct {
	Timestamp   time.Time
	CPUPercent  float64
	MemTotal    uint64
	MemUsed     uint64
	MemPercent  float64
	DiskTotal   uint64
	DiskUsed    uint64
	DiskPercent float64
}

func rootPath() string {
	if runtime.GOOS == "windows" {
		return "C:\\"
	}
	return "/"
}

func Observe() (SystemState, error) {
	// CPU — on attend 1 seconde pour avoir une mesure réelle
	cpuPercent, err := cpu.Percent(1*time.Second, false)
	if err != nil {
		return SystemState{}, fmt.Errorf("cpu: %w", err)
	}

	// RAM
	memStats, err := mem.VirtualMemory()
	if err != nil {
		return SystemState{}, fmt.Errorf("mem: %w", err)
	}

	// Disque (racine du système)
	diskStats, err := disk.Usage(rootPath())
	if err != nil {
		return SystemState{}, fmt.Errorf("disk: %w", err)
	}

	return SystemState{
		Timestamp:   time.Now(),
		CPUPercent:  cpuPercent[0],
		MemTotal:    memStats.Total / 1024 / 1024,
		MemUsed:     memStats.Used / 1024 / 1024,
		MemPercent:  memStats.UsedPercent,
		DiskTotal:   diskStats.Total / 1024 / 1024 / 1024,
		DiskUsed:    diskStats.Used / 1024 / 1024 / 1024,
		DiskPercent: diskStats.UsedPercent,
	}, nil
}

func (s SystemState) Display() {
	diskLabel := rootPath()

	cpuColor := metricColor(s.CPUPercent, 70, 85)
	ramColor := metricColor(s.MemPercent, 70, 90)
	diskColor := metricColor(s.DiskPercent, 70, 85)

	fmt.Printf("\n\033[97m┌─[ JARVINX OBSERVE ]──────────────────────────┐\033[0m\n")
	fmt.Printf("\033[97m│\033[0m \033[90m%s\033[0m\n", s.Timestamp.Format("15:04:05"))
	fmt.Printf("\033[97m│\033[0m CPU   : %s%.1f%%%s\n", cpuColor, s.CPUPercent, "\033[0m")
	fmt.Printf("\033[97m│\033[0m RAM   : %s%d MB\033[0m / %d MB (%.1f%%)\n",
		ramColor, s.MemUsed, s.MemTotal, s.MemPercent)
	fmt.Printf("\033[97m│\033[0m DISK  : %s%d GB\033[0m / %d GB (%.1f%%) [%s]\n",
		diskColor, s.DiskUsed, s.DiskTotal, s.DiskPercent, diskLabel)
	fmt.Printf("\033[97m└──────────────────────────────────────────────┘\033[0m\n")
}

func metricColor(pct, warn, crit float64) string {
	if pct >= crit {
		return "\033[31m"
	}
	if pct >= warn {
		return "\033[33m"
	}
	return "\033[32m"
}
