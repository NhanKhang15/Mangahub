package server

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

const (
	readDeadline  = 5 * time.Minute
	writeDeadline = 10 * time.Second
)

// Client wraps a single accepted TCP connection. ReadLoop parses incoming
// JSON control messages (currently just "subscribe"); WriteLoop is driven via
// the unbuffered send method called by the hub.
type Client struct {
	hub    *Hub
	conn   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer

	mu     sync.Mutex
	closed bool
	userID string // set on subscribe
}

func NewClient(hub *Hub, conn net.Conn) *Client {
	return &Client{
		hub:    hub,
		conn:   conn,
		reader: bufio.NewReader(conn),
		writer: bufio.NewWriter(conn),
	}
}

func (c *Client) RemoteAddr() string { return c.conn.RemoteAddr().String() }

// Run blocks until the connection terminates. Should be launched in its own
// goroutine. Auto-unregisters from the hub on exit.
func (c *Client) Run() {
	c.hub.Register(c)
	defer c.hub.Unregister(c)

	if err := c.sendRaw(AckMessage{Type: "system", Content: "connected"}); err != nil {
		return
	}

	for {
		_ = c.conn.SetReadDeadline(time.Now().Add(readDeadline))
		line, err := c.reader.ReadBytes('\n')
		if err != nil {
			if !errors.Is(err, io.EOF) {
				log.Printf("tcpsync: read %s: %v", c.RemoteAddr(), err)
			}
			return
		}
		if len(line) == 0 {
			continue
		}

		var msg map[string]any
		if err := json.Unmarshal(line, &msg); err != nil {
			log.Printf("tcpsync: bad json from %s: %v", c.RemoteAddr(), err)
			_ = c.sendRaw(AckMessage{Type: "error", Content: "invalid json"})
			continue
		}

		switch msg["type"] {
		case "subscribe":
			uid, _ := msg["user_id"].(string)
			if uid == "" {
				_ = c.sendRaw(AckMessage{Type: "error", Content: "user_id required"})
				continue
			}
			c.hub.Subscribe(c, uid)
			_ = c.sendRaw(AckMessage{Type: "subscribed", Content: uid})
		case "ping":
			_ = c.sendRaw(AckMessage{Type: "pong"})
		default:
			_ = c.sendRaw(AckMessage{Type: "error", Content: "unknown type"})
		}
	}
}

// Send is called by the hub to push a ProgressUpdate. Returns an error if the
// connection is already closed.
func (c *Client) Send(u ProgressUpdate) error {
	return c.sendRaw(u)
}

func (c *Client) sendRaw(v any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return errors.New("closed")
	}
	_ = c.conn.SetWriteDeadline(time.Now().Add(writeDeadline))
	enc := json.NewEncoder(c.writer)
	if err := enc.Encode(v); err != nil {
		return err
	}
	return c.writer.Flush()
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil
	}
	c.closed = true
	return c.conn.Close()
}
