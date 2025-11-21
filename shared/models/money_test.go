package models

import (
	"encoding/json"
	"testing"
)

func TestNewMoney(t *testing.T) {
	m := NewMoney(1000, USD)
	if m.Amount != 1000 {
		t.Errorf("Expected amount 1000, got %d", m.Amount)
	}
	if m.Currency != USD {
		t.Errorf("Expected currency USD, got %s", m.Currency)
	}
}

func TestNewMoneyFromFloat(t *testing.T) {
	tests := []struct {
		name     string
		amount   float64
		currency Currency
		expected int64
	}{
		{"whole number", 10.0, USD, 1000},
		{"decimal", 10.50, USD, 1050},
		{"small decimal", 0.01, USD, 1},
		{"large amount", 1234.56, EUR, 123456},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMoneyFromFloat(tt.amount, tt.currency)
			if m.Amount != tt.expected {
				t.Errorf("Expected amount %d, got %d", tt.expected, m.Amount)
			}
			if m.Currency != tt.currency {
				t.Errorf("Expected currency %s, got %s", tt.currency, m.Currency)
			}
		})
	}
}

func TestMoney_Add(t *testing.T) {
	tests := []struct {
		name        string
		m1          Money
		m2          Money
		expected    int64
		expectError bool
	}{
		{
			"same currency",
			NewMoney(1000, USD),
			NewMoney(500, USD),
			1500,
			false,
		},
		{
			"different currencies",
			NewMoney(1000, USD),
			NewMoney(500, EUR),
			0,
			true,
		},
		{
			"negative amounts",
			NewMoney(-500, USD),
			NewMoney(1000, USD),
			500,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.m1.Add(tt.m2)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result.Amount != tt.expected {
					t.Errorf("Expected %d, got %d", tt.expected, result.Amount)
				}
			}
		})
	}
}

func TestMoney_Subtract(t *testing.T) {
	tests := []struct {
		name        string
		m1          Money
		m2          Money
		expected    int64
		expectError bool
	}{
		{
			"same currency",
			NewMoney(1000, USD),
			NewMoney(500, USD),
			500,
			false,
		},
		{
			"different currencies",
			NewMoney(1000, USD),
			NewMoney(500, EUR),
			0,
			true,
		},
		{
			"result negative",
			NewMoney(500, USD),
			NewMoney(1000, USD),
			-500,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.m1.Subtract(tt.m2)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result.Amount != tt.expected {
					t.Errorf("Expected %d, got %d", tt.expected, result.Amount)
				}
			}
		})
	}
}

func TestMoney_Multiply(t *testing.T) {
	m := NewMoney(100, USD)
	result := m.Multiply(5)

	if result.Amount != 500 {
		t.Errorf("Expected 500, got %d", result.Amount)
	}
}

func TestMoney_Divide(t *testing.T) {
	tests := []struct {
		name     string
		amount   int64
		divisor  int64
		expected int64
	}{
		{"normal division", 1000, 2, 500},
		{"division by zero", 1000, 0, 1000}, // Should return original
		{"uneven division", 1000, 3, 333},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMoney(tt.amount, USD)
			result := m.Divide(tt.divisor)
			if result.Amount != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result.Amount)
			}
		})
	}
}

func TestMoney_IsZero(t *testing.T) {
	if !NewMoney(0, USD).IsZero() {
		t.Error("IsZero() should return true for zero amount")
	}
	if NewMoney(100, USD).IsZero() {
		t.Error("IsZero() should return false for non-zero amount")
	}
}

func TestMoney_IsPositive(t *testing.T) {
	if !NewMoney(100, USD).IsPositive() {
		t.Error("IsPositive() should return true for positive amount")
	}
	if NewMoney(-100, USD).IsPositive() {
		t.Error("IsPositive() should return false for negative amount")
	}
	if NewMoney(0, USD).IsPositive() {
		t.Error("IsPositive() should return false for zero amount")
	}
}

func TestMoney_IsNegative(t *testing.T) {
	if !NewMoney(-100, USD).IsNegative() {
		t.Error("IsNegative() should return true for negative amount")
	}
	if NewMoney(100, USD).IsNegative() {
		t.Error("IsNegative() should return false for positive amount")
	}
}

func TestMoney_Comparisons(t *testing.T) {
	m1 := NewMoney(1000, USD)
	m2 := NewMoney(500, USD)
	m3 := NewMoney(1000, USD)
	m4 := NewMoney(1000, EUR)

	// GreaterThan
	if !m1.GreaterThan(m2) {
		t.Error("1000 should be greater than 500")
	}
	if m2.GreaterThan(m1) {
		t.Error("500 should not be greater than 1000")
	}
	if m1.GreaterThan(m4) {
		t.Error("Different currencies should return false")
	}

	// LessThan
	if !m2.LessThan(m1) {
		t.Error("500 should be less than 1000")
	}
	if m1.LessThan(m2) {
		t.Error("1000 should not be less than 500")
	}

	// Equal
	if !m1.Equal(m3) {
		t.Error("Equal amounts and currencies should be equal")
	}
	if m1.Equal(m4) {
		t.Error("Different currencies should not be equal")
	}

	// GreaterThanOrEqual
	if !m1.GreaterThanOrEqual(m3) {
		t.Error("Equal amounts should satisfy GreaterThanOrEqual")
	}
	if !m1.GreaterThanOrEqual(m2) {
		t.Error("1000 should be >= 500")
	}

	// LessThanOrEqual
	if !m1.LessThanOrEqual(m3) {
		t.Error("Equal amounts should satisfy LessThanOrEqual")
	}
	if !m2.LessThanOrEqual(m1) {
		t.Error("500 should be <= 1000")
	}
}

func TestMoney_ToFloat(t *testing.T) {
	tests := []struct {
		amount   int64
		expected float64
	}{
		{1000, 10.00},
		{1050, 10.50},
		{1, 0.01},
		{-500, -5.00},
	}

	for _, tt := range tests {
		m := NewMoney(tt.amount, USD)
		if m.ToFloat() != tt.expected {
			t.Errorf("Expected %.2f, got %.2f", tt.expected, m.ToFloat())
		}
	}
}

func TestMoney_String(t *testing.T) {
	m := NewMoney(1050, USD)
	expected := "10.50 USD"
	if m.String() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, m.String())
	}
}

func TestMoney_Validate(t *testing.T) {
	validMoney := NewMoney(1000, USD)
	if err := validMoney.Validate(); err != nil {
		t.Errorf("Valid money should not return error: %v", err)
	}

	invalidMoney := NewMoney(1000, "XXX")
	if err := invalidMoney.Validate(); err == nil {
		t.Error("Invalid currency should return error")
	}
}

func TestMoney_JSON(t *testing.T) {
	original := NewMoney(1050, USD)

	// Marshal
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal
	var decoded Money
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Compare
	if !original.Equal(decoded) {
		t.Errorf("Original %v != Decoded %v", original, decoded)
	}
}
