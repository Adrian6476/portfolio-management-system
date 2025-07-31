package services

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin in development
		// In production, you should check the origin properly
		return true
	},
}

// Client represents a WebSocket client
type Client struct {
	ID            string
	Conn          *websocket.Conn
	Send          chan []byte
	Hub           *WebSocketHub
	UserID        string
	Subscriptions map[string]bool // Track what the client is subscribed to
}

// WebSocketHub manages all WebSocket connections
type WebSocketHub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	Register   chan *Client // Exported
	Unregister chan *Client // Exported
	logger     *zap.Logger
	mutex      sync.RWMutex
}

// Message types for WebSocket communication
type WSMessage struct {
	Type      string      `json:"type"`
	Symbol    string      `json:"symbol,omitempty"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

// Portfolio update message
type PortfolioUpdate struct {
	TotalValue                float64 `json:"total_value"`
	DailyChange               float64 `json:"daily_change"`
	DailyChangePercent        float64 `json:"daily_change_percent"`
	UnrealizedGainLoss        float64 `json:"unrealized_gain_loss"`
	UnrealizedGainLossPercent float64 `json:"unrealized_gain_loss_percent"`
}

// Price update message
type PriceUpdate struct {
	Symbol        string  `json:"symbol"`
	CurrentPrice  float64 `json:"current_price"`
	Change        float64 `json:"change"`
	ChangePercent float64 `json:"change_percent"`
	High          float64 `json:"high"`
	Low           float64 `json:"low"`
	Volume        int64   `json:"volume,omitempty"`
}

// NewWebSocketHub creates a new WebSocket hub
func NewWebSocketHub(logger *zap.Logger) *WebSocketHub {
	return &WebSocketHub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		logger:     logger,
	}
}

// Run starts the WebSocket hub
func (h *WebSocketHub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
			h.logger.Info("Client connected", zap.String("client_id", client.ID))

			// Send welcome message
			welcome := WSMessage{
				Type:      "connected",
				Data:      map[string]string{"status": "connected", "client_id": client.ID},
				Timestamp: time.Now().Unix(),
			}
			if data, err := json.Marshal(welcome); err == nil {
				select {
				case client.Send <- data:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}

		case client := <-h.Unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
				h.logger.Info("Client disconnected", zap.String("client_id", client.ID))
			}
			h.mutex.Unlock()

		case message := <-h.broadcast:
			h.mutex.RLock()
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
			h.mutex.RUnlock()
		}
	}
}

// BroadcastPortfolioUpdate sends portfolio updates to all connected clients
func (h *WebSocketHub) BroadcastPortfolioUpdate(update PortfolioUpdate) {
	message := WSMessage{
		Type:      "portfolio_update",
		Data:      update,
		Timestamp: time.Now().Unix(),
	}

	if data, err := json.Marshal(message); err == nil {
		select {
		case h.broadcast <- data:
		default:
			h.logger.Warn("Failed to broadcast portfolio update: channel full")
		}
	}
}

// BroadcastPriceUpdate sends price updates to subscribed clients
func (h *WebSocketHub) BroadcastPriceUpdate(update PriceUpdate) {
	message := WSMessage{
		Type:      "price_update",
		Symbol:    update.Symbol,
		Data:      update,
		Timestamp: time.Now().Unix(),
	}

	if data, err := json.Marshal(message); err == nil {
		h.mutex.RLock()
		for client := range h.clients {
			// Only send to clients subscribed to this symbol
			if client.Subscriptions[update.Symbol] || client.Subscriptions["portfolio"] {
				select {
				case client.Send <- data:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
		}
		h.mutex.RUnlock()
	}
}

// GetConnectedClients returns the number of connected clients
func (h *WebSocketHub) GetConnectedClients() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return len(h.clients)
}

// ReadPump handles incoming messages from the client (exported)
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle subscription messages
		var msg WSMessage
		if err := json.Unmarshal(message, &msg); err == nil {
			switch msg.Type {
			case "subscribe":
				if symbol, ok := msg.Data.(string); ok {
					c.Subscriptions[symbol] = true
					c.Hub.logger.Info("Client subscribed",
						zap.String("client_id", c.ID),
						zap.String("symbol", symbol))
				}
			case "unsubscribe":
				if symbol, ok := msg.Data.(string); ok {
					delete(c.Subscriptions, symbol)
					c.Hub.logger.Info("Client unsubscribed",
						zap.String("client_id", c.ID),
						zap.String("symbol", symbol))
				}
			}
		}
	}
}

// WritePump handles outgoing messages to the client (exported)
func (c *Client) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
