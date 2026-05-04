package poller

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"mangahub-backend/internal/core/ws"
	mangaService "mangahub-backend/internal/modules/manga/service"
)

type MangaPoller struct {
	hub      *ws.Hub
	mangaSvc *mangaService.Service
	interval time.Duration
}

func NewMangaPoller(hub *ws.Hub, mangaSvc *mangaService.Service, interval time.Duration) *MangaPoller {
	return &MangaPoller{
		hub:      hub,
		mangaSvc: mangaSvc,
		interval: interval,
	}
}

func (p *MangaPoller) Run(ctx context.Context) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	// initial chapter state for mock
	chapterState := make(map[string]int)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// "fake poller" logic to trigger new chapter
			rooms := p.hub.GetActiveRooms()
			for _, room := range rooms {
				if strings.HasPrefix(room, "manga:") {
					mangaID := strings.TrimPrefix(room, "manga:")
					
					// Simulate finding a new chapter
					currentChap := chapterState[mangaID]
					if currentChap == 0 {
						currentChap = 100 // starting mock chapter
					}
					currentChap++
					chapterState[mangaID] = currentChap

					log.Printf("Poller found new chapter %d for manga %s", currentChap, mangaID)

					// Broadcast new chapter event
					p.hub.Broadcast(&ws.RoomMessage{
						Room:    room,
						Type:    "new_chapter",
						Manga:   mangaID,
						Content: fmt.Sprintf("Chapter %d released!", currentChap),
						Chapter: currentChap,
						TS:      time.Now().Format(time.RFC3339),
					})
				}
			}
		}
	}
}
