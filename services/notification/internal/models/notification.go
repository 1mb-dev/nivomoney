package models

import (
	"encoding/json"

	"github.com/1mb-dev/nivomoney/shared/models"
)

// NotificationChannel represents the delivery channel for a notification.
type NotificationChannel string

const (
	ChannelSMS   NotificationChannel = "sms"    // SMS text message
	ChannelEmail NotificationChannel = "email"  // Email message
	ChannelPush  NotificationChannel = "push"   // Push notification
	ChannelInApp NotificationChannel = "in_app" // In-app notification
)

// NotificationType represents the type of notification.
type NotificationType string

const (
	TypeOTP              NotificationType = "otp"               // One-time password
	TypeTransactionAlert NotificationType = "transaction_alert" // Transaction notification
	TypeAccountAlert     NotificationType = "account_alert"     // Account-related alert
	TypeKYCUpdate        NotificationType = "kyc_update"        // KYC status update
	TypeWelcome          NotificationType = "welcome"           // Welcome message
	TypeSecurityAlert    NotificationType = "security_alert"    // Security-related alert
	TypeWalletAlert      NotificationType = "wallet_alert"      // Wallet-related alert
	TypeSystemAlert      NotificationType = "system_alert"      // System notification
)

// NotificationStatus represents the delivery status of a notification.
type NotificationStatus string

const (
	StatusQueued    NotificationStatus = "queued"    // Queued for delivery
	StatusSent      NotificationStatus = "sent"      // Sent to provider
	StatusDelivered NotificationStatus = "delivered" // Successfully delivered
	StatusFailed    NotificationStatus = "failed"    // Delivery failed
)

// NotificationPriority represents the priority level of a notification.
type NotificationPriority string

const (
	PriorityCritical NotificationPriority = "critical" // OTP, security alerts
	PriorityHigh     NotificationPriority = "high"     // Transaction alerts
	PriorityNormal   NotificationPriority = "normal"   // Regular notifications
	PriorityLow      NotificationPriority = "low"      // Marketing, updates
)

// Notification represents a notification in the system.
type Notification struct {
	ID            string                 `json:"id" db:"id"`
	UserID        *string                `json:"user_id,omitempty" db:"user_id"` // Null for system-wide notifications
	Channel       NotificationChannel    `json:"channel" db:"channel"`
	Type          NotificationType       `json:"type" db:"type"`
	Priority      NotificationPriority   `json:"priority" db:"priority"`
	Recipient     string                 `json:"recipient" db:"recipient"`       // Email address or phone number
	Subject       string                 `json:"subject,omitempty" db:"subject"` // For email/push
	Body          string                 `json:"body" db:"body"`
	TemplateID    *string                `json:"template_id,omitempty" db:"template_id"`
	Status        NotificationStatus     `json:"status" db:"status"`
	CorrelationID *string                `json:"correlation_id,omitempty" db:"correlation_id"` // For idempotency
	SourceService string                 `json:"source_service" db:"source_service"`
	Metadata      map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	RetryCount    int                    `json:"retry_count" db:"retry_count"`
	FailureReason *string                `json:"failure_reason,omitempty" db:"failure_reason"`
	QueuedAt      models.Timestamp       `json:"queued_at" db:"queued_at"`
	SentAt        *models.Timestamp      `json:"sent_at,omitempty" db:"sent_at"`
	DeliveredAt   *models.Timestamp      `json:"delivered_at,omitempty" db:"delivered_at"`
	FailedAt      *models.Timestamp      `json:"failed_at,omitempty" db:"failed_at"`
	CreatedAt     models.Timestamp       `json:"created_at" db:"created_at"`
	UpdatedAt     models.Timestamp       `json:"updated_at" db:"updated_at"`
}

// IsQueued returns true if the notification is queued.
func (n *Notification) IsQueued() bool {
	return n.Status == StatusQueued
}

// IsSent returns true if the notification has been sent.
func (n *Notification) IsSent() bool {
	return n.Status == StatusSent
}

// IsDelivered returns true if the notification was delivered.
func (n *Notification) IsDelivered() bool {
	return n.Status == StatusDelivered
}

// IsFailed returns true if the notification failed.
func (n *Notification) IsFailed() bool {
	return n.Status == StatusFailed
}

// IsCritical returns true if the notification is critical priority.
func (n *Notification) IsCritical() bool {
	return n.Priority == PriorityCritical
}

// SendNotificationRequest represents a request to send a notification.
type SendNotificationRequest struct {
	UserID        *string                `json:"user_id,omitempty" validate:"omitempty,uuid"`
	Channel       NotificationChannel    `json:"channel" validate:"required,oneof=sms email push in_app"`
	Type          NotificationType       `json:"type" validate:"required"`
	Priority      NotificationPriority   `json:"priority,omitempty" validate:"omitempty,oneof=critical high normal low"`
	Recipient     string                 `json:"recipient" validate:"required"`
	TemplateID    *string                `json:"template_id,omitempty" validate:"omitempty,uuid"`
	Subject       string                 `json:"subject,omitempty" validate:"omitempty,max=200"`
	Body          string                 `json:"body,omitempty" validate:"omitempty,max=5000"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	CorrelationID *string                `json:"correlation_id,omitempty" validate:"omitempty,max=100"`
	MetadataRaw   json.RawMessage        `json:"metadata,omitempty"`
}

// GetMetadata parses and returns the metadata map.
func (r *SendNotificationRequest) GetMetadata() (map[string]interface{}, error) {
	if len(r.MetadataRaw) == 0 {
		return make(map[string]interface{}), nil
	}
	var metadata map[string]interface{}
	if err := json.Unmarshal(r.MetadataRaw, &metadata); err != nil {
		return nil, err
	}
	return metadata, nil
}

// SendNotificationResponse represents the response after sending a notification.
type SendNotificationResponse struct {
	NotificationID string             `json:"notification_id"`
	Status         NotificationStatus `json:"status"`
	QueuedAt       models.Timestamp   `json:"queued_at"`
}

// ListNotificationsRequest represents a request to list notifications with filters.
type ListNotificationsRequest struct {
	UserID        *string              `json:"user_id,omitempty" validate:"omitempty,uuid"`
	Channel       *NotificationChannel `json:"channel,omitempty"`
	Type          *NotificationType    `json:"type,omitempty"`
	Status        *NotificationStatus  `json:"status,omitempty"`
	SourceService *string              `json:"source_service,omitempty"`
	StartDate     *models.Timestamp    `json:"start_date,omitempty"`
	EndDate       *models.Timestamp    `json:"end_date,omitempty"`
	Limit         int                  `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
	Offset        int                  `json:"offset,omitempty" validate:"omitempty,min=0"`
}

// ListNotificationsResponse represents the response for listing notifications.
type ListNotificationsResponse struct {
	Notifications []*Notification `json:"notifications"`
	Total         int64           `json:"total"`
	Limit         int             `json:"limit"`
	Offset        int             `json:"offset"`
}

// NotificationStats represents statistics for notifications.
type NotificationStats struct {
	TotalNotifications int64                       `json:"total_notifications"`
	ByChannel          map[NotificationChannel]int `json:"by_channel"`
	ByStatus           map[NotificationStatus]int  `json:"by_status"`
	ByType             map[NotificationType]int    `json:"by_type"`
	SuccessRate        float64                     `json:"success_rate"` // Percentage
	AverageRetries     float64                     `json:"average_retries"`
}
