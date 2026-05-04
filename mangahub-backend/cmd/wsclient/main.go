package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	token := flag.String("token", "", "JWT token to connect")
	room := flag.String("room", "manga:one-piece", "Room to subscribe to")
	addr := flag.String("addr", "localhost:8080", "http service address")
	flag.Parse()

	if *token == "" {
		log.Fatal("Please provide a JWT token using -token flag")
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws", RawQuery: "token=" + *token}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	// Send subscribe message
	subMsg := map[string]string{
		"type": "subscribe",
		"room": *room,
	}
	subBytes, _ := json.Marshal(subMsg)
	err = c.WriteMessage(websocket.TextMessage, subBytes)
	if err != nil {
		log.Println("write subscribe error:", err)
		return
	}
	log.Printf("sent subscribe to %s", *room)

	ticker := time.NewTicker(time.Second * 54)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			// Client automatically handles standard pong to ping, but we can also send custom ping
			pingMsg := map[string]string{"type": "ping"}
			b, _ := json.Marshal(pingMsg)
			err := c.WriteMessage(websocket.TextMessage, b)
			if err != nil {
				log.Println("write ping error:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
