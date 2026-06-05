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

	// Baseline réseau — premier tick aura un delta valide
	prevNet, hasNet := tools.ReadNetCounters()
	prevNetAt := time.Now()

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

			// Delta réseau depuis le tick précédent
			currNet, ok := tools.ReadNetCounters()
			if ok && hasNet {
				elapsed := time.Since(prevNetAt).Seconds()
				state.NetRecvMBps, state.NetSentMBps = tools.DeltaMBps(prevNet, currNet, elapsed)
			}
			if ok {
				prevNet = currNet
				hasNet = true
			}
			prevNetAt = time.Now()

			snap := memory.Snapshot{
				Timestamp:   state.Timestamp,
				CPUPercent:  state.CPUPercent,
				MemUsed:     state.MemUsed,
				MemTotal:    state.MemTotal,
				MemPercent:  state.MemPercent,
				SwapUsed:    state.SwapUsed,
				SwapTotal:   state.SwapTotal,
				SwapPercent: state.SwapPercent,
				DiskUsed:    state.DiskUsed,
				DiskTotal:   state.DiskTotal,
				DiskPercent: state.DiskPercent,
				NetRecvMBps: state.NetRecvMBps,
				NetSentMBps: state.NetSentMBps,
				LoadAvg1:    state.LoadAvg1,
				LoadAvg5:    state.LoadAvg5,
				LoadAvg15:   state.LoadAvg15,
				TopProcs:    toMemProcInfos(state.TopProcs),
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

func toMemProcInfos(procs []tools.ProcInfo) []memory.ProcInfo {
	if len(procs) == 0 {
		return nil
	}
	result := make([]memory.ProcInfo, len(procs))
	for i, p := range procs {
		result[i] = memory.ProcInfo{PID: p.PID, Name: p.Name, MemMB: p.MemMB}
	}
	return result
}
