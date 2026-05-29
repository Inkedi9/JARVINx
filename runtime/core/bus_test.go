package core

import (
	"testing"
	"time"
)

func TestBus_SingleSubscriber(t *testing.T) {
	bus := NewBus(10)
	ch := bus.Subscribe("test")

	bus.Publish(Event{Type: EventObserved, Payload: "data"})

	select {
	case e := <-ch:
		if e.Type != EventObserved {
			t.Errorf("expected EventObserved, got %s", e.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("expected event, got timeout")
	}
}

func TestBus_FanOut(t *testing.T) {
	bus := NewBus(10)

	ch1 := bus.Subscribe("consumer-1")
	ch2 := bus.Subscribe("consumer-2")
	ch3 := bus.Subscribe("consumer-3")

	bus.Publish(Event{Type: EventObserved, Payload: "broadcast"})

	// Les 3 subscribers doivent recevoir l'événement
	for i, ch := range []<-chan Event{ch1, ch2, ch3} {
		select {
		case e := <-ch:
			if e.Type != EventObserved {
				t.Errorf("subscriber %d: expected EventObserved, got %s", i+1, e.Type)
			}
		case <-time.After(100 * time.Millisecond):
			t.Errorf("subscriber %d did not receive event", i+1)
		}
	}
}

func TestBus_NoEventStealing(t *testing.T) {
	bus := NewBus(10)

	ch1 := bus.Subscribe("consumer-1")
	ch2 := bus.Subscribe("consumer-2")

	// Publie 3 événements
	for i := 0; i < 3; i++ {
		bus.Publish(Event{Type: EventObserved, Payload: i})
	}

	// Les deux subscribers doivent avoir reçu les 3 événements
	for _, ch := range []<-chan Event{ch1, ch2} {
		count := 0
		for {
			select {
			case <-ch:
				count++
			default:
				if count != 3 {
					t.Errorf("expected 3 events, got %d", count)
				}
				goto next
			}
		}
	next:
	}
}

func TestBus_BufferFullDropsEvent(t *testing.T) {
	bus := NewBus(2) // buffer très petit
	ch := bus.Subscribe("slow-consumer")

	// Publie plus que le buffer
	bus.Publish(Event{Type: EventObserved})
	bus.Publish(Event{Type: EventObserved})
	bus.Publish(Event{Type: EventObserved}) // celui-ci sera droppé

	// Doit avoir exactement 2 événements dans le buffer
	count := 0
	for {
		select {
		case <-ch:
			count++
		default:
			goto done
		}
	}
done:
	if count != 2 {
		t.Errorf("expected 2 events (buffer size), got %d", count)
	}
}

func TestBus_Unsubscribe(t *testing.T) {
	bus := NewBus(10)
	bus.Subscribe("sub-1")
	bus.Subscribe("sub-2")

	if bus.Len() != 2 {
		t.Errorf("expected 2 subscribers, got %d", bus.Len())
	}

	bus.Unsubscribe("sub-1")

	if bus.Len() != 1 {
		t.Errorf("expected 1 subscriber after unsubscribe, got %d", bus.Len())
	}
}

func TestBus_PublishMultipleTypes(t *testing.T) {
	bus := NewBus(10)
	ch := bus.Subscribe("consumer")

	events := []EventType{EventObserved, EventDecided, EventExecuted, EventError}
	for _, et := range events {
		bus.Publish(Event{Type: et})
	}

	received := make([]EventType, 0, 4)
	for i := 0; i < 4; i++ {
		select {
		case e := <-ch:
			received = append(received, e.Type)
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("timeout waiting for event %d", i)
		}
	}

	if len(received) != 4 {
		t.Errorf("expected 4 events, got %d", len(received))
	}
}

func TestBus_Len(t *testing.T) {
	bus := NewBus(10)

	if bus.Len() != 0 {
		t.Error("expected 0 subscribers initially")
	}

	bus.Subscribe("a")
	bus.Subscribe("b")

	if bus.Len() != 2 {
		t.Errorf("expected 2 subscribers, got %d", bus.Len())
	}
}
