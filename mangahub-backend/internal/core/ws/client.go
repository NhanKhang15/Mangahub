package ws

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = 54 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	Hub *Hub

	// The websocket connection.
	Conn *websocket.Conn

	// User ID associated with this client.
	UserID string

	// Buffered channel of outbound messages.
	send chan interface{}
}

func NewClient(hub *Hub, conn *websocket.Conn, userID string) *Client {
	return &Client{
		Hub:    hub,
		Conn:   conn,
		UserID: userID,
		send:   make(chan interface{}, 256),
	}
}

// readPump pumps messages from the websocket connection to the hub.
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { _ = c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var baseMsg BaseMessage
		if err := json.Unmarshal(message, &baseMsg); err != nil {
			log.Printf("invalid json message: %v", err)
			continue
		}

		switch baseMsg.Type {
		case "subscribe":
			if baseMsg.Room != "" {
				c.Hub.Subscribe(c, baseMsg.Room)
			}
		case "unsubscribe":
			if baseMsg.Room != "" {
				c.Hub.Unsubscribe(c, baseMsg.Room)
			}
		case "ping":
			// Handled by the ping/pong loop, but if client sends explicit json {"type":"ping"},
			// we can respond with explicit {"type":"pong"} or just let the ws ping/pong handle it.
			// The requirements say {"type":"ping"} is from client -> server. Let's send a pong.
			c.send <- map[string]string{"type": "pong"}
		}
	}
}

// writePump pumps messages from the hub to the websocket connection.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteJSON(message); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
