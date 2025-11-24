package models

import (
	"time"
)

// RiskEvent represents a risk evaluation event for audit trail
type RiskEvent struct {
	ID            string                 `json:"id" db:"id"`
	TransactionID string                 `json:"transaction_id" db:"transaction_id"` // Related transaction ID
	UserID        string                 `json:"user_id" db:"user_id"`               // User being evaluated
	RuleID        *string                `json:"rule_id,omitempty" db:"rule_id"`     // Rule that triggered (null if no rules triggered)
	RuleType      *RuleType              `json:"rule_type,omitempty" db:"rule_type"` // Type of rule triggered
	RiskScore     int                    `json:"risk_score" db:"risk_score"`         // Risk score (0-100)
	Action        RiskAction             `json:"action" db:"action"`                 // Action taken
	Reason        string                 `json:"reason" db:"reason"`                 // Human-readable reason
	Metadata      map[string]interface{} `json:"metadata,omitempty" db:"metadata"`   // JSONB additional context
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
}

// EvaluationRequest represents a request to evaluate risk for a transaction
type EvaluationRequest struct {
	TransactionID   string `json:"transaction_id"`
	UserID          string `json:"user_id"`
	Amount          int64  `json:"amount"` // Amount in smallest currency unit
	Currency        string `json:"currency"`
	TransactionType string `json:"transaction_type"` // transfer, deposit, withdrawal
	FromWalletID    string `json:"from_wallet_id,omitempty"`
	ToWalletID      string `json:"to_wallet_id,omitempty"`
}

// EvaluationResult represents the result of a risk evaluation
type EvaluationResult struct {
	Allowed        bool       `json:"allowed"`         // Whether transaction is allowed
	Action         RiskAction `json:"action"`          // Action to take
	RiskScore      int        `json:"risk_score"`      // Risk score (0-100)
	Reason         string     `json:"reason"`          // Human-readable reason
	TriggeredRules []string   `json:"triggered_rules"` // IDs of rules that were triggered
	EventID        string     `json:"event_id"`        // ID of the risk event created
}
