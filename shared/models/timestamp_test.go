package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewTimestamp(t *testing.T) {
	now := time.Now()
	ts := NewTimestamp(now)

	if !ts.Equal(NewTimestamp(now)) {
		t.Error("NewTimestamp should create equal timestamps for same time")
	}
}

func TestNow(t *testing.T) {
	before := time.Now()
	ts := Now()
	after := time.Now()

	if ts.Before(NewTimestamp(before)) {
		t.Error("Now() should not be before the before time")
	}
	if ts.After(NewTimestamp(after)) {
		t.Error("Now() should not be after the after time")
	}
}

func TestTimestamp_JSON(t *testing.T) {
	// Test with valid timestamp
	original := NewTimestamp(time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC))

	// Marshal
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal
	var decoded Timestamp
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Compare (using time.Equal which handles timezone differences)
	if !original.Time.Equal(decoded.Time) {
		t.Errorf("Original %v != Decoded %v", original, decoded)
	}

	// Test with zero timestamp
	zero := Timestamp{}
	data, err = json.Marshal(zero)
	if err != nil {
		t.Fatalf("Failed to marshal zero timestamp: %v", err)
	}
	if string(data) != "null" {
		t.Errorf("Expected 'null' for zero timestamp, got %s", string(data))
	}

	// Test unmarshaling null
	var nullTs Timestamp
	if err := json.Unmarshal([]byte("null"), &nullTs); err != nil {
		t.Fatalf("Failed to unmarshal null: %v", err)
	}
	if !nullTs.IsZero() {
		t.Error("Unmarshaled null should be zero timestamp")
	}
}

func TestTimestamp_String(t *testing.T) {
	ts := NewTimestamp(time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC))
	expected := "2025-01-15T10:30:00Z"

	if ts.String() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, ts.String())
	}

	// Test zero timestamp
	zero := Timestamp{}
	if zero.String() != "" {
		t.Errorf("Expected empty string for zero timestamp, got '%s'", zero.String())
	}
}

func TestTimestamp_Comparisons(t *testing.T) {
	t1 := NewTimestamp(time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC))
	t2 := NewTimestamp(time.Date(2025, 1, 15, 11, 0, 0, 0, time.UTC))
	t3 := NewTimestamp(time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC))

	// Before
	if !t1.Before(t2) {
		t.Error("t1 should be before t2")
	}
	if t2.Before(t1) {
		t.Error("t2 should not be before t1")
	}

	// After
	if !t2.After(t1) {
		t.Error("t2 should be after t1")
	}
	if t1.After(t2) {
		t.Error("t1 should not be after t2")
	}

	// Equal
	if !t1.Equal(t3) {
		t.Error("t1 should equal t3")
	}
	if t1.Equal(t2) {
		t.Error("t1 should not equal t2")
	}
}

func TestTimestamp_Value(t *testing.T) {
	ts := NewTimestamp(time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC))

	value, err := ts.Value()
	if err != nil {
		t.Fatalf("Value() returned error: %v", err)
	}

	if _, ok := value.(time.Time); !ok {
		t.Error("Value() should return time.Time")
	}

	// Test zero timestamp
	zero := Timestamp{}
	value, err = zero.Value()
	if err != nil {
		t.Fatalf("Value() returned error for zero: %v", err)
	}
	if value != nil {
		t.Error("Value() should return nil for zero timestamp")
	}
}

func TestTimestamp_Scan(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		expectError bool
	}{
		{"time.Time", time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC), false},
		{"string", "2025-01-15T10:00:00Z", false},
		{"bytes", []byte("2025-01-15T10:00:00Z"), false},
		{"nil", nil, false},
		{"invalid type", 12345, true},
		{"invalid string", "invalid-date", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ts Timestamp
			err := ts.Scan(tt.input)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}
