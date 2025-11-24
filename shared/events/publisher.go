package events

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// Publisher publishes events to the Gateway's SSE broker.
type Publisher struct {
	gatewayURL  string
	httpClient  *http.Client
	serviceName string
}

// PublishConfig configures the event publisher.
type PublishConfig struct {
	GatewayURL  string
	ServiceName string
	Timeout     time.Duration
}

// NewPublisher creates a new event publisher.
func NewPublisher(config PublishConfig) *Publisher {
	// Default gateway URL from environment or use provided
	gatewayURL := config.GatewayURL
	if gatewayURL == "" {
		gatewayURL = os.Getenv("GATEWAY_URL")
	}
	if gatewayURL == "" {
		gatewayURL = "http://gateway:8000"
	}

	// Default timeout
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	return &Publisher{
		gatewayURL:  gatewayURL,
		serviceName: config.ServiceName,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// BroadcastPayload represents the JSON payload for broadcasting events.
type BroadcastPayload struct {
	Topic string                 `json:"topic"`
	Type  string                 `json:"type"`
	Data  map[string]interface{} `json:"data"`
}

// PublishEvent publishes an event to the SSE broker via the Gateway.
// Topic determines which subscribers receive the event.
// EventType is the event name (e.g., "transaction.created").
// Data contains the event payload.
func (p *Publisher) PublishEvent(topic, eventType string, data map[string]interface{}) error {
	// Add metadata
	if data == nil {
		data = make(map[string]interface{})
	}
	data["service"] = p.serviceName
	data["published_at"] = time.Now().UTC().Format(time.RFC3339)

	// Prepare payload
	payload := BroadcastPayload{
		Topic: topic,
		Type:  eventType,
		Data:  data,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	// Prepare request
	url := fmt.Sprintf("%s/api/v1/events/broadcast", p.gatewayURL)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to publish event: status %d", resp.StatusCode)
	}

	return nil
}

// PublishEventAsync publishes an event asynchronously (fire and forget).
// Errors are logged but not returned.
func (p *Publisher) PublishEventAsync(topic, eventType string, data map[string]interface{}) {
	go func() {
		if err := p.PublishEvent(topic, eventType, data); err != nil {
			// In production, use proper logging
			fmt.Printf("Failed to publish event %s/%s: %v\n", topic, eventType, err)
		}
	}()
}

// PublishTransactionEvent publishes a transaction-related event.
func (p *Publisher) PublishTransactionEvent(eventType string, transactionID string, data map[string]interface{}) {
	if data == nil {
		data = make(map[string]interface{})
	}
	data["transaction_id"] = transactionID
	p.PublishEventAsync("transactions", eventType, data)
}

// PublishWalletEvent publishes a wallet-related event.
func (p *Publisher) PublishWalletEvent(eventType string, walletID string, data map[string]interface{}) {
	if data == nil {
		data = make(map[string]interface{})
	}
	data["wallet_id"] = walletID
	p.PublishEventAsync("wallets", eventType, data)
}

// PublishUserEvent publishes a user-related event.
func (p *Publisher) PublishUserEvent(eventType string, userID string, data map[string]interface{}) {
	if data == nil {
		data = make(map[string]interface{})
	}
	data["user_id"] = userID
	p.PublishEventAsync("users", eventType, data)
}

// PublishRiskEvent publishes a risk-related event.
func (p *Publisher) PublishRiskEvent(eventType string, data map[string]interface{}) {
	p.PublishEventAsync("risk", eventType, data)
}
