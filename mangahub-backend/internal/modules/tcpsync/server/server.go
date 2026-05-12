package server

import (
	"context"
	"errors"
	"log"
	"net"
	"sync"
)

// Server owns the raw TCP listener and accepts incoming client connections.
// Each accepted conn becomes a *Client managed by the Hub.
type Server struct {
	addr string
	hub  *Hub

	mu       sync.Mutex
	listener net.Listener
	wg       sync.WaitGroup
}

func New(addr string, hub *Hub) *Server {
	return &Server{addr: addr, hub: hub}
}

// Start opens the TCP listener and runs the accept loop until ctx is cancelled
// or Stop is called. Blocking; usually invoked in its own goroutine.
func (s *Server) Start(ctx context.Context) error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.listener = lis
	s.mu.Unlock()

	log.Printf("tcpsync: listening on %s", lis.Addr())

	go func() {
		<-ctx.Done()
		_ = s.Stop()
	}()

	for {
		conn, err := lis.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				s.wg.Wait()
				return nil
			}
			log.Printf("tcpsync: accept error: %v", err)
			continue
		}

		client := NewClient(s.hub, conn)
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			client.Run()
		}()
	}
}

func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.listener == nil {
		return nil
	}
	err := s.listener.Close()
	s.listener = nil
	return err
}
