package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/laura838/shld-i-buy-this/agent"
	"github.com/laura838/shld-i-buy-this/cache"
	"github.com/laura838/shld-i-buy-this/db"
	"github.com/laura838/shld-i-buy-this/models"
	"github.com/redis/go-redis/v9"
)

// Server holds the dependencies needed to handle HTTP requests.
type Server struct {
	db      *sql.DB
	profile *models.UserProfile
	redis   *redis.Client
}

// NewServer creates a Server with the given dependencies.
func NewServer(db *sql.DB, profile *models.UserProfile, redisClient *redis.Client) *Server {
	return &Server{db: db, profile: profile, redis: redisClient}
}

type evaluateRequest struct {
	Product string `json:"product"`
}

type evaluateResponse struct {
	ProductName  string              `json:"product_name"`
	Verdict      models.Verdict      `json:"verdict"`
	Reasoning    string              `json:"reasoning"`
	Alternatives []models.Alternative `json:"alternatives,omitempty"`
}

type errorResponse struct {
	Error   string         `json:"error"`
	Code    string         `json:"code,omitempty"`
	Verdict models.Verdict `json:"verdict,omitempty"`
}

// Start wires routes and starts listening on port 8080.
func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/evaluate", loggingMiddleware(corsMiddleware(s.handleEvaluate)))
	mux.HandleFunc("/sessions", loggingMiddleware(corsMiddleware(s.handleSessions)))
	mux.HandleFunc("/profile", loggingMiddleware(corsMiddleware(s.handleProfile)))
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

	ctx := context.Background()

	// 1. Rate limit: 10 requests per hour globally.
	ok, err := cache.CheckRateLimit(ctx, s.redis, "ratelimit:global", 10, time.Hour)
	if err != nil {
		log.Printf("rate limit check error: %v", err)
	} else if !ok {
		writeJSON(w, http.StatusTooManyRequests, errorResponse{Error: "rate limit exceeded"})
		return
	}

	productKey := strings.ToLower(strings.TrimSpace(req.Product))

	// 2. Cooldown: reject if this product was recently evaluated with do_not_buy or wait.
	cooldownVerdict, err := cache.CheckCooldown(ctx, s.redis, "cooldown:"+productKey)
	if err != nil {
		log.Printf("cooldown check error: %v", err)
	} else if cooldownVerdict != "" {
		writeJSON(w, http.StatusTooManyRequests, errorResponse{Error: "you recently evaluated this product", Verdict: models.Verdict(cooldownVerdict)})
		return
	}

	// 3. Cache: return cached decision if available.
	cached, err := cache.GetCachedDecision(ctx, s.redis, "cache:"+productKey)
	if err != nil {
		log.Printf("cache get error: %v", err)
	} else if cached != nil {
		log.Printf("cache hit for %q", productKey)
		writeJSON(w, http.StatusOK, evaluateResponse{ProductName: cached.ProductName, Verdict: cached.Label, Reasoning: cached.Reasoning, Alternatives: cached.Alternatives})
		return
	}

	if s.profile == nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "profile not set up yet"})
		return
	}

	// 4. Evaluate via Claude.
	decision, err := agent.Evaluate(req.Product, *s.profile)
	if err != nil {
		var agentErr *agent.AgentError
		if errors.As(err, &agentErr) {
			writeJSON(w, http.StatusInternalServerError, errorResponse{
				Error: agentErr.Message,
				Code:  string(agentErr.Code),
			})
		} else {
			log.Printf("unclassified evaluate error: %v", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{
				Error: "Something went wrong — try again.",
				Code:  string(agent.ErrCodeUnknown),
			})
		}
		return
	}

	// 5. Cache the decision under both the canonical product_name and the raw
	// input key so either path hits on subsequent lookups.
	cacheKey := strings.ToLower(strings.TrimSpace(decision.ProductName))
	if cacheKey == "" {
		cacheKey = productKey
	}
	if err := cache.SetCachedDecision(ctx, s.redis, "cache:"+cacheKey, decision, 24*time.Hour); err != nil {
		log.Printf("cache set error: %v", err)
	}
	if cacheKey != productKey {
		if err := cache.SetCachedDecision(ctx, s.redis, "cache:"+productKey, decision, 24*time.Hour); err != nil {
			log.Printf("cache set error (raw key): %v", err)
		}
	}

	// 6. Set cooldown for do_not_buy and wait verdicts under both keys.
	if decision.Label == models.VerdictDoNotBuy || decision.Label == models.VerdictWait {
		verdict := string(decision.Label)
		if err := cache.SetCooldown(ctx, s.redis, "cooldown:"+cacheKey, verdict, time.Hour); err != nil {
			log.Printf("cooldown set error: %v", err)
		}
		if cacheKey != productKey {
			if err := cache.SetCooldown(ctx, s.redis, "cooldown:"+productKey, verdict, time.Hour); err != nil {
				log.Printf("cooldown set error (raw key): %v", err)
			}
		}
	}

	session := models.Session{
		Product:      req.Product,
		ProductName:  decision.ProductName,
		Verdict:      decision.Label,
		Reasoning:    decision.Reasoning,
		Alternatives: decision.Alternatives,
		CreatedAt:    time.Now(),
	}
	if err := db.WriteSession(s.db, session); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, evaluateResponse{ProductName: decision.ProductName, Verdict: decision.Label, Reasoning: decision.Reasoning, Alternatives: decision.Alternatives})
}

type sessionResponse struct {
	ID           int                  `json:"id"`
	Product      string               `json:"product"`
	ProductName  string               `json:"product_name"`
	Verdict      models.Verdict       `json:"verdict"`
	Reasoning    string               `json:"reasoning"`
	Alternatives []models.Alternative `json:"alternatives,omitempty"`
	CreatedAt    time.Time            `json:"created_at"`
}

// handleSessions returns sessions ordered by created_at descending.
// Supports optional ?limit=N&offset=N for pagination; omitting limit returns all rows.
func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var rows *sql.Rows
	var err error
	limitStr := r.URL.Query().Get("limit")
	asc := r.URL.Query().Get("asc") == "1"
	order := "DESC"
	if asc {
		order = "ASC"
	}
	if limitStr != "" {
		limit, _ := strconv.Atoi(limitStr)
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		rows, err = s.db.Query(
			"SELECT id, product, product_name, verdict, reasoning, alternatives, created_at FROM sessions ORDER BY created_at "+order+" LIMIT $1 OFFSET $2",
			limit, offset,
		)
	} else {
		rows, err = s.db.Query("SELECT id, product, product_name, verdict, reasoning, alternatives, created_at FROM sessions ORDER BY created_at " + order)
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}
	defer rows.Close()

	sessions := []sessionResponse{}
	for rows.Next() {
		var sess sessionResponse
		var altsJSON []byte
		if err := rows.Scan(&sess.ID, &sess.Product, &sess.ProductName, &sess.Verdict, &sess.Reasoning, &altsJSON, &sess.CreatedAt); err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: err.Error()})
			return
		}
		if len(altsJSON) > 0 {
			json.Unmarshal(altsJSON, &sess.Alternatives)
		}
		sessions = append(sessions, sess)
	}
	if err := rows.Err(); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, sessions)
}

type profileRequest struct {
	Name             string `json:"name"`
	PurchasePatterns string `json:"purchase_patterns"`
}

type profileResponse struct {
	Name             string `json:"name"`
	PurchasePatterns string `json:"purchase_patterns"`
}

// handleProfile handles GET /profile and PUT /profile.
func (s *Server) handleProfile(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		profile, err := db.ReadProfile(s.db)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: err.Error()})
			return
		}
		if profile == nil {
			writeJSON(w, http.StatusOK, struct{}{})
			return
		}
		writeJSON(w, http.StatusOK, profileResponse{Name: profile.Name, PurchasePatterns: profile.PurchasePatterns})

	case http.MethodPut:
		var req profileRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
			return
		}
		profile := models.UserProfile{Name: req.Name, PurchasePatterns: req.PurchasePatterns}
		if err := db.WriteProfile(s.db, profile); err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: err.Error()})
			return
		}
		s.profile = &profile
		writeJSON(w, http.StatusOK, profileResponse{Name: profile.Name, PurchasePatterns: profile.PurchasePatterns})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// corsMiddleware adds CORS headers to allow requests from any origin and
// responds to preflight OPTIONS requests, which browsers send before a
// cross-origin POST with a JSON body.
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
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
