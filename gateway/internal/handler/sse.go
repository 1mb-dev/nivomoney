package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/vnykmshr/nivo/shared/events"
	"github.com/vnykmshr/nivo/shared/logger"
)

// SSEHandler handles Server-Sent Events connections.
type SSEHandler struct {
	broker *events.Broker
	logger *logger.Logger
}

// NewSSEHandler creates a new SSE handler.
func NewSSEHandler(broker *events.Broker, log *logger.Logger) *SSEHandler {
	return &SSEHandler{
		broker: broker,
		logger: log,
	}
}

// HandleEvents handles SSE connections from clients.
func (h *SSEHandler) HandleEvents(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Get request ID for logging
	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = "unknown"
	}

	// Create a new client
	clientID := fmt.Sprintf("%s-%d", requestID, time.Now().UnixNano())
	client := events.NewClient(clientID)

	// Get topics from query parameters (comma-separated)
	topics := r.URL.Query().Get("topics")
	if topics == "" {
		topics = "all" // Subscribe to all topics by default
	}

	// Subscribe to requested topics
	client.Subscribe(topics)

	h.logger.WithField("client_id", clientID).
		WithField("topics", topics).
		Info("SSE client connected")

	// Register client with broker
	h.broker.Register(client)

	// Make sure we unregister when done
	defer func() {
		h.broker.Unregister(client)
		h.logger.WithField("client_id", clientID).Info("SSE client disconnected")
	}()

	// Get the flusher interface
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Send initial connection message
	initialEvent := events.Event{
		Type: "connected",
		Data: map[string]interface{}{
			"client_id": clientID,
			"topics":    topics,
			"message":   "Connected to Nivo event stream",
		},
		Timestamp: time.Now(),
	}
	_, _ = fmt.Fprint(w, events.FormatSSE(initialEvent))
	flusher.Flush()

	// Send periodic heartbeat
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Listen for events
	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			// Client disconnected
			return

		case event := <-client.Channel:
			// Send event to client
			_, _ = fmt.Fprint(w, events.FormatSSE(event))
			flusher.Flush()

		case <-ticker.C:
			// Send heartbeat
			heartbeat := events.Event{
				Type: "heartbeat",
				Data: map[string]interface{}{
					"timestamp": time.Now().Unix(),
				},
				Timestamp: time.Now(),
			}
			_, _ = fmt.Fprint(w, events.FormatSSE(heartbeat))
			flusher.Flush()
		}
	}
}

// BroadcastRequest represents the request payload for broadcasting events.
type BroadcastRequest struct {
	Topic string                 `json:"topic"`
	Type  string                 `json:"type"`
	Data  map[string]interface{} `json:"data"`
}

// HandleBroadcast handles broadcast requests from services.
func (h *SSEHandler) HandleBroadcast(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req BroadcastRequest

	// Try to read JSON body first
	if r.Header.Get("Content-Type") == "application/json" {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}
	} else {
		// Fallback to query parameters for backward compatibility
		req.Topic = r.URL.Query().Get("topic")
		req.Type = r.URL.Query().Get("type")
		message := r.URL.Query().Get("message")

		if message != "" {
			req.Data = map[string]interface{}{
				"message": message,
			}
		}
	}

	// Apply defaults
	if req.Topic == "" {
		req.Topic = "all"
	}
	if req.Type == "" {
		req.Type = "notification"
	}
	if req.Data == nil {
		req.Data = make(map[string]interface{})
	}

	// Add timestamp if not present
	if _, ok := req.Data["timestamp"]; !ok {
		req.Data["timestamp"] = time.Now().Format(time.RFC3339)
	}

	// Broadcast the event
	h.broker.Broadcast(req.Topic, req.Type, req.Data)

	h.logger.WithField("topic", req.Topic).
		WithField("type", req.Type).
		Info("Broadcast event sent")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, `{"success":true,"message":"Event broadcasted","clients":%d}`, h.broker.GetClientCount())
}

// HandleStats returns SSE statistics.
func (h *SSEHandler) HandleStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, `{"connected_clients":%d,"status":"healthy"}`, h.broker.GetClientCount())
}
