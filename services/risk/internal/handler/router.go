package handler

import (
	"net/http"

	"github.com/vnykmshr/nivo/services/risk/internal/service"
	"github.com/vnykmshr/nivo/shared/logger"
	"github.com/vnykmshr/nivo/shared/metrics"
	"github.com/vnykmshr/nivo/shared/middleware"
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

	// Risk evaluation endpoint (called by transaction service)
	mux.HandleFunc("POST /api/v1/risk/evaluate", r.riskHandler.EvaluateTransaction)

	// Risk rules management endpoints
	mux.HandleFunc("GET /api/v1/risk/rules", r.riskHandler.GetAllRules)
	mux.HandleFunc("GET /api/v1/risk/rules/{id}", r.riskHandler.GetRuleByID)
	mux.HandleFunc("POST /api/v1/risk/rules", r.riskHandler.CreateRule)
	mux.HandleFunc("PUT /api/v1/risk/rules/{id}", r.riskHandler.UpdateRule)
	mux.HandleFunc("DELETE /api/v1/risk/rules/{id}", r.riskHandler.DeleteRule)

	// Risk events endpoints
	mux.HandleFunc("GET /api/v1/risk/events/{id}", r.riskHandler.GetEventByID)
	mux.HandleFunc("GET /api/v1/risk/transactions/{transactionId}/events", r.riskHandler.GetEventsByTransactionID)
	mux.HandleFunc("GET /api/v1/risk/users/{userId}/events", r.riskHandler.GetEventsByUserID)

	// Create logger for middleware
	log := logger.NewDefault("risk")

	// Apply middleware using Chain
	handler := middleware.Chain(mux,
		r.metrics.Middleware("risk"),
		middleware.RequestID(),
		middleware.Logging(log),
		middleware.CORS(middleware.DefaultCORSConfig()),
	)

	return handler
}

// healthCheck handles health check requests
func (r *Router) healthCheck(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"service":"risk","status":"healthy","version":"1.0.0"}`))
}
