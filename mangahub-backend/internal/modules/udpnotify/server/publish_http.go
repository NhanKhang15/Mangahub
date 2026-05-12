package server

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
)

// PublishHTTP exposes a small HTTP API that the gateway uses to push a chapter
// notification. The actual fan-out happens over UDP from Server.Broadcast.
type PublishHTTP struct {
	srv   *Server
	token string
}

func NewPublishHTTP(srv *Server, token string) *PublishHTTP {
	return &PublishHTTP{srv: srv, token: token}
}

func (p *PublishHTTP) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", p.handleHealth)
	mux.HandleFunc("/publish", p.handlePublish)
	mux.HandleFunc("/stats", p.handleStats)
	mux.HandleFunc("/ack", p.handleAck)
	return mux
}

func (p *PublishHTTP) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

type publishRequest struct {
	MangaID    string `json:"manga_id"`
	MangaTitle string `json:"manga_title"`
	Chapter    int    `json:"chapter"`
	Message    string `json:"message"`
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

	msgID := newMsgID()
	notif := Notification{
		Type:       "new_chapter",
		MsgID:      msgID,
		MangaID:    req.MangaID,
		MangaTitle: req.MangaTitle,
		Chapter:    req.Chapter,
		Message:    req.Message,
	}
	sent, err := p.srv.Broadcast(notif)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"msg_id":      msgID,
		"sent_to":     sent,
		"registered":  p.srv.Registry().Count(),
	})
}

func (p *PublishHTTP) handleStats(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"registered": p.srv.Registry().Count(),
	})
}

func (p *PublishHTTP) handleAck(w http.ResponseWriter, r *http.Request) {
	msgID := r.URL.Query().Get("msg_id")
	if msgID == "" {
		http.Error(w, "msg_id required", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"msg_id": msgID,
		"acks":   p.srv.AckCount(msgID),
	})
}

func newMsgID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
