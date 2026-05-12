package poller

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"mangahub-backend/internal/core/ws"
	"mangahub-backend/internal/gateway/notifier"
	catalogpb "mangahub-backend/proto/catalogpb"
)

// MangaPoller watches active WS rooms and fakes "new chapter" events for
// each subscribed manga. It does not query the manga catalog directly — when
// it does need data, it goes through the catalog gRPC client like any other
// gateway-side caller.
type MangaPoller struct {
	hub      *ws.Hub
	catalog  catalogpb.MangaCatalogClient
	interval time.Duration
	notifier *notifier.Client
}

func NewMangaPoller(hub *ws.Hub, catalog catalogpb.MangaCatalogClient, interval time.Duration, n *notifier.Client) *MangaPoller {
	return &MangaPoller{
		hub:      hub,
		catalog:  catalog,
		interval: interval,
		notifier: n,
	}
}

func (p *MangaPoller) Run(ctx context.Context) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	chapterState := make(map[string]int)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rooms := p.hub.GetActiveRooms()
			for _, room := range rooms {
				if !strings.HasPrefix(room, "manga:") {
					continue
				}
				mangaID := strings.TrimPrefix(room, "manga:")

				// Best-effort sanity check that the manga still exists in
				// the catalog before broadcasting a fake update. Failures
				// are logged but do not stop the poller — the Hub layer is
				// the source of truth for who is subscribed.
				if p.catalog != nil {
					getCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
					if _, err := p.catalog.GetManga(getCtx, &catalogpb.GetMangaRequest{Id: mangaID}); err != nil {
						cancel()
						log.Printf("poller: skip room %s: %v", room, err)
						continue
					}
					cancel()
				}

				current := chapterState[mangaID]
				if current == 0 {
					current = 100
				}
				current++
				chapterState[mangaID] = current

				log.Printf("Poller found new chapter %d for manga %s", current, mangaID)

				p.hub.Broadcast(&ws.RoomMessage{
					Room:    room,
					Type:    "new_chapter",
					Manga:   mangaID,
					Content: fmt.Sprintf("Chapter %d released!", current),
					Chapter: current,
					TS:      time.Now().Format(time.RFC3339),
				})

				// Also fan out via UDP to any registered notification clients.
				if p.notifier != nil {
					ctx2, cancel := context.WithTimeout(ctx, 2*time.Second)
					p.notifier.PublishChapter(ctx2, notifier.ChapterEvent{
						MangaID: mangaID,
						Chapter: current,
						Message: fmt.Sprintf("Chapter %d released!", current),
					})
					cancel()
				}
			}
		}
	}
}
