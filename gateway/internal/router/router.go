package router

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/vnykmshr/nivo/gateway/internal/middleware"
	"github.com/vnykmshr/nivo/gateway/internal/proxy"
	"github.com/vnykmshr/nivo/shared/logger"
	sharedMiddleware "github.com/vnykmshr/nivo/shared/middleware"
)

// Router configures HTTP routes for the API Gateway.
type Router struct {
	gateway   *proxy.Gateway
	validator *middleware.JWTValidator
	logger    *logger.Logger
}

// NewRouter creates a new router with all handlers and middleware.
func NewRouter(gateway *proxy.Gateway, log *logger.Logger) *Router {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		panic("JWT_SECRET environment variable is required")
	}

	return &Router{
		gateway:   gateway,
		validator: middleware.NewJWTValidator(jwtSecret),
		logger:    log,
	}
}

// SetupRoutes configures all HTTP routes for the gateway.
func (r *Router) SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	// Health check endpoint (gateway-level)
	mux.HandleFunc("GET /health", r.healthCheck)
	mux.HandleFunc("GET /api/health", r.healthCheck)

	// Public routes (no authentication required)
	// Authentication endpoints - these should go directly to identity service
	mux.HandleFunc("POST /api/v1/identity/auth/register", r.gateway.ProxyRequest)
	mux.HandleFunc("POST /api/v1/identity/auth/login", r.gateway.ProxyRequest)

	// Protected routes (authentication required)
	// All other API routes require authentication
	authenticatedHandler := r.validator.Authenticate(http.HandlerFunc(r.gateway.ProxyRequest))
	mux.Handle("/api/v1/", authenticatedHandler)

	// Apply middleware chain
	handler := r.applyMiddleware(mux)

	return handler
}

// applyMiddleware applies the middleware chain to the handler.
func (r *Router) applyMiddleware(handler http.Handler) http.Handler {
	// Apply CORS
	handler = sharedMiddleware.CORS(sharedMiddleware.DefaultCORSConfig())(handler)

	// Apply request ID generation
	handler = sharedMiddleware.RequestID()(handler)

	// Apply logging
	handler = sharedMiddleware.Logging(r.logger)(handler)

	// Apply panic recovery
	handler = sharedMiddleware.Recovery(r.logger)(handler)

	// Apply rate limiting (gateway-wide)
	handler = sharedMiddleware.RateLimit(sharedMiddleware.DefaultRateLimitConfig())(handler)

	return handler
}

// healthCheck is the gateway health check endpoint.
func (r *Router) healthCheck(w http.ResponseWriter, req *http.Request) {
	health := map[string]interface{}{
		"status":  "healthy",
		"service": "gateway",
		"version": "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(health); err != nil {
		r.logger.WithError(err).Error("failed to encode health check response")
	}
}
