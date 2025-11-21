// Package models provides common domain types for Nivo services.
package models

import (
	"encoding/json"
	"fmt"
)

// Money represents a monetary amount in the smallest currency unit (e.g., cents).
// Using int64 avoids floating-point precision issues.
type Money struct {
	Amount   int64    `json:"amount"`   // Amount in smallest unit (cents)
	Currency Currency `json:"currency"` // Currency code
}

// NewMoney creates a new Money instance.
func NewMoney(amount int64, currency Currency) Money {
	return Money{
		Amount:   amount,
		Currency: currency,
	}
}

// NewMoneyFromFloat creates Money from a float value (e.g., 10.50 USD).
// The float is multiplied by 100 to convert to cents.
func NewMoneyFromFloat(amount float64, currency Currency) Money {
	return Money{
		Amount:   int64(amount * 100),
		Currency: currency,
	}
}

// Add adds two Money values. Returns error if currencies don't match.
func (m Money) Add(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, fmt.Errorf("cannot add different currencies: %s and %s", m.Currency, other.Currency)
	}
	return Money{
		Amount:   m.Amount + other.Amount,
		Currency: m.Currency,
	}, nil
}

// Subtract subtracts two Money values. Returns error if currencies don't match.
func (m Money) Subtract(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, fmt.Errorf("cannot subtract different currencies: %s and %s", m.Currency, other.Currency)
	}
	return Money{
		Amount:   m.Amount - other.Amount,
		Currency: m.Currency,
	}, nil
}

// Multiply multiplies Money by an integer factor.
func (m Money) Multiply(factor int64) Money {
	return Money{
		Amount:   m.Amount * factor,
		Currency: m.Currency,
	}
}

// Divide divides Money by an integer divisor.
func (m Money) Divide(divisor int64) Money {
	if divisor == 0 {
		return m
	}
	return Money{
		Amount:   m.Amount / divisor,
		Currency: m.Currency,
	}
}

// IsZero returns true if the amount is zero.
func (m Money) IsZero() bool {
	return m.Amount == 0
}

// IsPositive returns true if the amount is positive.
func (m Money) IsPositive() bool {
	return m.Amount > 0
}

// IsNegative returns true if the amount is negative.
func (m Money) IsNegative() bool {
	return m.Amount < 0
}

// GreaterThan returns true if m > other. Returns false if currencies don't match.
func (m Money) GreaterThan(other Money) bool {
	if m.Currency != other.Currency {
		return false
	}
	return m.Amount > other.Amount
}

// GreaterThanOrEqual returns true if m >= other. Returns false if currencies don't match.
func (m Money) GreaterThanOrEqual(other Money) bool {
	if m.Currency != other.Currency {
		return false
	}
	return m.Amount >= other.Amount
}

// LessThan returns true if m < other. Returns false if currencies don't match.
func (m Money) LessThan(other Money) bool {
	if m.Currency != other.Currency {
		return false
	}
	return m.Amount < other.Amount
}

// LessThanOrEqual returns true if m <= other. Returns false if currencies don't match.
func (m Money) LessThanOrEqual(other Money) bool {
	if m.Currency != other.Currency {
		return false
	}
	return m.Amount <= other.Amount
}

// Equal returns true if both amount and currency are equal.
func (m Money) Equal(other Money) bool {
	return m.Amount == other.Amount && m.Currency == other.Currency
}

// ToFloat converts Money to float64 representation (e.g., 1050 cents = 10.50).
func (m Money) ToFloat() float64 {
	return float64(m.Amount) / 100.0
}

// String returns a human-readable string representation.
func (m Money) String() string {
	return fmt.Sprintf("%.2f %s", m.ToFloat(), m.Currency)
}

// Validate checks if the Money value is valid.
func (m Money) Validate() error {
	if err := m.Currency.Validate(); err != nil {
		return fmt.Errorf("invalid currency: %w", err)
	}
	return nil
}

// MarshalJSON implements json.Marshaler.
func (m Money) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Amount   int64    `json:"amount"`
		Currency Currency `json:"currency"`
	}{
		Amount:   m.Amount,
		Currency: m.Currency,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (m *Money) UnmarshalJSON(data []byte) error {
	var v struct {
		Amount   int64    `json:"amount"`
		Currency Currency `json:"currency"`
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	m.Amount = v.Amount
	m.Currency = v.Currency
	return nil
}
