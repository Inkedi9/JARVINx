package agents

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	jxlog "github.com/Inkedi9/jarvinx/jxlog"
)

// Notifier est l'interface que tout canal de notification doit implémenter
type Notifier interface {
	Name() string
	Send(alert Alert) error
}

// NotifierDispatcher gère plusieurs notifiers
type NotifierDispatcher struct {
	notifiers  []Notifier
	dryRun     bool
	httpClient *http.Client
}

func NewNotifierDispatcher(dryRun bool) *NotifierDispatcher {
	return &NotifierDispatcher{
		dryRun:     dryRun,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (d *NotifierDispatcher) Register(n Notifier) {
	d.notifiers = append(d.notifiers, n)
	jxlog.Info("NOTIFIER", fmt.Sprintf("Canal enregistré : %s", n.Name()))
}

func (d *NotifierDispatcher) Dispatch(alert Alert) {
	if len(d.notifiers) == 0 {
		return
	}

	for _, n := range d.notifiers {
		if d.dryRun {
			jxlog.Info("DRY-RUN", fmt.Sprintf(
				"[%s] alert simulée — %s : %s", n.Name(), alert.Metric, alert.Message,
			))
			continue
		}

		if err := n.Send(alert); err != nil {
			jxlog.Error("NOTIFIER", fmt.Sprintf("[%s] échec envoi : %v", n.Name(), err))
		} else {
			jxlog.Info("NOTIFIER", fmt.Sprintf("[%s] notifié — %s", n.Name(), alert.Metric))
		}
	}
}

// ── Discord ──────────────────────────────────────────────────────────────────

type DiscordNotifier struct {
	webhookURL string
	client     *http.Client
}

func NewDiscordNotifier(webhookURL string) *DiscordNotifier {
	return &DiscordNotifier{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (n *DiscordNotifier) Name() string { return "discord" }

func (n *DiscordNotifier) Send(alert Alert) error {
	color := 16776960 // jaune warning
	if alert.Level == AlertCritical {
		color = 16711680 // rouge critical
	}

	emoji := "⚠️"
	if alert.Level == AlertCritical {
		emoji = "🚨"
	}

	payload := map[string]any{
		"username": "JARVINx",
		"embeds": []map[string]any{
			{
				"title":       emoji + " " + string(alert.Level) + " — " + alert.Metric,
				"description": alert.Message,
				"color":       color,
				"fields": []map[string]any{
					{"name": "Valeur", "value": fmt.Sprintf("%.1f%%", alert.Value), "inline": true},
					{"name": "Seuil", "value": fmt.Sprintf("%.1f%%", alert.Threshold), "inline": true},
					{"name": "Cycles", "value": fmt.Sprintf("%d", alert.CyclesAbove), "inline": true},
				},
				"footer":    map[string]any{"text": "JARVINx · Autonomous Agent Runtime"},
				"timestamp": alert.Timestamp.Format(time.RFC3339),
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	resp, err := n.client.Post(n.webhookURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("post: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("discord status: %d", resp.StatusCode)
	}

	return nil
}

// ── Slack ────────────────────────────────────────────────────────────────────

type SlackNotifier struct {
	webhookURL string
	client     *http.Client
}

func NewSlackNotifier(webhookURL string) *SlackNotifier {
	return &SlackNotifier{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (n *SlackNotifier) Name() string { return "slack" }

func (n *SlackNotifier) Send(alert Alert) error {
	emoji := ":warning:"
	if alert.Level == AlertCritical {
		emoji = ":rotating_light:"
	}

	text := fmt.Sprintf("%s *%s — %s*\n%s\nValeur: `%.1f%%` | Seuil: `%.1f%%`",
		emoji,
		string(alert.Level),
		alert.Metric,
		alert.Message,
		alert.Value,
		alert.Threshold,
	)

	payload := map[string]any{"text": text}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	resp, err := n.client.Post(n.webhookURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("post: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack status: %d", resp.StatusCode)
	}

	return nil
}

// ── Ntfy ─────────────────────────────────────────────────────────────────────

type NtfyNotifier struct {
	serverURL string
	topic     string
	client    *http.Client
}

func NewNtfyNotifier(serverURL, topic string) *NtfyNotifier {
	return &NtfyNotifier{
		serverURL: serverURL,
		topic:     topic,
		client:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (n *NtfyNotifier) Name() string { return "ntfy" }

func (n *NtfyNotifier) Send(alert Alert) error {
	priority := "default"
	if alert.Level == AlertCritical {
		priority = "urgent"
	}

	url := fmt.Sprintf("%s/%s", n.serverURL, n.topic)

	req, err := http.NewRequest("POST", url, bytes.NewBufferString(alert.Message))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Title", fmt.Sprintf("JARVINx — %s %s", alert.Level, alert.Metric))
	req.Header.Set("Priority", priority)
	req.Header.Set("Tags", "jarvinx,alert")

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("post: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ntfy status: %d", resp.StatusCode)
	}

	return nil
}

// ── Gotify ───────────────────────────────────────────────────────────────────

type GotifyNotifier struct {
	serverURL string
	token     string
	client    *http.Client
}

func NewGotifyNotifier(serverURL, token string) *GotifyNotifier {
	return &GotifyNotifier{
		serverURL: serverURL,
		token:     token,
		client:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (n *GotifyNotifier) Name() string { return "gotify" }

func (n *GotifyNotifier) Send(alert Alert) error {
	priority := 5
	if alert.Level == AlertCritical {
		priority = 10
	}

	payload := map[string]any{
		"title":    fmt.Sprintf("JARVINx — %s %s", alert.Level, alert.Metric),
		"message":  alert.Message,
		"priority": priority,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	url := fmt.Sprintf("%s/message", n.serverURL)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Gotify-Key", n.token)

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("post: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("gotify status: %d", resp.StatusCode)
	}

	return nil
}
