// tcpclient is a CLI demo client for the TCP progress sync server. It connects,
// subscribes to a user_id, and prints every ProgressUpdate it receives.
//
//	go run ./cmd/tcpclient -addr=localhost:9000 -user=665a1d8f3c0e2a4b1c8d9e10
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:9000", "TCP server address")
	user := flag.String("user", "", "user_id to subscribe to (hex ObjectID)")
	flag.Parse()

	if *user == "" {
		fmt.Fprintln(os.Stderr, "missing -user")
		os.Exit(2)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	conn, err := net.DialTimeout("tcp", *addr, 5*time.Second)
	if err != nil {
		log.Fatalf("dial %s: %v", *addr, err)
	}
	defer conn.Close()
	log.Printf("connected to %s as user=%s", *addr, *user)

	// Send subscribe message.
	sub := map[string]string{"type": "subscribe", "user_id": *user}
	b, _ := json.Marshal(sub)
	b = append(b, '\n')
	if _, err := conn.Write(b); err != nil {
		log.Fatalf("subscribe: %v", err)
	}

	// Reader goroutine. Each line from the server is one JSON message.
	done := make(chan struct{})
	go func() {
		defer close(done)
		scanner := bufio.NewScanner(conn)
		scanner.Buffer(make([]byte, 64*1024), 1024*1024)
		for scanner.Scan() {
			line := scanner.Bytes()
			var msg map[string]any
			if err := json.Unmarshal(line, &msg); err != nil {
				log.Printf("bad json: %s", line)
				continue
			}
			pretty, _ := json.MarshalIndent(msg, "", "  ")
			fmt.Printf("\n--- received ---\n%s\n", pretty)
		}
		if err := scanner.Err(); err != nil {
			log.Printf("read: %v", err)
		}
	}()

	select {
	case <-ctx.Done():
		log.Println("interrupted")
	case <-done:
		log.Println("server closed connection")
	}
}
