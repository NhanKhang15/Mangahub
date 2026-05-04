package ws

import (
	"log"
	"sync"
)

// Hub maintains the set of active clients and broadcasts messages to the clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan *RoomMessage

	// Direct messages to specific users.
	direct chan *DirectMessage

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	// rooms maps a room name to a map of clients in that room.
	rooms map[string]map[*Client]bool

	// byUser maps a user ID to a client.
	byUser map[string]*Client

	mu sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan *RoomMessage),
		direct:     make(chan *DirectMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		rooms:      make(map[string]map[*Client]bool),
		byUser:     make(map[string]*Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.byUser[client.UserID] = client
			h.mu.Unlock()

			// Send connection success message
			client.send <- &SystemMessage{
				Type:    "system",
				Content: "connected",
			}

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				delete(h.byUser, client.UserID)
				close(client.send)
				
				// Remove client from all rooms
				for room, clientsInRoom := range h.rooms {
					if _, inRoom := clientsInRoom[client]; inRoom {
						delete(clientsInRoom, client)
						if len(clientsInRoom) == 0 {
							delete(h.rooms, room)
						}
					}
				}
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.Lock()
			clientsInRoom, ok := h.rooms[message.Room]
			if ok {
				for client := range clientsInRoom {
					select {
					case client.send <- message:
					default:
						close(client.send)
						delete(h.clients, client)
						delete(clientsInRoom, client)
						delete(h.byUser, client.UserID)
					}
				}
				if len(clientsInRoom) == 0 {
					delete(h.rooms, message.Room)
				}
			}
			h.mu.Unlock()

		case message := <-h.direct:
			h.mu.Lock()
			if client, ok := h.byUser[message.UserID]; ok {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
					delete(h.byUser, client.UserID)
					
					// Remove from rooms
					for room, clientsInRoom := range h.rooms {
						delete(clientsInRoom, client)
						if len(clientsInRoom) == 0 {
							delete(h.rooms, room)
						}
					}
				}
			}
			h.mu.Unlock()
		}
	}
}

// Subscribe adds a client to a room
func (h *Hub) Subscribe(client *Client, room string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.rooms[room] == nil {
		h.rooms[room] = make(map[*Client]bool)
	}
	h.rooms[room][client] = true
	log.Printf("Client %s subscribed to %s", client.UserID, room)
}

// Unsubscribe removes a client from a room
func (h *Hub) Unsubscribe(client *Client, room string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clientsInRoom, ok := h.rooms[room]; ok {
		delete(clientsInRoom, client)
		if len(clientsInRoom) == 0 {
			delete(h.rooms, room)
		}
	}
	log.Printf("Client %s unsubscribed from %s", client.UserID, room)
}

// GetActiveRooms returns a list of all currently active rooms (rooms with at least 1 subscriber)
func (h *Hub) GetActiveRooms() []string {
	h.mu.Lock()
	defer h.mu.Unlock()

	rooms := make([]string, 0, len(h.rooms))
	for r := range h.rooms {
		rooms = append(rooms, r)
	}
	return rooms
}

// Broadcast sends a message to all clients in a room
func (h *Hub) Broadcast(msg *RoomMessage) {
	h.broadcast <- msg
}

// SendDirect sends a message to a specific user
func (h *Hub) SendDirect(msg *DirectMessage) {
	h.direct <- msg
}

// Register registers a new client to the hub
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister unregisters a client from the hub
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}
