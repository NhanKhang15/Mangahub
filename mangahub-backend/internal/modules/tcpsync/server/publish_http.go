package server

import (
	"encoding/json"
	"net/http"
	"time"
)

// PublishHTTP exposes a tiny HTTP API that the gateway uses to publish a
// ProgressUpdate without speaking the TCP protocol itself. This is the only
// inbound path for events. The actual TCP socket on the data port is
// outbound-only (server -> subscribed clients).
type PublishHTTP struct {
	hub   *Hub
	token string // shared secret; empty disables auth (dev only)
}

func NewPublishHTTP(hub *Hub, token string) *PublishHTTP {
	return &PublishHTTP{hub: hub, token: token}
}

func (p *PublishHTTP) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", p.handleHealth)
	mux.HandleFunc("/publish", p.handlePublish)
	mux.HandleFunc("/stats", p.handleStats)
	return mux
}

func (p *PublishHTTP) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

type publishRequest struct {
	UserID     string `json:"user_id"`
	MangaID    string `json:"manga_id"`
	MangaTitle string `json:"manga_title"`
	Chapter    int    `json:"chapter"`
	Status     string `json:"status"`
}

func (p *PublishHTTP) handlePublish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if p.token != "" && r.Header.Get("X-Internal-Token") != p.token {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	var req publishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.UserID == "" {
		http.Error(w, "user_id required", http.StatusBadRequest)
		return
	}

	update := ProgressUpdate{
		Type:       "progress_update",
		UserID:     req.UserID,
		MangaID:    req.MangaID,
		MangaTitle: req.MangaTitle,
		Chapter:    req.Chapter,
		Status:     req.Status,
		Timestamp:  time.Now().Unix(),
	}
	delivered := p.hub.Subscribers(req.UserID)
	queued := p.hub.Publish(update)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"queued":      queued,
		"subscribers": delivered,
	})
}

func (p *PublishHTTP) handleStats(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"clients": p.hub.TotalClients(),
	})
}
