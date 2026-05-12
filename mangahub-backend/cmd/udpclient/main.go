// udpclient is a CLI demo client for the UDP notification server. It sends a
// REGISTER datagram, periodically heartbeats, and prints every Notification it
// receives. With -ack it also sends an ACK back to the server for each
// notification (demonstrating the optional delivery confirmation feature).
//
//	go run ./cmd/udpclient -server=localhost:9001 -id=desktop-01 -ack
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	server := flag.String("server", "127.0.0.1:9001", "UDP server addr")
	clientID := flag.String("id", "udpclient", "client id")
	ack := flag.Bool("ack", false, "send ACK back on each notification")
	heartbeat := flag.Duration("heartbeat", 30*time.Second, "heartbeat interval")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	saddr, err := net.ResolveUDPAddr("udp", *server)
	if err != nil {
		log.Fatalf("resolve: %v", err)
	}
	// Bind to an ephemeral local port so we can both send and receive.
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	defer conn.Close()
	log.Printf("local addr: %s -> server %s", conn.LocalAddr(), saddr)

	// REGISTER
	reg, _ := json.Marshal(map[string]string{"type": "REGISTER", "client_id": *clientID})
	if _, err := conn.WriteToUDP(reg, saddr); err != nil {
		log.Fatalf("register: %v", err)
	}
	log.Println("REGISTER sent")

	// Heartbeat goroutine
	go func() {
		t := time.NewTicker(*heartbeat)
		defer t.Stop()
		hb, _ := json.Marshal(map[string]string{"type": "HEARTBEAT", "client_id": *clientID})
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				_, _ = conn.WriteToUDP(hb, saddr)
			}
		}
	}()

	// Read loop
	go func() {
		buf := make([]byte, 64*1024)
		for {
			_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			n, _, err := conn.ReadFromUDP(buf)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				if ne, ok := err.(net.Error); ok && ne.Timeout() {
					continue
				}
				log.Printf("read: %v", err)
				continue
			}
			var msg map[string]any
			if err := json.Unmarshal(buf[:n], &msg); err != nil {
				log.Printf("bad json: %s", buf[:n])
				continue
			}
			pretty, _ := json.MarshalIndent(msg, "", "  ")
			fmt.Printf("\n--- received ---\n%s\n", pretty)

			if *ack {
				if mid, _ := msg["msg_id"].(string); mid != "" {
					ackMsg, _ := json.Marshal(map[string]string{"type": "ACK", "msg_id": mid})
					_, _ = conn.WriteToUDP(ackMsg, saddr)
				}
			}
		}
	}()

	<-ctx.Done()
	unreg, _ := json.Marshal(map[string]string{"type": "UNREGISTER", "client_id": *clientID})
	_, _ = conn.WriteToUDP(unreg, saddr)
	log.Println("UNREGISTER sent, exiting")
}
