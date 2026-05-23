package memory

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const maxHistory = 10

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

type State struct {
	filepath string
	History  []Snapshot `json:"history"`
}

func NewState(filepath string) *State {
	s := &State{filepath: filepath}
	s.load()
	return s
}

func (s *State) Add(snap Snapshot) {
	s.History = append(s.History, snap)

	// On garde seulement les N derniers snapshots
	if len(s.History) > maxHistory {
		s.History = s.History[len(s.History)-maxHistory:]
	}
}

func (s *State) Last(n int) []Snapshot {
	if n > len(s.History) {
		n = len(s.History)
	}
	return s.History[len(s.History)-n:]
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
		// Fichier inexistant au premier démarrage — c'est normal
		return
	}

	if err := json.Unmarshal(data, s); err != nil {
		fmt.Printf("[ STATE ] Attention : state.json corrompu, reset\n")
		s.History = nil
	}
}
