package models

import (
	"encoding/json"

	"github.com/vnykmshr/nivo/shared/models"
)

// NotificationTemplate represents a notification template with variable substitution.
type NotificationTemplate struct {
	ID              string              `json:"id" db:"id"`
	Name            string              `json:"name" db:"name"` // Unique identifier (e.g., "otp_sms", "transaction_alert_email")
	Channel         NotificationChannel `json:"channel" db:"channel"`
	SubjectTemplate string              `json:"subject_template,omitempty" db:"subject_template"` // For email/push
	BodyTemplate    string              `json:"body_template" db:"body_template"`
	Version         int                 `json:"version" db:"version"`
	Metadata        map[string]string   `json:"metadata,omitempty" db:"metadata"`
	CreatedAt       models.Timestamp    `json:"created_at" db:"created_at"`
	UpdatedAt       models.Timestamp    `json:"updated_at" db:"updated_at"`
}

// CreateTemplateRequest represents a request to create a notification template.
type CreateTemplateRequest struct {
	Name            string              `json:"name" validate:"required,min=3,max=100"`
	Channel         NotificationChannel `json:"channel" validate:"required,oneof=sms email push in_app"`
	SubjectTemplate string              `json:"subject_template,omitempty" validate:"omitempty,max=200"`
	BodyTemplate    string              `json:"body_template" validate:"required,max=5000"`
	MetadataRaw     json.RawMessage     `json:"metadata,omitempty"`
}

// GetMetadata parses and returns the metadata map.
func (r *CreateTemplateRequest) GetMetadata() (map[string]string, error) {
	if len(r.MetadataRaw) == 0 {
		return make(map[string]string), nil
	}
	var metadata map[string]string
	if err := json.Unmarshal(r.MetadataRaw, &metadata); err != nil {
		return nil, err
	}
	return metadata, nil
}

// UpdateTemplateRequest represents a request to update a notification template.
type UpdateTemplateRequest struct {
	SubjectTemplate *string         `json:"subject_template,omitempty" validate:"omitempty,max=200"`
	BodyTemplate    *string         `json:"body_template,omitempty" validate:"omitempty,max=5000"`
	MetadataRaw     json.RawMessage `json:"metadata,omitempty"`
}

// GetMetadata parses and returns the metadata map.
func (r *UpdateTemplateRequest) GetMetadata() (map[string]string, error) {
	if len(r.MetadataRaw) == 0 {
		return nil, nil
	}
	var metadata map[string]string
	if err := json.Unmarshal(r.MetadataRaw, &metadata); err != nil {
		return nil, err
	}
	return metadata, nil
}

// PreviewTemplateRequest represents a request to preview a template with variables.
type PreviewTemplateRequest struct {
	Variables map[string]interface{} `json:"variables"`
}

// PreviewTemplateResponse represents the rendered template preview.
type PreviewTemplateResponse struct {
	Subject      string           `json:"subject,omitempty"`
	Body         string           `json:"body"`
	RenderedAt   models.Timestamp `json:"rendered_at"`
	VariableUsed []string         `json:"variables_used"` // List of variables that were substituted
}
