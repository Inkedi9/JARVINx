package memory

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const maxHistory = 20

type Snapshot struct {
	Timestamp   time.Time `json:"timestamp"`
	CPUPercent  float64   `json:"cpu_percent"`
	MemUsed     uint64    `json:"mem_used_mb"`
	MemTotal    uint64    `json:"mem_total_mb"`
	MemPercent  float64   `json:"mem_percent"`
	DiskUsed    uint64    `json:"disk_used_gb"`
	DiskTotal   uint64    `json:"disk_total_gb"`
	DiskPercent float64   `json:"disk_percent"`
}

type CycleRecord struct {
	Snapshot  Snapshot  `json:"snapshot"`
	Action    string    `json:"action"`
	Analysis  string    `json:"analysis"`
	Reason    string    `json:"reason"`
	Command   string    `json:"command,omitempty"`
	CycleNum  int       `json:"cycle_num"`
	Timestamp time.Time `json:"timestamp"`
}

type State struct {
	filepath string
	History  []Snapshot    `json:"history"`
	Cycles   []CycleRecord `json:"cycles"`
	CycleNum int           `json:"cycle_num"`
}

func NewState(filepath string) *State {
	s := &State{filepath: filepath}
	s.load()
	return s
}

func (s *State) Add(snap Snapshot) {
	s.History = append(s.History, snap)
	if len(s.History) > maxHistory {
		s.History = s.History[len(s.History)-maxHistory:]
	}
}

func (s *State) AddCycle(record CycleRecord) {
	s.CycleNum++
	record.CycleNum = s.CycleNum
	s.Cycles = append(s.Cycles, record)
	if len(s.Cycles) > maxHistory {
		s.Cycles = s.Cycles[len(s.Cycles)-maxHistory:]
	}
}

func (s *State) Last(n int) []Snapshot {
	if n > len(s.History) {
		n = len(s.History)
	}
	return s.History[len(s.History)-n:]
}

func (s *State) LastCycles(n int) []CycleRecord {
	if n > len(s.Cycles) {
		n = len(s.Cycles)
	}
	return s.Cycles[len(s.Cycles)-n:]
}

func (s *State) Save() error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}
	if err := os.WriteFile(s.filepath, data, 0644); err != nil {
		return fmt.Errorf("write state: %w", err)
	}
	return nil
}

func (s *State) load() {
	data, err := os.ReadFile(s.filepath)
	if err != nil {
		return
	}
	if err := json.Unmarshal(data, s); err != nil {
		fmt.Printf("[ STATE ] state.json corrompu, reset\n")
		s.History = nil
		s.Cycles = nil
	}
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
