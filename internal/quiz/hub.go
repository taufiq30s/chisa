package quiz

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/taufiq30s/chisa/utils"
)

// Client represents a single connected WebSocket client.
type Client struct {
	conn      *websocket.Conn
	sessionID string
	send      chan []byte
}

// ClientRegistration carries a new client and its session association.
type ClientRegistration struct {
	client    *Client
	sessionID string
}

// BroadcastMessage targets a specific session with a JSON payload.
type BroadcastMessage struct {
	sessionID string
	payload   []byte
}

// wsMessage is the wire format sent to all WS clients on a queue update.
type wsMessage struct {
	Event string        `json:"event"`
	Data  []queueRecord `json:"data"`
}

type queueRecord struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	Avatar    string `json:"avatar"`
	Timestamp string `json:"timestamp"`
	Position  int    `json:"position"`
}

// Hub manages all connected WebSocket clients, grouped by session ID.
// All map mutations happen exclusively inside the Run() goroutine.
type Hub struct {
	sessions   map[string]map[*Client]bool
	broadcast  chan BroadcastMessage
	register   chan *ClientRegistration
	unregister chan *Client
	mu         sync.RWMutex
}

// NewHub creates a Hub ready to be started with Run().
func NewHub() *Hub {
	return &Hub{
		sessions:   make(map[string]map[*Client]bool),
		broadcast:  make(chan BroadcastMessage, 64),
		register:   make(chan *ClientRegistration, 16),
		unregister: make(chan *Client, 16),
	}
}

// Run processes register/unregister/broadcast events sequentially.
// Must be called in its own goroutine.
func (h *Hub) Run() {
	for {
		select {
		case reg := <-h.register:
			clients := h.sessions[reg.sessionID]
			if clients == nil {
				clients = make(map[*Client]bool)
				h.sessions[reg.sessionID] = clients
			}
			clients[reg.client] = true

		case client := <-h.unregister:
			if clients, ok := h.sessions[client.sessionID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.send)
				}
				if len(clients) == 0 {
					delete(h.sessions, client.sessionID)
				}
			}

		case msg := <-h.broadcast:
			if clients, ok := h.sessions[msg.sessionID]; ok {
				for client := range clients {
					select {
					case client.send <- msg.payload:
					default:
						// Slow client — drop and unregister
						delete(clients, client)
						close(client.send)
					}
				}
			}
		}
	}
}

// BroadcastQueueUpdate serialises queue entries and fans out to all clients
// subscribed to sessionID.
func (h *Hub) BroadcastQueueUpdate(sessionID string, entries []QueueEntry) {
	records := make([]queueRecord, len(entries))
	for i, e := range entries {
		records[i] = queueRecord{
			UserID:    e.UserID,
			Username:  e.Username,
			Avatar:    e.Avatar,
			Timestamp: e.Timestamp.UTC().Format("2006-01-02T15:04:05Z"),
			Position:  i + 1,
		}
	}
	msg := wsMessage{Event: "QUEUE_UPDATE", Data: records}
	payload, err := json.Marshal(msg)
	if err != nil {
		utils.ErrorLog.Printf("Quiz hub: failed to marshal queue update: %v\n", err)
		return
	}
	h.broadcast <- BroadcastMessage{sessionID: sessionID, payload: payload}
}

// CloseSession disconnects all clients subscribed to sessionID with the
// given WebSocket close code.
func (h *Hub) CloseSession(sessionID string, code int, reason string) {
	h.mu.RLock()
	clients := h.sessions[sessionID]
	h.mu.RUnlock()

	closeMsg := websocket.FormatCloseMessage(code, reason)
	for client := range clients {
		client.conn.WriteMessage(websocket.CloseMessage, closeMsg)
		client.conn.Close()
		h.unregister <- client
	}
}

// writePump pumps messages from the send channel to the WebSocket connection.
func (c *Client) writePump(hub *Hub) {
	defer func() {
		hub.unregister <- c
		c.conn.Close()
	}()
	for msg := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}
