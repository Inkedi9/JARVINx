package core

import (
	"fmt"
	"sync"

	jxlog "github.com/Inkedi9/jarvinx/jxlog"
)

type EventType string

const (
	EventObserved EventType = "observed"
	EventDecided  EventType = "decided"
	EventExecuted EventType = "executed"
	EventError    EventType = "error"
)

type Event struct {
	Type    EventType
	Payload any
}

// subscriber représente un consommateur enregistré
type subscriber struct {
	ch   chan Event
	name string
}

// Bus — pub/sub avec fan-out vers tous les subscribers
type Bus struct {
	mu          sync.RWMutex
	subscribers []*subscriber
	bufferSize  int
}

func NewBus(bufferSize int) *Bus {
	return &Bus{bufferSize: bufferSize}
}

// Subscribe enregistre un nouveau consommateur et retourne son canal dédié
func (b *Bus) Subscribe(name string) <-chan Event {
	b.mu.Lock()
	defer b.mu.Unlock()

	sub := &subscriber{
		ch:   make(chan Event, b.bufferSize),
		name: name,
	}
	b.subscribers = append(b.subscribers, sub)

	jxlog.Debug("BUS", fmt.Sprintf("Subscriber enregistré : %s", name))
	return sub.ch
}

// Publish envoie l'événement à tous les subscribers — fan-out
func (b *Bus) Publish(e Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, sub := range b.subscribers {
		select {
		case sub.ch <- e:
			// événement livré
		default:
			jxlog.Warn("BUS", fmt.Sprintf("Buffer plein pour '%s' — événement '%s' ignoré",
				sub.name, e.Type))
		}
	}
}

// Unsubscribe retire un subscriber par nom et ferme son canal
func (b *Bus) Unsubscribe(name string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for i, sub := range b.subscribers {
		if sub.name == name {
			close(sub.ch)
			b.subscribers = append(b.subscribers[:i], b.subscribers[i+1:]...)
			jxlog.Debug("BUS", fmt.Sprintf("Subscriber retiré : %s", name))
			return
		}
	}
}

// Len retourne le nombre de subscribers actifs
func (b *Bus) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.subscribers)
}
