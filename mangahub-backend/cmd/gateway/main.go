package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"mangahub-backend/internal/core/config"
	mongoplat "mangahub-backend/internal/core/database"
	"mangahub-backend/internal/core/external"
	"mangahub-backend/internal/core/router"
	"mangahub-backend/internal/core/ws"
	"mangahub-backend/internal/core/poller"
	
	artistService "mangahub-backend/internal/modules/artist/service"
	authService "mangahub-backend/internal/modules/auth/service"
	mangaService "mangahub-backend/internal/modules/manga/service"
	progressService "mangahub-backend/internal/modules/progress/service"
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

	mangaRepo := mangaService.NewMongoRepo(cl.DB)
	artistRepo := artistService.NewMongoRepo(cl.DB)
	progressRepo := progressService.NewMongoRepo(cl.DB)
	authRepo := authService.NewMongoUserRepository(cl.DB)

	mdClient := external.NewMangaDexClient(cfg.MangaDexBase, cfg.MangaDexToken)
	var malClient *external.MyAnimeListClient
	if cfg.MALClientID != "" {
		malClient = external.NewMyAnimeListClient(cfg.MALBase, cfg.MALClientID)
	}
	alClient := external.NewAniListClient(cfg.AniListBase, cfg.AniListToken)
	agg := external.NewAggregator(mdClient, malClient, alClient)

	deps := router.Deps{
		MongoClient: cl.Mongo,
		MangaSvc:    mangaService.NewService(mangaRepo),
		ArtistSvc:   artistService.NewService(artistRepo),
		ProgressSvc: progressService.NewService(progressRepo),
		AuthSvc:     authService.NewAuthService(authRepo, cfg.JWTSecret),
		Aggregator:  agg,
		AdminToken:  cfg.AdminToken,
	}

	hub := ws.NewHub()
	go hub.Run()
	
	deps.Hub = hub

	mangaPoller := poller.NewMangaPoller(hub, deps.MangaSvc, cfg.PollInterval)
	go mangaPoller.Run(rootCtx)

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
