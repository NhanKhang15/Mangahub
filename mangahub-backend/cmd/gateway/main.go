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

	"mangahub-backend/internal/core/config"
	mongoplat "mangahub-backend/internal/core/database"
	"mangahub-backend/internal/core/external"
	"mangahub-backend/internal/core/poller"
	"mangahub-backend/internal/core/router"
	"mangahub-backend/internal/core/ws"
	"mangahub-backend/internal/gateway/grpcclient"
	"mangahub-backend/internal/gateway/notifier"

	authService "mangahub-backend/internal/modules/auth/service"
)

func main() {
	cfg := config.Load()

	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cl, err := mongoplat.Connect(rootCtx, cfg.MongoURI, cfg.MongoDB)
	if err != nil {
		log.Fatalf("mongo connect: %v", err)
	}
	defer func() {
		ctx2, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = cl.Close(ctx2)
	}()

	if err := cl.EnsureIndexes(rootCtx); err != nil {
		log.Fatalf("ensure indexes: %v", err)
	}
	log.Printf("mongo OK: db=%s", cfg.MongoDB)

	// Auth still runs in-process at the gateway because every request uses
	// it to authenticate before fanning out to downstream services.
	authRepo := authService.NewMongoUserRepository(cl.DB)
	authSvc := authService.NewAuthService(authRepo, cfg.JWTSecret)

	// Domain reads/writes (manga, artist, progress, prefs) all go through gRPC.
	addrs := grpcclient.Addresses{
		Catalog:  getenv("CATALOG_GRPC_ADDR", "localhost:50051"),
		Artist:   getenv("ARTIST_GRPC_ADDR", "localhost:50052"),
		Progress: getenv("PROGRESS_GRPC_ADDR", "localhost:50053"),
		Prefs:    getenv("PREFS_GRPC_ADDR", "localhost:50054"),
	}
	clients, err := grpcclient.Dial(addrs)
	if err != nil {
		log.Fatalf("grpc dial: %v", err)
	}
	defer clients.Close()
	log.Printf("grpc clients: catalog=%s artist=%s progress=%s prefs=%s",
		addrs.Catalog, addrs.Artist, addrs.Progress, addrs.Prefs)

	mdClient := external.NewMangaDexClient(cfg.MangaDexBase, cfg.MangaDexToken)
	var malClient *external.MyAnimeListClient
	if cfg.MALClientID != "" {
		malClient = external.NewMyAnimeListClient(cfg.MALBase, cfg.MALClientID)
	}
	alClient := external.NewAniListClient(cfg.AniListBase, cfg.AniListToken)
	agg := external.NewAggregator(mdClient, malClient, alClient)

	hub := ws.NewHub()
	go hub.Run()

	notifierClient := notifier.New(cfg.TCPPublishURL, cfg.UDPPublishURL, cfg.InternalToken)
	if cfg.TCPPublishURL != "" || cfg.UDPPublishURL != "" {
		log.Printf("notifier: tcp=%s udp=%s", cfg.TCPPublishURL, cfg.UDPPublishURL)
	}

	mangaPoller := poller.NewMangaPoller(hub, clients.Catalog, cfg.PollInterval, notifierClient)
	go mangaPoller.Run(rootCtx)

	deps := router.Deps{
		MongoClient: cl.Mongo,
		Clients:     clients,
		Aggregator:  agg,
		AdminToken:  cfg.AdminToken,
		AuthSvc:     authSvc,
		Hub:         hub,
		Notifier:    notifierClient,
	}

	r := router.NewRouter(cfg.Env, deps)

	srv := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("gateway listening on :%s", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http server: %v", err)
		}
	}()

	<-rootCtx.Done()
	log.Println("shutting down")
	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		log.Printf("graceful shutdown: %v", err)
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
