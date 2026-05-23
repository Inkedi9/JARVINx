package tools

import (
	"fmt"
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
	diskPath := "C:\\"
	diskStats, err := disk.Usage(diskPath)
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
	fmt.Printf("\n┌─[ JARVINX OBSERVE ]──────────────────────────┐\n")
	fmt.Printf("│ %s\n", s.Timestamp.Format("15:04:05"))
	fmt.Printf("│ CPU   : %.1f%%\n", s.CPUPercent)
	fmt.Printf("│ RAM   : %d MB / %d MB (%.1f%%)\n", s.MemUsed, s.MemTotal, s.MemPercent)
	fmt.Printf("│ DISK  : %d GB / %d GB (%.1f%%)\n", s.DiskUsed, s.DiskTotal, s.DiskPercent)
	fmt.Printf("└──────────────────────────────────────────────┘\n")
}
