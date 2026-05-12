package server

import (
	"log"
	"sync"
)

// Hub keeps track of every connected TCP client and routes ProgressUpdate
// events to the subscribers of the matching user_id. The hub itself is
// transport-agnostic — Client wraps the actual net.Conn.
type Hub struct {
	mu        sync.Mutex
	byUser    map[string]map[*Client]struct{} // user_id -> set of subscribed clients
	allClients map[*Client]struct{}

	register   chan *Client
	unregister chan *Client
	broadcast  chan ProgressUpdate
	stop       chan struct{}
}

func NewHub() *Hub {
	return &Hub{
		byUser:     make(map[string]map[*Client]struct{}),
		allClients: make(map[*Client]struct{}),
		register:   make(chan *Client, 32),
		unregister: make(chan *Client, 32),
		broadcast:  make(chan ProgressUpdate, 256),
		stop:       make(chan struct{}),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case <-h.stop:
			return
		case c := <-h.register:
			h.mu.Lock()
			h.allClients[c] = struct{}{}
			h.mu.Unlock()
			log.Printf("tcpsync: client connected from %s", c.RemoteAddr())
		case c := <-h.unregister:
			h.removeClient(c)
		case update := <-h.broadcast:
			h.dispatch(update)
		}
	}
}

func (h *Hub) Stop() { close(h.stop) }

// Register adds a newly-accepted client. Until the client sends a
// SubscribeMessage, it will not receive any broadcasts.
func (h *Hub) Register(c *Client) { h.register <- c }

// Unregister removes a client (e.g. on disconnect).
func (h *Hub) Unregister(c *Client) { h.unregister <- c }

// Subscribe links a client to a specific user_id so it starts receiving
// updates for that user.
func (h *Hub) Subscribe(c *Client, userID string) {
	if userID == "" {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	c.userID = userID
	if h.byUser[userID] == nil {
		h.byUser[userID] = make(map[*Client]struct{})
	}
	h.byUser[userID][c] = struct{}{}
	log.Printf("tcpsync: %s subscribed to user=%s", c.RemoteAddr(), userID)
}

// Publish queues a progress update for fan-out. Returns false if the broadcast
// channel is saturated, so the caller can decide whether to retry or drop.
func (h *Hub) Publish(u ProgressUpdate) bool {
	select {
	case h.broadcast <- u:
		return true
	default:
		return false
	}
}

// Subscribers returns the current subscriber count for a user_id. Useful for
// publish responses so the HTTP API can answer "delivered to N clients".
func (h *Hub) Subscribers(userID string) int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.byUser[userID])
}

// TotalClients returns the total number of connected clients regardless of
// subscription state.
func (h *Hub) TotalClients() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.allClients)
}

func (h *Hub) dispatch(u ProgressUpdate) {
	h.mu.Lock()
	subs := h.byUser[u.UserID]
	targets := make([]*Client, 0, len(subs))
	for c := range subs {
		targets = append(targets, c)
	}
	h.mu.Unlock()

	for _, c := range targets {
		if err := c.Send(u); err != nil {
			log.Printf("tcpsync: send to %s failed: %v", c.RemoteAddr(), err)
			h.Unregister(c)
		}
	}
}

func (h *Hub) removeClient(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.allClients[c]; !ok {
		return
	}
	delete(h.allClients, c)
	if c.userID != "" {
		if set, ok := h.byUser[c.userID]; ok {
			delete(set, c)
			if len(set) == 0 {
				delete(h.byUser, c.userID)
			}
		}
	}
	_ = c.Close()
	log.Printf("tcpsync: client %s disconnected", c.RemoteAddr())
}
