package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	simconfig "github.com/1mb-dev/nivomoney/services/simulation/internal/config"
	"github.com/1mb-dev/nivomoney/services/simulation/internal/handler"
	simmetrics "github.com/1mb-dev/nivomoney/services/simulation/internal/metrics"
	"github.com/1mb-dev/nivomoney/services/simulation/internal/service"
	"github.com/1mb-dev/nivomoney/shared/config"
	"github.com/1mb-dev/nivomoney/shared/database"
	"github.com/1mb-dev/nivomoney/shared/metrics"
	"github.com/golang-jwt/jwt/v5"
)

const serviceName = "simulation"

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[%s] Failed to load configuration: %v", serviceName, err)
	}

	// Setup logging
	log.Printf("[%s] Starting Simulation Engine Service...", serviceName)
	log.Printf("[%s] Environment: %s", serviceName, cfg.Environment)
	log.Printf("[%s] Port: %d", serviceName, cfg.ServicePort)

	// Connect to database
	db, err := database.NewFromURL(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("[%s] Failed to connect to database: %v", serviceName, err)
	}
	defer func() { _ = db.Close() }()

	log.Printf("[%s] Connected to database successfully", serviceName)

	// Get Gateway URL and admin token
	gatewayURL := getEnvOrDefault("GATEWAY_URL", "http://gateway:8000")
	adminToken := os.Getenv("ADMIN_TOKEN")
	if adminToken == "" {
		// Generate a service token using JWT_SECRET
		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			log.Fatalf("[%s] Neither ADMIN_TOKEN nor JWT_SECRET set - cannot authenticate", serviceName)
		}
		var err error
		adminToken, err = generateServiceToken(jwtSecret)
		if err != nil {
			log.Fatalf("[%s] Failed to generate service token: %v", serviceName, err)
		}
		log.Printf("[%s] Generated service token (expires in 1 year)", serviceName)
	}

	log.Printf("[%s] Gateway URL: %s", serviceName, gatewayURL)

	// Initialize gateway client
	gatewayClient := service.NewGatewayClient(gatewayURL, adminToken)

	// Initialize simulation configuration
	simulationConfig := simconfig.NewDefaultConfig()

	// Check for demo mode environment variable
	if getEnvOrDefault("SIMULATION_MODE", "realistic") == "demo" {
		simulationConfig = simconfig.NewDemoConfig()
		log.Printf("[%s] Running in DEMO mode", serviceName)
	}

	// Initialize simulation metrics
	simulationMetrics := simmetrics.NewSimulationMetrics()
	simulationMetrics.SetMode(string(simulationConfig.Mode))

	// Initialize Prometheus metrics collector for HTTP request tracking
	metricsCollector := metrics.NewCollector("simulation")

	// Initialize simulation engine with config and metrics
	simulationEngine := service.NewSimulationEngine(db.DB, gatewayClient, simulationConfig, simulationMetrics)

	// Initialize handler with config and metrics
	simulationHandler := handler.NewSimulationHandler(simulationEngine, simulationConfig, simulationMetrics)

	// Setup routes
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"healthy","service":"simulation"}`))
	})

	// Simulation control endpoints
	mux.HandleFunc("GET /api/v1/simulation/status", simulationHandler.GetStatus)
	mux.HandleFunc("POST /api/v1/simulation/start", simulationHandler.StartSimulation)
	mux.HandleFunc("POST /api/v1/simulation/stop", simulationHandler.StopSimulation)

	// Admin configuration endpoints
	mux.HandleFunc("GET /api/v1/simulation/config", simulationHandler.GetConfig)
	mux.HandleFunc("PUT /api/v1/simulation/config", simulationHandler.UpdateConfig)
	mux.HandleFunc("POST /api/v1/simulation/mode", simulationHandler.SetMode)

	// Metrics endpoints (JSON)
	mux.HandleFunc("GET /api/v1/simulation/metrics", simulationHandler.GetMetrics)
	mux.HandleFunc("POST /api/v1/simulation/metrics/reset", simulationHandler.ResetMetrics)

	// Prometheus metrics endpoint
	// Updates simulation-specific gauges before returning standard Prometheus format
	mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) {
		// Update Prometheus gauges from simulation metrics before scrape
		simulationMetrics.UpdatePrometheusMetrics()
		metrics.Handler().ServeHTTP(w, r)
	})

	log.Printf("[%s] Routes configured", serviceName)

	// Create a cancellable context for the simulation engine
	simCtx, simCancel := context.WithCancel(context.Background())

	// Auto-start simulation if enabled
	autoStart := getEnvOrDefault("AUTO_START_SIMULATION", "true")
	if autoStart == "true" {
		log.Printf("[%s] Auto-starting simulation...", serviceName)
		go func() {
			// Wait a bit for services to be ready
			time.Sleep(10 * time.Second)
			simulationEngine.Start(simCtx)
		}()
	}

	// Apply metrics middleware to track HTTP requests
	handler := metricsCollector.Middleware("simulation")(mux)

	// Create HTTP server
	addr := fmt.Sprintf(":%d", cfg.ServicePort)
	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("[%s] Server listening on %s", serviceName, addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[%s] Server failed to start: %v", serviceName, err)
		}
	}()

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Wait for interrupt signal
	<-quit
	log.Printf("[%s] Shutting down server...", serviceName)

	// Stop simulation - cancel context and call Stop()
	simCancel()
	simulationEngine.Stop()

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[%s] Server forced to shutdown: %v", serviceName, err)
	}

	log.Printf("[%s] Server stopped gracefully", serviceName)
}

// getEnvOrDefault returns the environment variable value or a default value.
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// ServiceClaims represents JWT claims for a service token.
type ServiceClaims struct {
	UserID      string   `json:"user_id"`
	Email       string   `json:"email"`
	Status      string   `json:"status"`
	AccountType string   `json:"account_type"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

// generateServiceToken creates a long-lived admin token for the simulation service.
// This token has full admin permissions and expires in 1 year.
func generateServiceToken(jwtSecret string) (string, error) {
	// Full set of admin permissions matching seed service's admin role
	// These are the specific permissions required by each service endpoint
	permissions := []string{
		// Identity service permissions
		"identity:auth:login",
		"identity:auth:logout",
		"identity:auth:refresh",
		"identity:profile:read",
		"identity:profile:update",
		"identity:profile:delete",
		"identity:users:read",
		"identity:users:create",
		"identity:users:update",
		"identity:users:delete",
		"identity:kyc:submit",
		"identity:kyc:read",
		"identity:kyc:verify",
		"identity:kyc:reject",
		"identity:kyc:list",
		"identity:user:suspend",
		"identity:user:unsuspend",
		// Wallet service permissions
		"wallet:wallet:create",
		"wallet:wallet:read",
		"wallet:wallet:update",
		"wallet:wallet:delete",
		"wallet:wallet:list",
		"wallet:wallet:freeze",
		"wallet:wallet:unfreeze",
		"wallet:beneficiary:manage",
		// Transaction service permissions
		"transaction:deposit:create",
		"transaction:transfer:create",
		"transaction:withdrawal:create",
		"transaction:transaction:create",
		"transaction:transaction:list",
		"transaction:transaction:read",
		"transaction:transaction:reverse",
		// Ledger service permissions
		"ledger:journal:read",
		"ledger:journal:list",
		"ledger:account:read",
		"ledger:account:list",
		// RBAC service permissions
		"rbac:role:read",
		"rbac:role:list",
		"rbac:permission:read",
		"rbac:permission:list",
		// Risk service permissions
		"risk:rule:read",
		"risk:rule:list",
		"risk:evaluation:read",
		// Notification service permissions
		"notification:notification:read",
		"notification:notification:list",
		"notification:notification:send",
	}

	claims := &ServiceClaims{
		UserID:      "00000000-0000-0000-0000-000000000000", // System user ID
		Email:       "simulation@system.nivo",
		Status:      "active",
		AccountType: "super_admin",
		Roles:       []string{"super_admin", "admin"},
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(365 * 24 * time.Hour)), // 1 year expiry
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "nivo-simulation",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}
