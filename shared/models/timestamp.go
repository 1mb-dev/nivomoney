package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// Timestamp wraps time.Time with custom JSON marshaling for ISO 8601 format.
type Timestamp struct {
	time.Time
}

// NewTimestamp creates a new Timestamp from time.Time.
func NewTimestamp(t time.Time) Timestamp {
	return Timestamp{Time: t}
}

// Now returns a Timestamp for the current time.
func Now() Timestamp {
	return Timestamp{Time: time.Now().UTC()}
}

// MarshalJSON implements json.Marshaler.
// Outputs ISO 8601 format: 2006-01-02T15:04:05Z07:00
func (t Timestamp) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(t.Format(time.RFC3339))
}

// UnmarshalJSON implements json.Unmarshaler.
func (t *Timestamp) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		t.Time = time.Time{}
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	parsed, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}

	t.Time = parsed
	return nil
}

// Value implements driver.Valuer for database storage.
func (t Timestamp) Value() (driver.Value, error) {
	if t.IsZero() {
		return nil, nil
	}
	return t.Time, nil
}

// Scan implements sql.Scanner for database retrieval.
func (t *Timestamp) Scan(value interface{}) error {
	if value == nil {
		t.Time = time.Time{}
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		t.Time = v
		return nil
	case []byte:
		parsed, err := time.Parse(time.RFC3339, string(v))
		if err != nil {
			return err
		}
		t.Time = parsed
		return nil
	case string:
		parsed, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return err
		}
		t.Time = parsed
		return nil
	default:
		return fmt.Errorf("cannot scan type %T into Timestamp", value)
	}
}

// String returns the ISO 8601 formatted string.
func (t Timestamp) String() string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

// Before reports whether t is before u.
func (t Timestamp) Before(u Timestamp) bool {
	return t.Time.Before(u.Time)
}

// After reports whether t is after u.
func (t Timestamp) After(u Timestamp) bool {
	return t.Time.After(u.Time)
}

// Equal reports whether t and u represent the same time instant.
func (t Timestamp) Equal(u Timestamp) bool {
	return t.Time.Equal(u.Time)
}
