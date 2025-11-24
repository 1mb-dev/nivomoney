// Package events provides a simple Server-Sent Events (SSE) broker for real-time event streaming.
package events

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Event represents a single event to be broadcasted.
type Event struct {
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// Client represents a connected SSE client.
type Client struct {
	ID      string
	Channel chan Event
	Topics  map[string]bool // Topics this client is subscribed to
	mu      sync.RWMutex
}

// NewClient creates a new SSE client.
func NewClient(id string) *Client {
	return &Client{
		ID:      id,
		Channel: make(chan Event, 100), // Buffer up to 100 events
		Topics:  make(map[string]bool),
	}
}

// Subscribe adds a topic to this client's subscriptions.
func (c *Client) Subscribe(topic string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Topics[topic] = true
}

// Unsubscribe removes a topic from this client's subscriptions.
func (c *Client) Unsubscribe(topic string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.Topics, topic)
}

// IsSubscribed checks if client is subscribed to a topic.
func (c *Client) IsSubscribed(topic string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Topics[topic]
}

// Broker manages SSE connections and event broadcasting.
type Broker struct {
	clients    map[string]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan BroadcastEvent
	mu         sync.RWMutex
}

// BroadcastEvent represents an event to be broadcasted to clients.
type BroadcastEvent struct {
	Topic string
	Event Event
}

// NewBroker creates a new SSE broker.
func NewBroker() *Broker {
	return &Broker{
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan BroadcastEvent, 1000), // Buffer broadcasts
	}
}

// Start starts the broker's event loop.
func (b *Broker) Start() {
	go func() {
		for {
			select {
			case client := <-b.register:
				b.mu.Lock()
				b.clients[client.ID] = client
				b.mu.Unlock()

			case client := <-b.unregister:
				b.mu.Lock()
				if _, ok := b.clients[client.ID]; ok {
					close(client.Channel)
					delete(b.clients, client.ID)
				}
				b.mu.Unlock()

			case event := <-b.broadcast:
				b.mu.RLock()
				for _, client := range b.clients {
					// Only send to clients subscribed to this topic (or "all")
					if client.IsSubscribed(event.Topic) || client.IsSubscribed("all") {
						select {
						case client.Channel <- event.Event:
						default:
							// Client's buffer is full, skip this event
						}
					}
				}
				b.mu.RUnlock()
			}
		}
	}()
}

// Register registers a new client with the broker.
func (b *Broker) Register(client *Client) {
	b.register <- client
}

// Unregister removes a client from the broker.
func (b *Broker) Unregister(client *Client) {
	b.unregister <- client
}

// Broadcast sends an event to all subscribed clients.
func (b *Broker) Broadcast(topic string, eventType string, data map[string]interface{}) {
	event := Event{
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now(),
	}

	b.broadcast <- BroadcastEvent{
		Topic: topic,
		Event: event,
	}
}

// GetClientCount returns the number of connected clients.
func (b *Broker) GetClientCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.clients)
}

// FormatSSE formats an event for Server-Sent Events protocol.
func FormatSSE(event Event) string {
	data, _ := json.Marshal(event)
	return fmt.Sprintf("event: %s\ndata: %s\n\n", event.Type, string(data))
}
