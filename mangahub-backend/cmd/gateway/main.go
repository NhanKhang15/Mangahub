package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"mangahub-backend/internal/artist"
	"mangahub-backend/internal/catalog"
	"mangahub-backend/internal/gateway"
	"mangahub-backend/internal/platform/config"
	mongoplat "mangahub-backend/internal/platform/mongo"
	"mangahub-backend/internal/progress"
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

	mangaRepo := catalog.NewMongoRepo(cl.DB)
	artistRepo := artist.NewMongoRepo(cl.DB)
	progressRepo := progress.NewMongoRepo(cl.DB)

	deps := gateway.Deps{
		MongoClient: cl.Mongo,
		MangaSvc:    catalog.NewService(mangaRepo),
		ArtistSvc:   artist.NewService(artistRepo),
		ProgressSvc: progress.NewService(progressRepo),
	}

	r := gateway.NewRouter(cfg.Env, deps)

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
