package handler

import (
	"net/http"

	"github.com/vnykmshr/nivo/services/identity/internal/service"
	"github.com/vnykmshr/nivo/shared/middleware"
)

// Router sets up HTTP routes for the Identity Service.
type Router struct {
	authHandler    *AuthHandler
	authMiddleware *AuthMiddleware
}

// NewRouter creates a new router with all handlers and middleware.
func NewRouter(authService *service.AuthService) *Router {
	return &Router{
		authHandler:    NewAuthHandler(authService),
		authMiddleware: NewAuthMiddleware(authService),
	}
}

// SetupRoutes configures all HTTP routes for the Identity Service.
func (r *Router) SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	// Public routes (no authentication required)
	mux.HandleFunc("POST /api/v1/auth/register", r.authHandler.Register)
	mux.HandleFunc("POST /api/v1/auth/login", r.authHandler.Login)

	// Protected routes (authentication required)
	mux.Handle("POST /api/v1/auth/logout",
		r.authMiddleware.Authenticate(http.HandlerFunc(r.authHandler.Logout)))

	mux.Handle("POST /api/v1/auth/logout-all",
		r.authMiddleware.Authenticate(http.HandlerFunc(r.authHandler.LogoutAll)))

	mux.Handle("GET /api/v1/auth/me",
		r.authMiddleware.Authenticate(http.HandlerFunc(r.authHandler.GetProfile)))

	mux.Handle("GET /api/v1/auth/kyc",
		r.authMiddleware.Authenticate(http.HandlerFunc(r.authHandler.GetKYC)))

	mux.Handle("PUT /api/v1/auth/kyc",
		r.authMiddleware.Authenticate(http.HandlerFunc(r.authHandler.UpdateKYC)))

	// Admin routes (authentication + permission required)
	kycVerifyPermission := r.authMiddleware.RequirePermission("identity:kyc:verify")
	kycRejectPermission := r.authMiddleware.RequirePermission("identity:kyc:reject")

	mux.Handle("POST /api/v1/admin/kyc/verify",
		r.authMiddleware.Authenticate(
			kycVerifyPermission(http.HandlerFunc(r.authHandler.VerifyKYC))))

	mux.Handle("POST /api/v1/admin/kyc/reject",
		r.authMiddleware.Authenticate(
			kycRejectPermission(http.HandlerFunc(r.authHandler.RejectKYC))))

	// Health check endpoint
	mux.HandleFunc("GET /health", healthCheck)

	// Apply CORS middleware
	corsMiddleware := middleware.CORS(middleware.DefaultCORSConfig())
	return corsMiddleware(mux)
}

// healthCheck is a simple health check endpoint.
func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"identity"}`))
}
