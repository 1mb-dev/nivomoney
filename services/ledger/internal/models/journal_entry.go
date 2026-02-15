package models

import (
	"encoding/json"

	"github.com/1mb-dev/nivomoney/shared/models"
)

// EntryStatus represents the status of a journal entry.
type EntryStatus string

const (
	EntryStatusDraft    EntryStatus = "draft"    // Draft, not yet posted
	EntryStatusPosted   EntryStatus = "posted"   // Posted to ledger
	EntryStatusVoided   EntryStatus = "voided"   // Voided (reversed)
	EntryStatusReversed EntryStatus = "reversed" // Reversed by another entry
)

// EntryType represents the type of journal entry.
type EntryType string

const (
	EntryTypeStandard  EntryType = "standard"  // Standard journal entry
	EntryTypeOpening   EntryType = "opening"   // Opening balance
	EntryTypeClosing   EntryType = "closing"   // Closing entry
	EntryTypeAdjusting EntryType = "adjusting" // Adjusting entry
	EntryTypeReversing EntryType = "reversing" // Reversing entry
)

// JournalEntry represents a complete transaction with multiple line items.
// In double-entry bookkeeping, total debits must equal total credits.
type JournalEntry struct {
	ID              string            `json:"id" db:"id"`
	EntryNumber     string            `json:"entry_number" db:"entry_number"` // Sequential entry number
	Type            EntryType         `json:"type" db:"type"`
	Status          EntryStatus       `json:"status" db:"status"`
	Description     string            `json:"description" db:"description"`
	ReferenceType   string            `json:"reference_type,omitempty" db:"reference_type"` // e.g., "transaction", "invoice"
	ReferenceID     string            `json:"reference_id,omitempty" db:"reference_id"`     // ID of referenced entity
	PostedAt        *models.Timestamp `json:"posted_at,omitempty" db:"posted_at"`
	PostedBy        *string           `json:"posted_by,omitempty" db:"posted_by"` // User ID who posted
	VoidedAt        *models.Timestamp `json:"voided_at,omitempty" db:"voided_at"`
	VoidedBy        *string           `json:"voided_by,omitempty" db:"voided_by"`
	VoidReason      *string           `json:"void_reason,omitempty" db:"void_reason"`
	ReversalEntryID *string           `json:"reversal_entry_id,omitempty" db:"reversal_entry_id"` // Entry that reversed this
	Metadata        map[string]string `json:"metadata,omitempty" db:"metadata"`                   // JSONB
	CreatedAt       models.Timestamp  `json:"created_at" db:"created_at"`
	UpdatedAt       models.Timestamp  `json:"updated_at" db:"updated_at"`

	// Embedded lines (loaded separately)
	Lines []LedgerLine `json:"lines,omitempty" db:"-"`
}

// IsBalanced returns true if total debits equal total credits.
func (j *JournalEntry) IsBalanced() bool {
	var totalDebits, totalCredits int64
	for _, line := range j.Lines {
		totalDebits += line.DebitAmount
		totalCredits += line.CreditAmount
	}
	return totalDebits == totalCredits
}

// TotalDebits returns the sum of all debit amounts.
func (j *JournalEntry) TotalDebits() int64 {
	var total int64
	for _, line := range j.Lines {
		total += line.DebitAmount
	}
	return total
}

// TotalCredits returns the sum of all credit amounts.
func (j *JournalEntry) TotalCredits() int64 {
	var total int64
	for _, line := range j.Lines {
		total += line.CreditAmount
	}
	return total
}

// LedgerLine represents a single debit or credit line in a journal entry.
type LedgerLine struct {
	ID           string            `json:"id" db:"id"`
	EntryID      string            `json:"entry_id" db:"entry_id"`
	AccountID    string            `json:"account_id" db:"account_id"`
	DebitAmount  int64             `json:"debit_amount" db:"debit_amount"`   // Amount in paise (0 if credit)
	CreditAmount int64             `json:"credit_amount" db:"credit_amount"` // Amount in paise (0 if debit)
	Description  string            `json:"description,omitempty" db:"description"`
	Metadata     map[string]string `json:"metadata,omitempty" db:"metadata"` // JSONB
	CreatedAt    models.Timestamp  `json:"created_at" db:"created_at"`

	// Embedded account info (loaded separately)
	Account *Account `json:"account,omitempty" db:"-"`
}

// IsDebit returns true if this is a debit line.
func (l *LedgerLine) IsDebit() bool {
	return l.DebitAmount > 0
}

// IsCredit returns true if this is a credit line.
func (l *LedgerLine) IsCredit() bool {
	return l.CreditAmount > 0
}

// Amount returns the non-zero amount (either debit or credit).
func (l *LedgerLine) Amount() int64 {
	if l.DebitAmount > 0 {
		return l.DebitAmount
	}
	return l.CreditAmount
}

// CreateJournalEntryRequest represents a request to create a journal entry.
type CreateJournalEntryRequest struct {
	Type          EntryType         `json:"type" validate:"required"`
	Description   string            `json:"description" validate:"required,min:5,max:500"`
	ReferenceType string            `json:"reference_type,omitempty" validate:"omitempty,max:50"`
	ReferenceID   string            `json:"reference_id,omitempty" validate:"omitempty,max:100"`
	Lines         []LedgerLineInput `json:"lines" validate:"required,min:2,dive"`
	MetadataRaw   json.RawMessage   `json:"metadata,omitempty" validate:"-"` // Raw JSON, parsed via GetMetadata()
}

// GetMetadata parses and returns the metadata map.
func (r *CreateJournalEntryRequest) GetMetadata() (map[string]string, error) {
	if len(r.MetadataRaw) == 0 {
		return make(map[string]string), nil
	}

	var metadata map[string]string
	if err := json.Unmarshal(r.MetadataRaw, &metadata); err != nil {
		return nil, err
	}
	return metadata, nil
}

// LedgerLineInput represents input for creating a ledger line.
type LedgerLineInput struct {
	AccountID    string          `json:"account_id" validate:"required,uuid"`
	DebitAmount  int64           `json:"debit_amount" validate:"min:0"`
	CreditAmount int64           `json:"credit_amount" validate:"min:0"`
	Description  string          `json:"description,omitempty" validate:"omitempty,max:500"`
	MetadataRaw  json.RawMessage `json:"metadata,omitempty" validate:"-"` // Raw JSON, parsed via GetMetadata()
}

// GetMetadata parses and returns the metadata map.
func (l *LedgerLineInput) GetMetadata() (map[string]string, error) {
	if len(l.MetadataRaw) == 0 {
		return make(map[string]string), nil
	}

	var metadata map[string]string
	if err := json.Unmarshal(l.MetadataRaw, &metadata); err != nil {
		return nil, err
	}
	return metadata, nil
}

// Validate checks if the line input is valid (either debit or credit, not both).
func (l *LedgerLineInput) Validate() error {
	if l.DebitAmount > 0 && l.CreditAmount > 0 {
		return &ValidationError{
			Field:   "lines",
			Message: "line cannot have both debit and credit amounts",
		}
	}
	if l.DebitAmount == 0 && l.CreditAmount == 0 {
		return &ValidationError{
			Field:   "lines",
			Message: "line must have either debit or credit amount",
		}
	}
	return nil
}

// PostJournalEntryRequest represents a request to post a draft journal entry.
type PostJournalEntryRequest struct {
	EntryID  string `json:"entry_id" validate:"required,uuid"`
	PostedBy string `json:"posted_by" validate:"required,uuid"`
}

// VoidJournalEntryRequest represents a request to void a journal entry.
type VoidJournalEntryRequest struct {
	EntryID    string `json:"entry_id" validate:"required,uuid"`
	VoidedBy   string `json:"voided_by" validate:"required,uuid"`
	VoidReason string `json:"void_reason" validate:"required,min:10,max:500"`
}

// ValidationError represents a validation error.
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return e.Message
}
