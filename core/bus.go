package core

import "fmt"

// EventType définit les types d'événements possibles dans JARVINx
type EventType string

const (
	EventObserved EventType = "observed"
	EventDecided  EventType = "decided"
	EventExecuted EventType = "executed"
	EventLogged   EventType = "logged"
	EventError    EventType = "error"
)

// Event est le message qui circule dans le bus
type Event struct {
	Type    EventType
	Payload any
}

// Bus est un canal de communication central entre les composants
type Bus struct {
	ch chan Event
}

func NewBus(bufferSize int) *Bus {
	return &Bus{
		ch: make(chan Event, bufferSize),
	}
}

func (b *Bus) Publish(e Event) {
	select {
	case b.ch <- e:
		// événement publié
	default:
		fmt.Printf("[ BUS ] Avertissement : bus plein, événement '%s' ignoré\n", e.Type)
	}
}

func (b *Bus) Subscribe() <-chan Event {
	return b.ch
}
