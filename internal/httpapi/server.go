package httpapi

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

type ReadinessChecker interface {
	Check(context.Context) (map[string]string, bool)
}

type Server struct {
	handler http.Handler
}

type healthResponse struct {
	Status       string            `json:"status"`
	Dependencies map[string]string `json:"dependencies,omitempty"`
}

type problem struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail,omitempty"`
}

func New(checker ReadinessChecker, logger *slog.Logger) *Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health/live", handleLiveness)
	mux.HandleFunc("GET /api/v1/health/ready", handleReadiness(checker))
	mux.HandleFunc("POST /api/v1/auth/login", handleAuthNotImplemented)
	mux.HandleFunc("POST /api/v1/auth/refresh", handleAuthNotImplemented)
	mux.HandleFunc("POST /api/v1/auth/logout", handleAuthNotImplemented)

	return &Server{handler: requestLogger(logger, mux)}
}

func (s *Server) Handler() http.Handler {
	return s.handler
}

func handleLiveness(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
}

func handleReadiness(checker ReadinessChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		dependencies, ready := checker.Check(ctx)
		if !ready {
			writeJSON(w, http.StatusServiceUnavailable, healthResponse{
				Status:       "degraded",
				Dependencies: dependencies,
			})
			return
		}

		writeJSON(w, http.StatusOK, healthResponse{Status: "ok", Dependencies: dependencies})
	}
}

func handleAuthNotImplemented(w http.ResponseWriter, _ *http.Request) {
	writeProblem(w, http.StatusNotImplemented, "Identity phase not implemented")
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeProblem(w http.ResponseWriter, status int, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(problem{
		Type:   "about:blank",
		Title:  http.StatusText(status),
		Status: status,
		Detail: detail,
	})
}

func requestLogger(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		next.ServeHTTP(w, r)
		logger.InfoContext(r.Context(), "request handled",
			"method", r.Method,
			"path", r.URL.Path,
			"duration_ms", time.Since(started).Milliseconds(),
		)
	})
}
