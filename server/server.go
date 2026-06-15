package server

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/laura838/shld-i-buy-this/agent"
	"github.com/laura838/shld-i-buy-this/db"
	"github.com/laura838/shld-i-buy-this/models"
)

// Server holds the dependencies needed to handle HTTP requests.
type Server struct {
	db      *sql.DB
	profile *models.UserProfile
}

// NewServer creates a Server with the given dependencies.
func NewServer(db *sql.DB, profile *models.UserProfile) *Server {
	return &Server{db: db, profile: profile}
}

type evaluateRequest struct {
	Product string `json:"product"`
}

type evaluateResponse struct {
	Verdict   models.Verdict `json:"verdict"`
	Reasoning string         `json:"reasoning"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// Start wires routes and starts listening on port 8080.
func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/evaluate", loggingMiddleware(corsMiddleware(s.handleEvaluate)))
	mux.HandleFunc("/sessions", loggingMiddleware(corsMiddleware(s.handleSessions)))
	return http.ListenAndServe(":8080", mux)
}

// handleEvaluate handles POST /evaluate.
func (s *Server) handleEvaluate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req evaluateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	decision, err := agent.Evaluate(req.Product, *s.profile)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}

	session := models.Session{
		Product:   req.Product,
		Verdict:   decision.Label,
		Reasoning: decision.Reasoning,
		CreatedAt: time.Now(),
	}
	if err := db.WriteSession(s.db, session); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, evaluateResponse{Verdict: decision.Label, Reasoning: decision.Reasoning})
}

type sessionResponse struct {
	Product   string         `json:"product"`
	Verdict   models.Verdict `json:"verdict"`
	Reasoning string         `json:"reasoning"`
	CreatedAt time.Time      `json:"created_at"`
}

// handleSessions returns all past sessions ordered by created_at descending.
func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	rows, err := s.db.Query("SELECT product, verdict, reasoning, created_at FROM sessions ORDER BY created_at DESC")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}
	defer rows.Close()

	sessions := []sessionResponse{}
	for rows.Next() {
		var sess sessionResponse
		if err := rows.Scan(&sess.Product, &sess.Verdict, &sess.Reasoning, &sess.CreatedAt); err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: err.Error()})
			return
		}
		sessions = append(sessions, sess)
	}
	if err := rows.Err(); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, sessions)
}

// corsMiddleware adds CORS headers to allow requests from any origin and
// responds to preflight OPTIONS requests, which browsers send before a
// cross-origin POST with a JSON body.
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// loggingMiddleware logs each request's method, path, status code, and duration.
func loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next(rec, r)
		log.Printf("%s %s %d %.2fms", r.Method, r.URL.Path, rec.status, float64(time.Since(start).Microseconds())/1000)
	}
}

// statusRecorder wraps http.ResponseWriter to capture the status code written.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
