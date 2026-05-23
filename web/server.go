package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Inkedi9/jarvinx/config"
	"github.com/Inkedi9/jarvinx/memory"
)

type Server struct {
	cfg   *config.Config
	state *memory.State
	port  int
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

var startTime = time.Now()

func NewServer(cfg *config.Config, state *memory.State, port int) *Server {
	return &Server{cfg: cfg, state: state, port: port}
}

func (s *Server) Start() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/api/status", s.withCORS(s.handleStatus))
	mux.HandleFunc("/api/history", s.withCORS(s.handleHistory))

	addr := fmt.Sprintf(":%d", s.port)
	fmt.Printf("[ WEB ] Dashboard → http://localhost%s\n", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Printf("[ WEB ] Erreur serveur : %v\n", err)
	}
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(indexHTML))
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

	// Inverser pour avoir le plus récent en premier
	for i, j := 0, len(cycles)-1; i < j; i, j = i+1, j-1 {
		cycles[i], cycles[j] = cycles[j], cycles[i]
	}

	resp := HistoryResponse{
		Cycles: cycles,
		Total:  s.state.CycleNum,
	}

	s.writeJSON(w, resp)
}

func (s *Server) writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) withCORS(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		h(w, r)
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
