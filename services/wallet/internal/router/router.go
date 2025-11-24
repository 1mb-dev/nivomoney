package router

import (
	"net/http"

	"github.com/vnykmshr/nivo/services/wallet/internal/handler"
	"github.com/vnykmshr/nivo/shared/metrics"
	"github.com/vnykmshr/nivo/shared/middleware"
)

// SetupRoutes configures all routes for the wallet service using Go 1.22+ stdlib router.
func SetupRoutes(walletHandler *handler.WalletHandler, jwtSecret string) http.Handler {
	mux := http.NewServeMux()

	// Health check endpoint (public)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"healthy","service":"wallet"}`))
	})

	// Metrics endpoint
	mux.Handle("GET /metrics", metrics.Handler())

	// Setup auth middleware
	authConfig := middleware.AuthConfig{
		JWTSecret: jwtSecret,
		SkipPaths: []string{"/health"},
	}
	authMiddleware := middleware.Auth(authConfig)

	// Permission middleware
	createWalletPerm := middleware.RequirePermission("wallet:wallet:create")
	readWalletPerm := middleware.RequirePermission("wallet:wallet:read")
	manageWalletPerm := middleware.RequireAnyPermission("wallet:wallet:activate", "wallet:wallet:freeze", "wallet:wallet:close")

	// ========================================================================
	// Wallet Management Endpoints
	// ========================================================================

	// Wallet CRUD operations
	mux.Handle("POST /api/v1/wallets", authMiddleware(createWalletPerm(http.HandlerFunc(walletHandler.CreateWallet))))
	mux.Handle("GET /api/v1/wallets/{id}", authMiddleware(readWalletPerm(http.HandlerFunc(walletHandler.GetWallet))))
	mux.Handle("GET /api/v1/wallets/{id}/balance", authMiddleware(readWalletPerm(http.HandlerFunc(walletHandler.GetWalletBalance))))

	// Wallet status management (admin/support operations)
	mux.Handle("POST /api/v1/wallets/{id}/activate", authMiddleware(manageWalletPerm(http.HandlerFunc(walletHandler.ActivateWallet))))
	mux.Handle("POST /api/v1/wallets/{id}/freeze", authMiddleware(manageWalletPerm(http.HandlerFunc(walletHandler.FreezeWallet))))
	mux.Handle("POST /api/v1/wallets/{id}/unfreeze", authMiddleware(manageWalletPerm(http.HandlerFunc(walletHandler.UnfreezeWallet))))
	mux.Handle("POST /api/v1/wallets/{id}/close", authMiddleware(manageWalletPerm(http.HandlerFunc(walletHandler.CloseWallet))))

	// User wallets listing
	mux.Handle("GET /api/v1/users/{userId}/wallets", authMiddleware(readWalletPerm(http.HandlerFunc(walletHandler.ListUserWallets))))

	// Apply middleware chain
	metricsCollector := metrics.NewCollector("wallet")
	handler := metricsCollector.Middleware("wallet")(mux)

	// Apply CORS
	corsMiddleware := middleware.CORS(middleware.DefaultCORSConfig())
	return corsMiddleware(handler)
}
