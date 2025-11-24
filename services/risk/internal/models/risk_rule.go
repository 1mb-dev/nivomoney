package models

import (
	"encoding/json"
	"time"
)

// RuleType represents the type of risk rule
type RuleType string

const (
	RuleTypeVelocity   RuleType = "velocity"    // Max transactions per time window
	RuleTypeDailyLimit RuleType = "daily_limit" // Max amount per day per user
	RuleTypeThreshold  RuleType = "threshold"   // Transaction amount threshold
)

// RiskAction represents the action to take when a rule is triggered
type RiskAction string

const (
	RiskActionAllow RiskAction = "allow" // Allow the transaction
	RiskActionBlock RiskAction = "block" // Block the transaction
	RiskActionFlag  RiskAction = "flag"  // Flag for review but allow
)

// RiskRule represents a risk evaluation rule
type RiskRule struct {
	ID         string                 `json:"id" db:"id"`
	RuleType   RuleType               `json:"rule_type" db:"rule_type"`
	Name       string                 `json:"name" db:"name"`
	Parameters map[string]interface{} `json:"parameters" db:"parameters"` // JSONB parameters specific to rule type
	Action     RiskAction             `json:"action" db:"action"`         // Action to take when triggered
	Enabled    bool                   `json:"enabled" db:"enabled"`
	CreatedAt  time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at" db:"updated_at"`
}

// VelocityRuleParams represents parameters for velocity check rule
type VelocityRuleParams struct {
	MaxTransactions int  `json:"max_transactions"` // Max number of transactions
	TimeWindowMins  int  `json:"time_window_mins"` // Time window in minutes
	PerUser         bool `json:"per_user"`         // Apply per user vs globally
}

// DailyLimitParams represents parameters for daily limit rule
type DailyLimitParams struct {
	MaxAmount int64  `json:"max_amount"` // Max amount in smallest currency unit
	Currency  string `json:"currency"`   // Currency code
	PerUser   bool   `json:"per_user"`   // Apply per user vs globally
}

// ThresholdParams represents parameters for threshold rule
type ThresholdParams struct {
	MinAmount int64  `json:"min_amount"` // Minimum amount to trigger (0 = no min)
	MaxAmount int64  `json:"max_amount"` // Maximum amount to trigger
	Currency  string `json:"currency"`   // Currency code
}

// UnmarshalParameters unmarshals the parameters into a specific struct
func (r *RiskRule) UnmarshalParameters(target interface{}) error {
	// Convert map to JSON bytes
	jsonBytes, err := json.Marshal(r.Parameters)
	if err != nil {
		return err
	}

	// Unmarshal into target struct
	return json.Unmarshal(jsonBytes, target)
}

// SetParameters sets the parameters from a struct
func (r *RiskRule) SetParameters(params interface{}) error {
	// Marshal params to JSON
	jsonBytes, err := json.Marshal(params)
	if err != nil {
		return err
	}

	// Unmarshal to map
	var paramsMap map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &paramsMap); err != nil {
		return err
	}

	r.Parameters = paramsMap
	return nil
}
