package core

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Inkedi9/jarvinx/jxlog"
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
	jxlog.Info("SCHEDULER", fmt.Sprintf("Intervalle → %v", d))
}

func (s *Scheduler) getInterval() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.interval
}

func (s *Scheduler) Start(ctx context.Context) {
	jxlog.Info("SCHEDULER", fmt.Sprintf("Starting — tick toutes les %v", s.getInterval()))

	ticker := time.NewTicker(s.getInterval())
	defer ticker.Stop()

	currentInterval := s.getInterval()

	for {
		select {
		case <-ctx.Done():
			jxlog.Info("SCHEDULER", "Arrêt propre")
			return

		case <-ticker.C:
			newInterval := s.getInterval()
			if newInterval != currentInterval {
				ticker.Stop()
				ticker = time.NewTicker(newInterval)
				currentInterval = newInterval
				jxlog.Info("SCHEDULER", fmt.Sprintf("Ticker mis à jour → %v", newInterval))
			}

			// Observe interruptible — répond au ctx.Done()
			state, err := tools.ObserveWithContext(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return // context annulé — sortie propre
				}
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
}

// Restart recrée le ticker avec le nouvel intervalle
// Appelé depuis CLI quand l'utilisateur change l'intervalle à chaud
func (s *Scheduler) Restart(ctx context.Context) {
	go s.Start(ctx)
}
