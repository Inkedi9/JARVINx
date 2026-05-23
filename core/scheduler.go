package core

import (
	"fmt"
	"sync"
	"time"

	"github.com/Inkedi9/jarvinx/memory"
	"github.com/Inkedi9/jarvinx/tools"
)

type Scheduler struct {
	interval time.Duration
	bus      *Bus
	mu       sync.RWMutex
}

func NewScheduler(interval time.Duration, bus *Bus) *Scheduler {
	return &Scheduler{
		interval: interval,
		bus:      bus,
	}
}

func (s *Scheduler) SetInterval(d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.interval = d
	fmt.Printf("[ SCHEDULER ] Intervalle → %v\n", d)
}

func (s *Scheduler) getInterval() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.interval
}

func (s *Scheduler) Start() {
	fmt.Printf("[ SCHEDULER ] Démarrage — tick toutes les %v\n", s.getInterval())

	for {
		time.Sleep(s.getInterval())

		state, err := tools.Observe()
		if err != nil {
			s.bus.Publish(Event{
				Type:    EventError,
				Payload: fmt.Sprintf("observe: %v", err),
			})
			continue
		}

		snap := memory.Snapshot{
			Timestamp:   state.Timestamp,
			CPUPercent:  state.CPUPercent,
			MemUsed:     state.MemUsed,
			MemTotal:    state.MemTotal,
			MemPercent:  state.MemPercent,
			DiskUsed:    state.DiskUsed,
			DiskTotal:   state.DiskTotal,
			DiskPercent: state.DiskPercent,
		}

		state.Display()

		s.bus.Publish(Event{
			Type:    EventObserved,
			Payload: snap,
		})
	}
}
