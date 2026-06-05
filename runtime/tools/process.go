package tools

import (
	"context"
	"sort"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

// ProcInfo mirrors memory.ProcInfo — kept in tools to avoid a circular import.
type ProcInfo struct {
	PID   int32
	Name  string
	MemMB uint64
}

// TopProcesses returns up to n processes sorted by resident memory (RSS).
// Enforces a 2s sub-timeout; fail-silent on any error.
func TopProcesses(ctx context.Context, n int) []ProcInfo {
	tCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	procs, err := process.ProcessesWithContext(tCtx)
	if err != nil {
		return nil
	}

	infos := make([]ProcInfo, 0, len(procs))
	for _, p := range procs {
		if tCtx.Err() != nil {
			break
		}
		name, err := p.NameWithContext(tCtx)
		if err != nil {
			continue
		}
		memInfo, err := p.MemoryInfoWithContext(tCtx)
		if err != nil || memInfo == nil {
			continue
		}
		infos = append(infos, ProcInfo{
			PID:   p.Pid,
			Name:  name,
			MemMB: memInfo.RSS / 1024 / 1024,
		})
	}

	sort.Slice(infos, func(i, j int) bool {
		return infos[i].MemMB > infos[j].MemMB
	})
	if len(infos) > n {
		infos = infos[:n]
	}
	return infos
}
