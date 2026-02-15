package handler

import (
	"net/http"
	"os"

	"github.com/1mb-dev/nivomoney/services/risk/internal/service"
	"github.com/1mb-dev/nivomoney/shared/logger"
	"github.com/1mb-dev/nivomoney/shared/metrics"
	"github.com/1mb-dev/nivomoney/shared/middleware"
)

// Router handles HTTP routing for the Risk Service
type Router struct {
	riskHandler *RiskHandler
	metrics     *metrics.Collector
}

// NewRouter creates a new router
func NewRouter(riskService *service.RiskService) *Router {
	return &Router{
		riskHandler: NewRiskHandler(riskService),
		metrics:     metrics.NewCollector("risk"),
	}
}

// SetupRoutes sets up all HTTP routes
func (r *Router) SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("GET /health", r.healthCheck)

	// Metrics endpoint
	mux.Handle("GET /metrics", metrics.Handler())

	// Risk evaluation endpoint (called by transaction service - internal only)
	mux.HandleFunc("POST /api/v1/risk/evaluate", r.riskHandler.EvaluateTransaction)

	// Create JWT auth middleware for admin endpoints
	authConfig := middleware.AuthConfig{
		JWTSecret: os.Getenv("JWT_SECRET"),
	}
	jwtAuth := middleware.Auth(authConfig)

	// Risk rules management endpoints (require authentication)
	mux.Handle("GET /api/v1/risk/rules", jwtAuth(http.HandlerFunc(r.riskHandler.GetAllRules)))
	mux.Handle("GET /api/v1/risk/rules/{id}", jwtAuth(http.HandlerFunc(r.riskHandler.GetRuleByID)))
	mux.Handle("POST /api/v1/risk/rules", jwtAuth(http.HandlerFunc(r.riskHandler.CreateRule)))
	mux.Handle("PUT /api/v1/risk/rules/{id}", jwtAuth(http.HandlerFunc(r.riskHandler.UpdateRule)))
	mux.Handle("DELETE /api/v1/risk/rules/{id}", jwtAuth(http.HandlerFunc(r.riskHandler.DeleteRule)))

	// Risk events endpoints (require authentication)
	mux.Handle("GET /api/v1/risk/events/{id}", jwtAuth(http.HandlerFunc(r.riskHandler.GetEventByID)))
	mux.Handle("GET /api/v1/risk/transactions/{transactionId}/events", jwtAuth(http.HandlerFunc(r.riskHandler.GetEventsByTransactionID)))
	mux.Handle("GET /api/v1/risk/users/{userId}/events", jwtAuth(http.HandlerFunc(r.riskHandler.GetEventsByUserID)))

	// Create logger for middleware
	log := logger.NewDefault("risk")

	// Apply middleware using Chain
	handler := middleware.Chain(mux,
		r.metrics.Middleware("risk"),
		middleware.RequestID(),
		middleware.Logging(log),
	)

	return handler
}

// healthCheck handles health check requests
func (r *Router) healthCheck(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"service":"risk","status":"healthy","version":"1.0.0"}`))
}
