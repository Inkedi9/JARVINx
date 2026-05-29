package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"time"

	"github.com/Inkedi9/jarvinx/agents"
	"github.com/Inkedi9/jarvinx/config"
	"github.com/Inkedi9/jarvinx/jxlog"
	"github.com/Inkedi9/jarvinx/memory"
)

type Server struct {
	cfg            *config.Config
	state          *memory.State
	registry       *agents.Registry
	port           int
	files          embed.FS
	allowedOrigins map[string]bool
}

type StatusResponse struct {
	Online    bool                `json:"online"`
	Model     string              `json:"model"`
	Interval  string              `json:"interval"`
	CycleNum  int                 `json:"cycle_num"`
	Uptime    string              `json:"uptime"`
	LastCycle *memory.CycleRecord `json:"last_cycle,omitempty"`
}

type HistoryResponse struct {
	Cycles []memory.CycleRecord `json:"cycles"`
	Total  int                  `json:"total"`
}

type AgentStatusResponse struct {
	Agents []agents.AgentStatus `json:"agents"`
	Total  int                  `json:"total"`
}

var startTime = time.Now()

func NewServer(cfg *config.Config, state *memory.State, registry *agents.Registry, port int, files embed.FS) *Server {
	// Construit une map pour lookup O(1)
	origins := make(map[string]bool, len(cfg.AllowedOrigins))
	for _, o := range cfg.AllowedOrigins {
		origins[o] = true
	}

	return &Server{
		cfg:            cfg,
		state:          state,
		registry:       registry,
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

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	data, err := s.files.ReadFile("static/index.html")
	if err != nil {
		http.Error(w, "index.html not found", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(data)
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	cycles := s.state.LastCycles(1)

	resp := StatusResponse{
		Online:   true,
		Model:    s.cfg.Model,
		Interval: s.cfg.Interval.String(),
		CycleNum: s.state.CycleNum,
		Uptime:   formatUptime(time.Since(startTime)),
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

func formatUptime(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	sec := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm %ds", h, m, sec)
	}
	return fmt.Sprintf("%dm %ds", m, sec)
}
