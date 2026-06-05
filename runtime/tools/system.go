package tools

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
)

const topN = 5

type SystemState struct {
	Timestamp   time.Time
	CPUPercent  float64
	MemTotal    uint64
	MemUsed     uint64
	MemPercent  float64
	SwapTotal   uint64
	SwapUsed    uint64
	SwapPercent float64
	DiskTotal   uint64
	DiskUsed    uint64
	DiskPercent float64
	NetRecvMBps float64
	NetSentMBps float64
	LoadAvg1    float64
	LoadAvg5    float64
	LoadAvg15   float64
	TopProcs    []ProcInfo
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

	// Swap — fail-silent (absent sur certains OS/configs)
	var swapTotal, swapUsed uint64
	var swapPercent float64
	if swapStats, swapErr := mem.SwapMemory(); swapErr == nil {
		swapTotal = swapStats.Total / 1024 / 1024
		swapUsed = swapStats.Used / 1024 / 1024
		swapPercent = swapStats.UsedPercent
	}

	// Load average — fail-silent sur Windows
	load1, load5, load15 := readLoadAvg()

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
		SwapTotal:   swapTotal,
		SwapUsed:    swapUsed,
		SwapPercent: swapPercent,
		DiskTotal:   diskStats.Total / 1024 / 1024 / 1024,
		DiskUsed:    diskStats.Used / 1024 / 1024 / 1024,
		DiskPercent: diskStats.UsedPercent,
		LoadAvg1:    load1,
		LoadAvg5:    load5,
		LoadAvg15:   load15,
		TopProcs:    TopProcesses(context.Background(), topN),
	}, nil
}

func ObserveWithContext(ctx context.Context) (SystemState, error) {
	// cpu.Percent avec context — interruptible
	done := make(chan struct {
		pct []float64
		err error
	}, 1)

	go func() {
		pct, err := cpu.Percent(1*time.Second, false)
		done <- struct {
			pct []float64
			err error
		}{pct, err}
	}()

	select {
	case <-ctx.Done():
		return SystemState{}, ctx.Err()
	case result := <-done:
		if result.err != nil {
			return SystemState{}, fmt.Errorf("cpu: %w", result.err)
		}

		memStats, err := mem.VirtualMemory()
		if err != nil {
			return SystemState{}, fmt.Errorf("mem: %w", err)
		}

		var swapTotal, swapUsed uint64
		var swapPercent float64
		if swapStats, swapErr := mem.SwapMemory(); swapErr == nil {
			swapTotal = swapStats.Total / 1024 / 1024
			swapUsed = swapStats.Used / 1024 / 1024
			swapPercent = swapStats.UsedPercent
		}

		load1, load5, load15 := readLoadAvg()

		diskStats, err := disk.Usage(rootPath())
		if err != nil {
			return SystemState{}, fmt.Errorf("disk: %w", err)
		}

		return SystemState{
			Timestamp:   time.Now(),
			CPUPercent:  result.pct[0],
			MemTotal:    memStats.Total / 1024 / 1024,
			MemUsed:     memStats.Used / 1024 / 1024,
			MemPercent:  memStats.UsedPercent,
			SwapTotal:   swapTotal,
			SwapUsed:    swapUsed,
			SwapPercent: swapPercent,
			DiskTotal:   diskStats.Total / 1024 / 1024 / 1024,
			DiskUsed:    diskStats.Used / 1024 / 1024 / 1024,
			DiskPercent: diskStats.UsedPercent,
			LoadAvg1:    load1,
			LoadAvg5:    load5,
			LoadAvg15:   load15,
			TopProcs:    TopProcesses(ctx, topN),
		}, nil
	}
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
	if s.SwapTotal > 0 {
		swapColor := metricColor(s.SwapPercent, 60, 80)
		fmt.Printf("\033[97m│\033[0m SWAP  : %s%d MB\033[0m / %d MB (%.1f%%)\n",
			swapColor, s.SwapUsed, s.SwapTotal, s.SwapPercent)
	}
	fmt.Printf("\033[97m│\033[0m DISK  : %s%d GB\033[0m / %d GB (%.1f%%) [%s]\n",
		diskColor, s.DiskUsed, s.DiskTotal, s.DiskPercent, diskLabel)
	if s.NetRecvMBps > 0 || s.NetSentMBps > 0 {
		fmt.Printf("\033[97m│\033[0m NET   : \033[36m↓%.2f MB/s  ↑%.2f MB/s\033[0m\n",
			s.NetRecvMBps, s.NetSentMBps)
	}
	if s.LoadAvg1 > 0 {
		fmt.Printf("\033[97m│\033[0m LOAD  : \033[90m%.2f  %.2f  %.2f\033[0m\n",
			s.LoadAvg1, s.LoadAvg5, s.LoadAvg15)
	}
	if len(s.TopProcs) > 0 {
		fmt.Printf("\033[97m│\033[0m PROCS :")
		for _, p := range s.TopProcs {
			fmt.Printf(" \033[90m%s(%dMB)\033[0m", p.Name, p.MemMB)
		}
		fmt.Printf("\n")
	}
	fmt.Printf("\033[97m└──────────────────────────────────────────────┘\033[0m\n")
}

// readLoadAvg returns load averages fail-silent (always 0 on Windows).
func readLoadAvg() (avg1, avg5, avg15 float64) {
	if s, err := load.Avg(); err == nil {
		return s.Load1, s.Load5, s.Load15
	}
	return 0, 0, 0
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
