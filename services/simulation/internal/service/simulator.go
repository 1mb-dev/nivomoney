package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/vnykmshr/nivo/services/simulation/internal/personas"
)

// UserWallet represents a user with their wallet
type UserWallet struct {
	UserID   string
	WalletID string
	Email    string
	Persona  personas.PersonaType
}

// SimulationEngine generates synthetic transactions
type SimulationEngine struct {
	db            *sql.DB
	gatewayClient *GatewayClient
	users         []UserWallet
	running       bool
}

// NewSimulationEngine creates a new simulation engine
func NewSimulationEngine(db *sql.DB, gatewayClient *GatewayClient) *SimulationEngine {
	return &SimulationEngine{
		db:            db,
		gatewayClient: gatewayClient,
		users:         make([]UserWallet, 0),
		running:       false,
	}
}

// LoadUsers loads users and wallets from the database
func (s *SimulationEngine) LoadUsers(ctx context.Context) error {
	query := `
		SELECT
			u.id as user_id,
			w.id as wallet_id,
			u.email
		FROM users u
		INNER JOIN wallets w ON u.id = w.user_id
		WHERE u.status = 'active' AND w.status = 'active'
		ORDER BY u.created_at DESC
		LIMIT 50
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	s.users = make([]UserWallet, 0)
	personaTypes := personas.AllPersonaTypes()

	for rows.Next() {
		var uw UserWallet
		if err := rows.Scan(&uw.UserID, &uw.WalletID, &uw.Email); err != nil {
			log.Printf("[simulation] Failed to scan user: %v", err)
			continue
		}

		// Assign random persona
		uw.Persona = personaTypes[rand.Intn(len(personaTypes))]
		s.users = append(s.users, uw)
	}

	log.Printf("[simulation] Loaded %d users for simulation", len(s.users))
	return nil
}

// Start starts the simulation engine
func (s *SimulationEngine) Start(ctx context.Context) {
	if s.running {
		log.Printf("[simulation] Engine already running")
		return
	}

	s.running = true
	log.Printf("[simulation] Starting simulation engine...")

	// Load users first
	if err := s.LoadUsers(ctx); err != nil {
		log.Printf("[simulation] Failed to load users: %v", err)
		return
	}

	if len(s.users) == 0 {
		log.Printf("[simulation] No users found for simulation")
		return
	}

	// Start simulation loop
	go s.simulationLoop(ctx)
}

// Stop stops the simulation engine
func (s *SimulationEngine) Stop() {
	log.Printf("[simulation] Stopping simulation engine...")
	s.running = false
}

// IsRunning returns whether the simulation is running
func (s *SimulationEngine) IsRunning() bool {
	return s.running
}

// simulationLoop runs the main simulation loop
func (s *SimulationEngine) simulationLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	log.Printf("[simulation] Simulation loop started (checking every minute)")

	for {
		select {
		case <-ctx.Done():
			s.running = false
			return

		case <-ticker.C:
			if !s.running {
				return
			}

			s.runSimulationCycle(ctx)
		}
	}
}

// runSimulationCycle runs one cycle of simulation
func (s *SimulationEngine) runSimulationCycle(ctx context.Context) {
	currentHour := time.Now().Hour()
	log.Printf("[simulation] Running simulation cycle (hour: %d)", currentHour)

	// For each user, check if they should transact based on their persona
	for _, user := range s.users {
		persona := personas.GetPersona(user.Persona)
		if persona == nil {
			continue
		}

		// Check if user is active at this hour
		if !persona.IsActiveHour(currentHour) {
			continue
		}

		// Random chance based on frequency (simulate realistic activity)
		if !s.shouldTransact(persona.TransactionFreq) {
			continue
		}

		// Generate transaction
		if err := s.generateTransaction(ctx, user, persona); err != nil {
			log.Printf("[simulation] Failed to generate transaction for %s: %v", user.Email, err)
		}
	}
}

// shouldTransact determines if a transaction should occur based on frequency
func (s *SimulationEngine) shouldTransact(freq time.Duration) bool {
	// Convert frequency to transactions per hour
	transactionsPerHour := float64(time.Hour) / float64(freq)

	// Probability that transaction occurs in this minute
	probability := transactionsPerHour / 60.0

	return rand.Float64() < probability
}

// generateTransaction generates a single transaction based on persona
func (s *SimulationEngine) generateTransaction(ctx context.Context, user UserWallet, persona *personas.Persona) error {
	txType := persona.SelectTransactionType()
	amount := persona.RandomAmount()
	description := fmt.Sprintf("Simulated %s by %s", txType, user.Persona)

	switch txType {
	case "deposit":
		return s.gatewayClient.CreateDeposit(user.WalletID, amount, description)

	case "transfer":
		// Select random recipient
		recipient := s.selectRandomUser(user.UserID)
		if recipient == nil {
			return fmt.Errorf("no recipient available")
		}
		return s.gatewayClient.CreateTransfer(user.WalletID, recipient.WalletID, amount, description)

	case "withdrawal":
		return s.gatewayClient.CreateWithdrawal(user.WalletID, amount, description)

	default:
		return fmt.Errorf("unknown transaction type: %s", txType)
	}
}

// selectRandomUser selects a random user different from the current user
func (s *SimulationEngine) selectRandomUser(excludeUserID string) *UserWallet {
	eligible := make([]UserWallet, 0)
	for _, u := range s.users {
		if u.UserID != excludeUserID {
			eligible = append(eligible, u)
		}
	}

	if len(eligible) == 0 {
		return nil
	}

	return &eligible[rand.Intn(len(eligible))]
}
