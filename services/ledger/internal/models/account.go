package models

import (
	"encoding/json"

	"github.com/vnykmshr/nivo/shared/models"
)

// AccountType represents the type of ledger account.
type AccountType string

const (
	AccountTypeAsset     AccountType = "asset"     // Assets: Debit increases, Credit decreases
	AccountTypeLiability AccountType = "liability" // Liabilities: Credit increases, Debit decreases
	AccountTypeEquity    AccountType = "equity"    // Equity: Credit increases, Debit decreases
	AccountTypeRevenue   AccountType = "revenue"   // Revenue: Credit increases, Debit decreases
	AccountTypeExpense   AccountType = "expense"   // Expenses: Debit increases, Credit decreases
)

// AccountStatus represents the status of a ledger account.
type AccountStatus string

const (
	AccountStatusActive   AccountStatus = "active"
	AccountStatusInactive AccountStatus = "inactive"
	AccountStatusClosed   AccountStatus = "closed"
)

// Account represents a ledger account in the chart of accounts.
type Account struct {
	ID          string            `json:"id" db:"id"`
	Code        string            `json:"code" db:"code"`                     // Account code (e.g., "1000" for Cash)
	Name        string            `json:"name" db:"name"`                     // Account name (e.g., "Cash in Hand")
	Type        AccountType       `json:"type" db:"type"`                     // Account type
	Currency    models.Currency   `json:"currency" db:"currency"`             // Account currency (default: INR)
	ParentID    *string           `json:"parent_id,omitempty" db:"parent_id"` // Parent account for hierarchical structure
	Balance     int64             `json:"balance" db:"balance"`               // Current balance in smallest unit (paise)
	DebitTotal  int64             `json:"debit_total" db:"debit_total"`       // Lifetime debit total
	CreditTotal int64             `json:"credit_total" db:"credit_total"`     // Lifetime credit total
	Status      AccountStatus     `json:"status" db:"status"`
	Metadata    map[string]string `json:"metadata,omitempty" db:"metadata"` // Additional metadata (JSONB)
	CreatedAt   models.Timestamp  `json:"created_at" db:"created_at"`
	UpdatedAt   models.Timestamp  `json:"updated_at" db:"updated_at"`
}

// IsDebitNormal returns true if this account type increases with debits.
func (a *Account) IsDebitNormal() bool {
	return a.Type == AccountTypeAsset || a.Type == AccountTypeExpense
}

// IsCreditNormal returns true if this account type increases with credits.
func (a *Account) IsCreditNormal() bool {
	return a.Type == AccountTypeLiability || a.Type == AccountTypeEquity || a.Type == AccountTypeRevenue
}

// CreateAccountRequest represents a request to create a new account.
type CreateAccountRequest struct {
	Code        string          `json:"code" validate:"required,min:1,max:20"`
	Name        string          `json:"name" validate:"required,min:2,max:200"`
	Type        AccountType     `json:"type" validate:"required"`
	Currency    models.Currency `json:"currency" validate:"required,len:3"`
	ParentID    *string         `json:"parent_id,omitempty" validate:"omitempty,uuid"`
	MetadataRaw json.RawMessage `json:"metadata,omitempty" validate:"-"` // Raw JSON, parsed via GetMetadata()
}

// GetMetadata parses and returns the metadata map.
func (r *CreateAccountRequest) GetMetadata() (map[string]string, error) {
	if len(r.MetadataRaw) == 0 {
		return make(map[string]string), nil
	}

	var metadata map[string]string
	if err := json.Unmarshal(r.MetadataRaw, &metadata); err != nil {
		return nil, err
	}
	return metadata, nil
}

// UpdateAccountRequest represents a request to update an account.
type UpdateAccountRequest struct {
	Name        string          `json:"name" validate:"required,min:2,max:200"`
	Status      AccountStatus   `json:"status" validate:"required"`
	MetadataRaw json.RawMessage `json:"metadata,omitempty" validate:"-"` // Raw JSON, parsed via GetMetadata()
}

// GetMetadata parses and returns the metadata map.
func (r *UpdateAccountRequest) GetMetadata() (map[string]string, error) {
	if len(r.MetadataRaw) == 0 {
		return make(map[string]string), nil
	}

	var metadata map[string]string
	if err := json.Unmarshal(r.MetadataRaw, &metadata); err != nil {
		return nil, err
	}
	return metadata, nil
}
