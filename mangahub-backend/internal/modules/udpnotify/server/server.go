package server

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

const (
	maxDatagram = 64 * 1024
	gcInterval  = 30 * time.Second
)

// Server owns the UDP socket and a Registry. It reads control datagrams from
// clients on one path and broadcasts Notifications on the other. ACK datagrams
// are counted per msg_id so callers can later inspect delivery confirmation.
type Server struct {
	addr     string
	registry *Registry

	mu       sync.Mutex
	conn     *net.UDPConn

	ackMu   sync.Mutex
	ackByID map[string]int // msg_id -> ack count
}

func New(addr string, ttl time.Duration) *Server {
	return &Server{
		addr:     addr,
		registry: NewRegistry(ttl),
		ackByID:  make(map[string]int),
	}
}

func (s *Server) Registry() *Registry { return s.registry }

func (s *Server) AckCount(msgID string) int {
	s.ackMu.Lock()
	defer s.ackMu.Unlock()
	return s.ackByID[msgID]
}

// Start opens the UDP socket and runs the read loop + GC loop until ctx is
// cancelled. Blocking.
func (s *Server) Start(ctx context.Context) error {
	laddr, err := net.ResolveUDPAddr("udp", s.addr)
	if err != nil {
		return err
	}
	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.conn = conn
	s.mu.Unlock()
	log.Printf("udpnotify: listening on %s", conn.LocalAddr())

	gcStop := make(chan struct{})
	go s.gcLoop(gcStop)
	defer close(gcStop)

	go func() {
		<-ctx.Done()
		_ = conn.Close()
	}()

	buf := make([]byte, maxDatagram)
	for {
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			log.Printf("udpnotify: read: %v", err)
			continue
		}
		s.handle(buf[:n], addr)
	}
}

func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conn == nil {
		return nil
	}
	err := s.conn.Close()
	s.conn = nil
	return err
}

func (s *Server) handle(payload []byte, addr *net.UDPAddr) {
	var msg ControlMessage
	if err := json.Unmarshal(payload, &msg); err != nil {
		log.Printf("udpnotify: bad json from %s: %v", addr, err)
		return
	}
	switch msg.Type {
	case "REGISTER":
		if s.registry.Register(addr, msg.ClientID) {
			log.Printf("udpnotify: REGISTER %s (id=%s)", addr, msg.ClientID)
		} else {
			log.Printf("udpnotify: re-REGISTER %s", addr)
		}
		s.sendRaw(addr, map[string]string{"type": "registered"})
	case "HEARTBEAT":
		if !s.registry.Touch(addr) {
			// auto-register on heartbeat from unknown client
			s.registry.Register(addr, msg.ClientID)
		}
	case "UNREGISTER":
		if s.registry.Remove(addr) {
			log.Printf("udpnotify: UNREGISTER %s", addr)
		}
	case "ACK":
		s.recordAck(msg.MsgID)
	default:
		log.Printf("udpnotify: unknown type %q from %s", msg.Type, addr)
	}
}

func (s *Server) recordAck(msgID string) {
	if msgID == "" {
		return
	}
	s.ackMu.Lock()
	defer s.ackMu.Unlock()
	s.ackByID[msgID]++
}

func (s *Server) sendRaw(addr *net.UDPAddr, v any) {
	b, err := json.Marshal(v)
	if err != nil {
		return
	}
	s.mu.Lock()
	conn := s.conn
	s.mu.Unlock()
	if conn == nil {
		return
	}
	_, _ = conn.WriteToUDP(b, addr)
}

// Broadcast sends the notification to every active registered client. Returns
// the number of datagrams successfully written.
func (s *Server) Broadcast(n Notification) (int, error) {
	s.mu.Lock()
	conn := s.conn
	s.mu.Unlock()
	if conn == nil {
		return 0, errors.New("udp server not running")
	}

	if n.Timestamp == 0 {
		n.Timestamp = time.Now().Unix()
	}
	b, err := json.Marshal(n)
	if err != nil {
		return 0, err
	}
	targets := s.registry.Active()
	var sent int32
	for _, addr := range targets {
		_ = conn.SetWriteDeadline(time.Now().Add(500 * time.Millisecond))
		if _, err := conn.WriteToUDP(b, addr); err != nil {
			log.Printf("udpnotify: write to %s: %v", addr, err)
			continue
		}
		atomic.AddInt32(&sent, 1)
	}
	return int(sent), nil
}

func (s *Server) gcLoop(stop <-chan struct{}) {
	tick := time.NewTicker(gcInterval)
	defer tick.Stop()
	for {
		select {
		case <-stop:
			return
		case <-tick.C:
			if removed := s.registry.GC(); removed > 0 {
				log.Printf("udpnotify: GC removed %d stale clients", removed)
			}
		}
	}
}
