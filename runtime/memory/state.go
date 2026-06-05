package memory

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

const (
	maxHistory          = 20
	currentStateVersion = 1
)

type Snapshot struct {
	Timestamp   time.Time `json:"timestamp"`
	CPUPercent  float64   `json:"cpu_percent"`
	MemUsed     uint64    `json:"mem_used_mb"`
	MemTotal    uint64    `json:"mem_total_mb"`
	MemPercent  float64   `json:"mem_percent"`
	SwapUsed    uint64    `json:"swap_used_mb,omitempty"`
	SwapTotal   uint64    `json:"swap_total_mb,omitempty"`
	SwapPercent float64   `json:"swap_percent,omitempty"`
	DiskUsed    uint64    `json:"disk_used_gb"`
	DiskTotal   uint64    `json:"disk_total_gb"`
	DiskPercent float64   `json:"disk_percent"`
	NetRecvMBps float64    `json:"net_recv_mbps,omitempty"`
	NetSentMBps float64    `json:"net_sent_mbps,omitempty"`
	LoadAvg1    float64    `json:"load_avg1,omitempty"`
	LoadAvg5    float64    `json:"load_avg5,omitempty"`
	LoadAvg15   float64    `json:"load_avg15,omitempty"`
	TopProcs    []ProcInfo `json:"top_procs,omitempty"`
}

type CycleRecord struct {
	Snapshot  Snapshot  `json:"snapshot"`
	Action    string    `json:"action"`
	Analysis  string    `json:"analysis"`
	Reason    string    `json:"reason"`
	Command   string    `json:"command,omitempty"`
	CycleNum  int       `json:"cycle_num"`
	Timestamp time.Time `json:"timestamp"`

	// Métriques au moment de la décision "execute" — pour le verify N+1
	TriggerCPU  float64 `json:"trigger_cpu,omitempty"`
	TriggerRAM  float64 `json:"trigger_ram,omitempty"`
	TriggerDisk float64 `json:"trigger_disk,omitempty"`

	Confidence float64 `json:"confidence,omitempty"`
}

type State struct {
	filepath string
	mu       sync.RWMutex

	Version  int           `json:"version"`
	History  []Snapshot    `json:"history"`
	Cycles   []CycleRecord `json:"cycles"`
	CycleNum int           `json:"cycle_num"`
}

func NewState(filepath string) *State {
	s := &State{filepath: filepath}
	s.load()
	return s
}

func (s *State) Add(snap Snapshot) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.History = append(s.History, snap)
	if len(s.History) > maxHistory {
		s.History = s.History[len(s.History)-maxHistory:]
	}
	return nil
}

func (s *State) AddCycle(record CycleRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.CycleNum++
	record.CycleNum = s.CycleNum
	s.Cycles = append(s.Cycles, record)
	if len(s.Cycles) > maxHistory {
		s.Cycles = s.Cycles[len(s.Cycles)-maxHistory:]
	}
	return nil
}

func (s *State) CurrentCycle() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.CycleNum
}

func (s *State) Last(n int) []Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if n > len(s.History) {
		n = len(s.History)
	}
	// Copie défensive — on retourne jamais une slice du slice interne
	result := make([]Snapshot, n)
	copy(result, s.History[len(s.History)-n:])
	return result
}

func (s *State) LastCycles(n int) []CycleRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if n > len(s.Cycles) {
		n = len(s.Cycles)
	}
	result := make([]CycleRecord, n)
	copy(result, s.Cycles[len(s.Cycles)-n:])
	return result
}

func (s *State) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}
	// 0644 → 0600
	if err := os.WriteFile(s.filepath, data, 0600); err != nil {
		return fmt.Errorf("write state: %w", err)
	}
	return nil
}

func (s *State) load() {
	data, err := os.ReadFile(s.filepath)
	if err != nil {
		s.Version = currentStateVersion
		return
	}
	if err := json.Unmarshal(data, s); err != nil {
		fmt.Printf("[ STATE ] state.json corrompu, reset\n")
		s.History = nil
		s.Cycles = nil
		s.Version = currentStateVersion
		return
	}
	s.migrateFrom(s.Version)
}

// migrateFrom upgrades the loaded state to currentStateVersion.
// Each version step applies its data changes before moving to the next.
func (s *State) migrateFrom(from int) {
	if from >= currentStateVersion {
		return
	}
	// v0 → v1: version field added, no structural change
	s.Version = currentStateVersion
}

func NewCycleRecord(snap Snapshot, action, analysis, reason, command string) CycleRecord {
	return CycleRecord{
		Snapshot:  snap,
		Action:    action,
		Analysis:  analysis,
		Reason:    reason,
		Command:   command,
		Timestamp: snap.Timestamp,
	}
}
