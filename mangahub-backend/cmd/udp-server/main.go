// udp-server is a standalone binary that runs the MangaHub UDP Notification
// service.
//
//	UDP_PORT      (default 9001) — UDP socket where clients REGISTER and receive
//	                               chapter notifications.
//	UDP_HTTP_PORT (default 9101) — internal HTTP API the gateway/poller calls to
//	                               publish broadcasts (POST /publish).
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"mangahub-backend/internal/modules/udpnotify/server"
)

func main() {
	udpAddr := ":" + getenv("UDP_PORT", "9001")
	httpAddr := ":" + getenv("UDP_HTTP_PORT", "9101")
	token := os.Getenv("INTERNAL_TOKEN")
	ttl := parseDuration("UDP_TTL", 90*time.Second)

	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	udpSrv := server.New(udpAddr, ttl)
	go func() {
		if err := udpSrv.Start(rootCtx); err != nil {
			log.Fatalf("udp-server: udp listener: %v", err)
		}
	}()

	publish := server.NewPublishHTTP(udpSrv, token)
	httpSrv := &http.Server{
		Addr:              httpAddr,
		Handler:           publish.Routes(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		log.Printf("udp-server: publish HTTP on %s (ttl=%s)", httpAddr, ttl)
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("udp-server: http listener: %v", err)
		}
	}()

	<-rootCtx.Done()
	log.Println("udp-server: shutting down")
	shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(shutCtx)
	_ = udpSrv.Stop()
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func parseDuration(k string, def time.Duration) time.Duration {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	if secs, err := strconv.Atoi(v); err == nil {
		return time.Duration(secs) * time.Second
	}
	if d, err := time.ParseDuration(v); err == nil {
		return d
	}
	return def
}
