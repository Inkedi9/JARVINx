package web

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"time"

	"github.com/Inkedi9/jarvinx/agents"
	"github.com/Inkedi9/jarvinx/config"
	"github.com/Inkedi9/jarvinx/jxlog"
	"github.com/Inkedi9/jarvinx/llm"
	"github.com/Inkedi9/jarvinx/memory"
	"github.com/Inkedi9/jarvinx/tools"
)

type Server struct {
	cfg            *config.Config
	state          *memory.State
	registry       *agents.Registry
	mainLogger     *memory.Logger
	alertLogger    *memory.Logger
	dailyReporter  *agents.DailyReporter
	port           int
	files          embed.FS
	allowedOrigins map[string]bool
}

type StatusResponse struct {
	Online       bool                `json:"online"`
	Model        string              `json:"model"`
	Interval     string              `json:"interval"`
	CycleNum     int                 `json:"cycle_num"`
	Uptime       string              `json:"uptime"`
	DryRun       bool                `json:"dry_run"`
	CircuitState string              `json:"circuit_state"`
	LastCycle    *memory.CycleRecord `json:"last_cycle,omitempty"`
}

type HistoryResponse struct {
	Cycles []memory.CycleRecord `json:"cycles"`
	Total  int                  `json:"total"`
}

type AgentStatusResponse struct {
	Agents []agents.AgentStatus `json:"agents"`
	Total  int                  `json:"total"`
}

type ToggleRequest struct {
	Name string `json:"name"`
}

type ToggleResponse struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
	Message string `json:"message"`
}

type DockerResponse struct {
	Available  bool                   `json:"available"`
	Containers []tools.ContainerState `json:"containers"`
	Total      int                    `json:"total"`
	Running    int                    `json:"running"`
	Exited     int                    `json:"exited"`
}

type LogsStatusResponse struct {
	MainLog  memory.LogStatus `json:"main_log"`
	AlertLog memory.LogStatus `json:"alert_log"`
}

// ── /api/file ────────────────────────────────────────────────────────────────

type FileAgentResponse struct {
	Enabled    bool     `json:"enabled"`
	WatchPaths []string `json:"watch_paths"`
	MaxSizeMB  int64    `json:"max_size_mb"`
	LastRun    string   `json:"last_run,omitempty"`
	RunCount   int      `json:"run_count"`
	AlertCount int      `json:"alert_count"`
	LastError  string   `json:"last_error,omitempty"`
}

// ── /api/daily-report ────────────────────────────────────────────────────────

type DailyReportResponse struct {
	Enabled     bool   `json:"enabled"`
	ScheduledAt string `json:"scheduled_at"` // "08:00"
	LastSent    string `json:"last_sent,omitempty"`
	NextSend    string `json:"next_send,omitempty"`
}

// ── /api/daily-report/send ───────────────────────────────────────────────────

type SendReportResponse struct {
	Sent    bool   `json:"sent"`
	Message string `json:"message"`
}

// ── /api/llm-context ─────────────────────────────────────────────────────────

type LLMContextResponse struct {
	CycleCount     int      `json:"cycle_count"`
	DominantAction string   `json:"dominant_action"`
	AlertRate      float64  `json:"alert_rate"`
	CPUTrend       string   `json:"cpu_trend"`
	RAMTrend       string   `json:"ram_trend"`
	DiskTrend      string   `json:"disk_trend"`
	RecentAlerts   []string `json:"recent_alerts"`
}

var startTime = time.Now()

func NewServer(
	cfg *config.Config,
	state *memory.State,
	registry *agents.Registry,
	mainLogger *memory.Logger,
	alertLogger *memory.Logger,
	dailyReporter *agents.DailyReporter,
	port int,
	files embed.FS,
) *Server {
	// Construit une map pour lookup O(1)
	origins := make(map[string]bool, len(cfg.AllowedOrigins))
	for _, o := range cfg.AllowedOrigins {
		origins[o] = true
	}

	return &Server{
		cfg:            cfg,
		state:          state,
		registry:       registry,
		mainLogger:     mainLogger,
		alertLogger:    alertLogger,
		dailyReporter:  dailyReporter,
		port:           port,
		files:          files,
		allowedOrigins: origins,
	}
}

func (s *Server) Start() {
	mux := http.NewServeMux()

	// Fichiers statiques
	staticFS, _ := fs.Sub(s.files, "static")
	mux.Handle("/static/", http.StripPrefix("/static/",
		http.FileServer(http.FS(staticFS))))

	// Routes API
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/history", s.handleHistory)
	mux.HandleFunc("/api/agents", s.handleAgents)
	mux.HandleFunc("/api/agents/toggle", s.handleAgentToggle)
	mux.HandleFunc("/api/docker", s.handleDocker)
	mux.HandleFunc("/api/logs/status", s.handleLogsStatus)
	mux.HandleFunc("/api/file", s.handleFile)
	mux.HandleFunc("/api/daily-report", s.handleDailyReport)
	mux.HandleFunc("/api/daily-report/send", s.handleDailyReportSend)
	mux.HandleFunc("/api/llm-context", s.handleLLMContext)

	// corsMiddleware est maintenant une méthode — accès à s.allowedOrigins
	handler := s.corsMiddleware(mux)

	addr := fmt.Sprintf(":%d", s.port)
	jxlog.Info("WEB", fmt.Sprintf("Dashboard → http://localhost%s", addr))

	if err := http.ListenAndServe(addr, handler); err != nil {
		jxlog.Error("WEB", fmt.Sprintf("Erreur serveur : %v", err))
	}
}

// corsMiddleware gère les CORS pour Next.js en dev et en prod
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if origin != "" && s.allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Max-Age", "86400")
		}

		if r.Method == http.MethodOptions {
			if origin != "" && s.allowedOrigins[origin] {
				w.WriteHeader(http.StatusNoContent)
			} else {
				w.WriteHeader(http.StatusForbidden)
			}
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ── /api/file ────────────────────────────────────────────────────────────────

func (s *Server) handleFile(w http.ResponseWriter, r *http.Request) {
	resp := FileAgentResponse{
		Enabled:    false,
		WatchPaths: s.cfg.FileWatchPaths,
		MaxSizeMB:  s.cfg.FileMaxSizeMB,
	}

	if a, ok := s.registry.Get("file"); ok {
		status := a.Status()
		resp.Enabled = status.Enabled
		resp.RunCount = status.RunCount
		resp.AlertCount = status.AlertCount
		resp.LastError = status.LastError
		if !status.LastRun.IsZero() {
			resp.LastRun = status.LastRun.Format(time.RFC3339)
		}
	}

	s.writeJSON(w, resp)
}

// ── /api/daily-report ────────────────────────────────────────────────────────

func (s *Server) handleDailyReport(w http.ResponseWriter, r *http.Request) {
	resp := DailyReportResponse{
		Enabled:     s.cfg.DailyReportEnabled,
		ScheduledAt: fmt.Sprintf("%02d:%02d", s.cfg.DailyReportHour, s.cfg.DailyReportMinute),
	}

	if s.dailyReporter != nil {
		lastSent := s.dailyReporter.LastSent()
		if !lastSent.IsZero() {
			resp.LastSent = lastSent.Format(time.RFC3339)
		}

		// Calcule le prochain envoi
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day(),
			s.cfg.DailyReportHour, s.cfg.DailyReportMinute, 0, 0, now.Location())
		if next.Before(now) {
			next = next.Add(24 * time.Hour)
		}
		resp.NextSend = next.Format(time.RFC3339)
	}

	s.writeJSON(w, resp)
}

// ── /api/daily-report/send ───────────────────────────────────────────────────

func (s *Server) handleDailyReportSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.dailyReporter == nil || !s.cfg.DailyReportEnabled {
		s.writeJSON(w, SendReportResponse{
			Sent:    false,
			Message: "DailyReporter non configuré — active JARVINX_DAILY_REPORT=true",
		})
		return
	}

	go s.dailyReporter.SendNow()

	s.writeJSON(w, SendReportResponse{
		Sent:    true,
		Message: "Rapport en cours d'envoi",
	})
}

// ── /api/llm-context ─────────────────────────────────────────────────────────

func (s *Server) handleLLMContext(w http.ResponseWriter, r *http.Request) {
	cycles := s.state.LastCycles(20)
	snapshots := s.state.Last(10)

	ctx := llm.BuildAdaptiveContext(cycles, snapshots, s.cfg.CPUAlertThreshold, s.cfg.RAMAlertThreshold, s.cfg.DiskAlertThreshold)

	s.writeJSON(w, LLMContextResponse{
		CycleCount:     ctx.CycleCount,
		DominantAction: ctx.DominantAction,
		AlertRate:      ctx.AlertRate,
		CPUTrend:       ctx.CPUTrend,
		RAMTrend:       ctx.RAMTrend,
		DiskTrend:      ctx.DiskTrend,
		RecentAlerts:   ctx.RecentAlerts,
	})
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	data, err := s.files.ReadFile("static/index.html")
	if err != nil {
		// 404 propre — ne pas exposer "index.html not found"
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := w.Write(data); err != nil {
		jxlog.Error("WEB", fmt.Sprintf("write index: %v", err))
	}
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	cycles := s.state.LastCycles(1)

	resp := StatusResponse{
		Online:   true,
		Model:    s.cfg.Model,
		Interval: s.cfg.Interval.String(),
		CycleNum: s.state.CycleNum,
		Uptime:   formatUptime(time.Since(startTime)),
		DryRun:   s.cfg.DryRun,
	}

	if stats := s.registry.CircuitStats(); stats != nil {
		resp.CircuitState = stats.State
	} else {
		resp.CircuitState = "unknown"
	}

	if len(cycles) > 0 {
		c := cycles[0]
		resp.LastCycle = &c
	}

	s.writeJSON(w, resp)
}

func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	cycles := s.state.LastCycles(10)

	// Inverser — plus récent en premier
	for i, j := 0, len(cycles)-1; i < j; i, j = i+1, j-1 {
		cycles[i], cycles[j] = cycles[j], cycles[i]
	}

	s.writeJSON(w, HistoryResponse{
		Cycles: cycles,
		Total:  s.state.CycleNum,
	})
}

func (s *Server) handleAgents(w http.ResponseWriter, r *http.Request) {
	statuses := s.registry.Statuses()

	s.writeJSON(w, AgentStatusResponse{
		Agents: statuses,
		Total:  len(statuses),
	})
}

func (s *Server) writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleAgentToggle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ToggleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	agent, found := s.registry.Get(req.Name)
	if !found {
		http.Error(w, fmt.Sprintf("agent '%s' not found", req.Name), http.StatusNotFound)
		return
	}

	// Toggle — inverse l'état actuel
	var msg string
	if agent.IsEnabled() {
		agent.Disable()
		msg = fmt.Sprintf("agent '%s' désactivé", req.Name)
	} else {
		agent.Enable()
		msg = fmt.Sprintf("agent '%s' activé", req.Name)
	}

	jxlog.Info("WEB", msg)

	s.writeJSON(w, ToggleResponse{
		Name:    req.Name,
		Enabled: agent.IsEnabled(),
		Message: msg,
	})
}

func (s *Server) handleLogsStatus(w http.ResponseWriter, r *http.Request) {
	resp := LogsStatusResponse{}

	if s.mainLogger != nil {
		resp.MainLog = s.mainLogger.Status()
	}
	if s.alertLogger != nil {
		resp.AlertLog = s.alertLogger.Status()
	}

	s.writeJSON(w, resp)
}

func formatUptime(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	sec := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm %ds", h, m, sec)
	}
	return fmt.Sprintf("%dm %ds", m, sec)
}

func (s *Server) handleDocker(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if !tools.DockerAvailable() {
		s.writeJSON(w, DockerResponse{Available: false})
		return
	}

	containers, err := tools.ListContainers(ctx)
	if err != nil {
		s.writeJSON(w, DockerResponse{Available: true})
		return
	}

	running := 0
	exited := 0
	for _, c := range containers {
		if c.Running {
			running++
		}
		if c.Exited {
			exited++
		}
	}

	s.writeJSON(w, DockerResponse{
		Available:  true,
		Containers: containers,
		Total:      len(containers),
		Running:    running,
		Exited:     exited,
	})
}
