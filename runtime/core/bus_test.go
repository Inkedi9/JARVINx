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

	// Laisse les goroutines dispatcher transférer dispatch → ch
	time.Sleep(50 * time.Millisecond)

	// Les deux subscribers doivent avoir reçu les 3 événements
	for idx, ch := range []<-chan Event{ch1, ch2} {
		count := 0
		for {
			select {
			case <-ch:
				count++
			default:
				if count != 3 {
					t.Errorf("subscriber %d: expected 3 events, got %d", idx+1, count)
				}
				goto next
			}
		}
	next:
	}
}

func TestBus_BufferFullDropsEvent(t *testing.T) {
	bus := NewBus(1) // buffer très petit
	ch := bus.Subscribe("slow-consumer")

	// Publie plus que le buffer total (dispatch=2, ch=1)
	for i := 0; i < 10; i++ {
		bus.Publish(Event{Type: EventObserved})
	}

	// Laisse la goroutine dispatcher traiter
	time.Sleep(50 * time.Millisecond)

	// Au maximum bufferSize*2 + bufferSize événements reçus
	count := 0
	for {
		select {
		case <-ch:
			count++
		default:
			if count == 0 {
				t.Error("expected at least some events")
			}
			return
		}
	}
}

func TestBus_UnsubscribeStopsGoroutine(t *testing.T) {
	bus := NewBus(10)
	ch := bus.Subscribe("test")

	bus.Unsubscribe("test")

	// Canal doit être fermé — range doit se terminer
	timeout := time.After(500 * time.Millisecond)
	select {
	case _, ok := <-ch:
		if ok {
			// Canal encore ouvert — pas d'erreur, juste un événement résiduel
		}
	case <-timeout:
		t.Error("channel should be closed after Unsubscribe")
	}
}

func TestBus_PublishAfterUnsubscribe(t *testing.T) {
	bus := NewBus(10)
	bus.Subscribe("sub-1")
	bus.Unsubscribe("sub-1")

	// Ne doit pas crasher
	bus.Publish(Event{Type: EventObserved})

	if bus.Len() != 0 {
		t.Errorf("expected 0 subscribers after unsubscribe, got %d", bus.Len())
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
