package agents

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// mockNotifier — notifier de test
type mockNotifier struct {
	name      string
	sendCount int
	lastAlert Alert
	shouldErr bool
}

func (m *mockNotifier) Name() string { return m.name }
func (m *mockNotifier) Send(alert Alert) error {
	m.sendCount++
	m.lastAlert = alert
	if m.shouldErr {
		return fmt.Errorf("mock error")
	}
	return nil
}

func TestNotifierDispatcher_Register(t *testing.T) {
	d := NewNotifierDispatcher(false)
	d.Register(&mockNotifier{name: "test"})

	if len(d.notifiers) != 1 {
		t.Errorf("expected 1 notifier, got %d", len(d.notifiers))
	}
}

func TestNotifierDispatcher_DispatchToAll(t *testing.T) {
	d := NewNotifierDispatcher(false)

	n1 := &mockNotifier{name: "notifier-1"}
	n2 := &mockNotifier{name: "notifier-2"}
	d.Register(n1)
	d.Register(n2)

	alert := Alert{
		Timestamp: time.Now(),
		Level:     AlertWarning,
		Metric:    "DISK",
		Value:     85.0,
		Threshold: 85.0,
		Message:   "test alert",
	}

	d.Dispatch(alert)

	if n1.sendCount != 1 {
		t.Errorf("expected notifier-1 to receive 1 alert, got %d", n1.sendCount)
	}
	if n2.sendCount != 1 {
		t.Errorf("expected notifier-2 to receive 1 alert, got %d", n2.sendCount)
	}
}

func TestNotifierDispatcher_DryRunSkipsSend(t *testing.T) {
	d := NewNotifierDispatcher(true) // dry-run

	n := &mockNotifier{name: "test"}
	d.Register(n)

	d.Dispatch(Alert{Metric: "CPU", Level: AlertCritical})

	if n.sendCount != 0 {
		t.Errorf("dry-run should not send, got %d sends", n.sendCount)
	}
}

func TestNotifierDispatcher_ErrorDoesNotStopOthers(t *testing.T) {
	d := NewNotifierDispatcher(false)

	bad := &mockNotifier{name: "bad", shouldErr: true}
	good := &mockNotifier{name: "good", shouldErr: false}
	d.Register(bad)
	d.Register(good)

	d.Dispatch(Alert{Metric: "CPU", Level: AlertCritical, Timestamp: time.Now()})

	if good.sendCount != 1 {
		t.Errorf("good notifier should still receive alert despite bad notifier error")
	}
}

func TestNotifierDispatcher_EmptyDoesNothing(t *testing.T) {
	d := NewNotifierDispatcher(false)
	// Pas de panic avec 0 notifiers
	d.Dispatch(Alert{Metric: "CPU"})
}

func TestDiscordNotifier_Name(t *testing.T) {
	n := NewDiscordNotifier("https://example.com")
	if n.Name() != "discord" {
		t.Errorf("expected 'discord', got '%s'", n.Name())
	}
}

func TestDiscordNotifier_Send(t *testing.T) {
	// Serveur HTTP de test qui simule Discord
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var payload map[string]any
		_ = json.NewDecoder(r.Body).Decode(&payload)

		if _, ok := payload["embeds"]; !ok {
			t.Error("expected embeds in Discord payload")
		}

		w.WriteHeader(http.StatusNoContent) // Discord retourne 204
	}))
	defer srv.Close()

	n := NewDiscordNotifier(srv.URL)
	err := n.Send(Alert{
		Timestamp: time.Now(),
		Level:     AlertCritical,
		Metric:    "CPU",
		Value:     90.0,
		Threshold: 85.0,
		Message:   "CPU critique",
	})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestSlackNotifier_Name(t *testing.T) {
	n := NewSlackNotifier("https://example.com")
	if n.Name() != "slack" {
		t.Errorf("expected 'slack', got '%s'", n.Name())
	}
}

func TestSlackNotifier_Send(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]any
		_ = json.NewDecoder(r.Body).Decode(&payload)

		if _, ok := payload["text"]; !ok {
			t.Error("expected text in Slack payload")
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	n := NewSlackNotifier(srv.URL)
	err := n.Send(Alert{
		Timestamp: time.Now(),
		Level:     AlertWarning,
		Metric:    "DISK",
		Value:     85.0,
		Threshold: 85.0,
		Message:   "Disk warning",
	})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestNtfyNotifier_Name(t *testing.T) {
	n := NewNtfyNotifier("https://ntfy.sh", "test")
	if n.Name() != "ntfy" {
		t.Errorf("expected 'ntfy', got '%s'", n.Name())
	}
}

func TestNtfyNotifier_Send(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Title") == "" {
			t.Error("expected Title header for ntfy")
		}
		if r.Header.Get("Priority") == "" {
			t.Error("expected Priority header for ntfy")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	n := NewNtfyNotifier(srv.URL, "test-topic")
	err := n.Send(Alert{
		Timestamp: time.Now(),
		Level:     AlertCritical,
		Metric:    "RAM",
		Value:     92.0,
		Threshold: 90.0,
		Message:   "RAM critique",
	})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestGotifyNotifier_Name(t *testing.T) {
	n := NewGotifyNotifier("https://gotify.example.com", "token123")
	if n.Name() != "gotify" {
		t.Errorf("expected 'gotify', got '%s'", n.Name())
	}
}

func TestGotifyNotifier_Send(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Token doit être dans le header, pas dans l'URL
		if got := r.Header.Get("X-Gotify-Key"); got != "mytoken" {
			t.Errorf("expected X-Gotify-Key=mytoken, got %q", got)
		}
		if r.URL.Query().Get("token") != "" {
			t.Error("token must not appear as URL query param")
		}

		var payload map[string]any
		_ = json.NewDecoder(r.Body).Decode(&payload)
		if _, ok := payload["title"]; !ok {
			t.Error("expected title in Gotify payload")
		}
		if _, ok := payload["priority"]; !ok {
			t.Error("expected priority in Gotify payload")
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	n := NewGotifyNotifier(srv.URL, "mytoken")
	err := n.Send(Alert{
		Timestamp: time.Now(),
		Level:     AlertWarning,
		Metric:    "DISK",
		Value:     86.0,
		Threshold: 85.0,
		Message:   "Disk warning",
	})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}
