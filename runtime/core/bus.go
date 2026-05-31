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
	EventLogged   EventType = "logged"
	EventError    EventType = "error"
)

type Event struct {
	Type    EventType
	Payload any
}

// subscriber représente un consommateur enregistré
type subscriber struct {
	name     string
	ch       chan Event // canal exposé au consommateur
	dispatch chan Event // canal interne — reçoit depuis Publish
	quit     chan struct{}
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

// Subscribe enregistre un subscriber et lance sa goroutine de dispatch
func (b *Bus) Subscribe(name string) <-chan Event {
	b.mu.Lock()
	defer b.mu.Unlock()

	sub := &subscriber{
		name:     name,
		ch:       make(chan Event, b.bufferSize),
		dispatch: make(chan Event, b.bufferSize*2), // buffer plus grand côté interne
		quit:     make(chan struct{}),
	}

	// Goroutine dédiée — transfère dispatch → ch
	go sub.run()

	b.subscribers = append(b.subscribers, sub)
	jxlog.Debug("BUS", fmt.Sprintf("Subscriber enregistré : %s", name))
	return sub.ch
}

// run est la goroutine dédiée du subscriber
func (s *subscriber) run() {
	for {
		select {
		case <-s.quit:
			close(s.ch)
			return
		case e := <-s.dispatch:
			select {
			case s.ch <- e:
				// livré au consommateur
			default:
				// consommateur lent — drop avec warning
				// on ne bloque PAS les autres subscribers
			}
		}
	}
}

// Publish envoie l'événement à tous les subscribers — fan-out
// Non-bloquant — ne tient pas le verrou pendant les envois
func (b *Bus) Publish(e Event) {
	b.mu.RLock()
	subs := make([]*subscriber, len(b.subscribers))
	copy(subs, b.subscribers)
	b.mu.RUnlock() // ← verrou libéré AVANT les envois

	for _, sub := range subs {
		select {
		case sub.dispatch <- e:
			// envoyé à la goroutine dédiée
		default:
			jxlog.Warn("BUS", fmt.Sprintf(
				"Buffer plein pour '%s' — événement '%s' ignoré",
				sub.name, e.Type,
			))
		}
	}
}

// Unsubscribe retire un subscriber et arrête sa goroutine
func (b *Bus) Unsubscribe(name string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for i, sub := range b.subscribers {
		if sub.name == name {
			close(sub.quit) // signal d'arrêt à la goroutine
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
