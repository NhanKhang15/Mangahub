// tcp-server is a standalone binary that runs the MangaHub TCP Progress Sync
// service. It serves two ports:
//
//	TCP_PORT  (default 9000) — raw TCP listener that downstream clients connect
//	                           to in order to receive progress updates.
//	HTTP_PORT (default 9100) — internal HTTP API the gateway uses to publish
//	                           events (POST /publish).
//
// The two are decoupled: the TCP socket is outbound-only (server -> client)
// after a client subscribes, and the HTTP API is inbound-only.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mangahub-backend/internal/modules/tcpsync/server"
)

func main() {
	tcpAddr := ":" + getenv("TCP_PORT", "9000")
	httpAddr := ":" + getenv("TCP_HTTP_PORT", "9100")
	token := os.Getenv("INTERNAL_TOKEN")

	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	hub := server.NewHub()
	go hub.Run()
	defer hub.Stop()

	tcpSrv := server.New(tcpAddr, hub)
	go func() {
		if err := tcpSrv.Start(rootCtx); err != nil {
			log.Fatalf("tcp-server: tcp listener: %v", err)
		}
	}()

	publish := server.NewPublishHTTP(hub, token)
	httpSrv := &http.Server{
		Addr:              httpAddr,
		Handler:           publish.Routes(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		log.Printf("tcp-server: publish HTTP on %s", httpAddr)
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("tcp-server: http listener: %v", err)
		}
	}()

	<-rootCtx.Done()
	log.Println("tcp-server: shutting down")
	shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(shutCtx)
	_ = tcpSrv.Stop()
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
