//go:build ignore

// scripts/seed.go — one-shot seeder used to bulk-load the top trending manga
// into the local Mongo. The build tag keeps it out of normal `go build`.
//
// Usage:
//
//	go run ./scripts/seed.go --top 100
//
// Reads MONGO_URI / MONGO_DB / MANGADEX_BASE / etc. from the environment
// (config.Load also reads ./.env when present).
package main

import (
	"context"
	"flag"
	"log"
	"time"

	"mangahub-backend/internal/core/external"
	"mangahub-backend/internal/core/config"
	mongoplat "mangahub-backend/internal/core/database"
	mangaService "mangahub-backend/internal/modules/manga/service"
)

func main() {
	top := flag.Int("top", 100, "number of trending manga to seed")
	flag.Parse()

	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	cl, err := mongoplat.Connect(ctx, cfg.MongoURI, cfg.MongoDB)
	if err != nil {
		log.Fatalf("mongo connect: %v", err)
	}
	defer func() {
		closeCtx, cancelClose := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelClose()
		_ = cl.Close(closeCtx)
	}()
	if err := cl.EnsureIndexes(ctx); err != nil {
		log.Fatalf("ensure indexes: %v", err)
	}

	mangaSvc := mangaService.NewService(mangaService.NewMongoRepo(cl.DB))

	md := external.NewMangaDexClient(cfg.MangaDexBase, cfg.MangaDexToken)
	al := external.NewAniListClient(cfg.AniListBase, cfg.AniListToken)
	var mal *external.MyAnimeListClient
	if cfg.MALClientID != "" {
		mal = external.NewMyAnimeListClient(cfg.MALBase, cfg.MALClientID)
	}
	agg := external.NewAggregator(md, mal, al)

	log.Printf("fetching top %d trending manga…", *top)
	items, err := agg.Trending(ctx, *top)
	if err != nil {
		log.Fatalf("trending: %v", err)
	}
	log.Printf("fetched %d manga; upserting…", len(items))

	var inserted, updated, skipped int
	for _, m := range items {
		action, _, err := mangaSvc.UpsertExternal(ctx, m)
		if err != nil {
			log.Printf("  skip %q: %v", m.Title, err)
			skipped++
			continue
		}
		switch action {
		case mangaService.UpsertInserted:
			inserted++
		case mangaService.UpsertUpdated:
			updated++
		}
	}
	log.Printf("seed done: inserted=%d updated=%d skipped=%d", inserted, updated, skipped)
}
