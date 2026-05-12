// Package server implements the TCP progress sync server. Connected clients
// subscribe to a user_id and receive every progress update the gateway
// publishes for that user. The protocol is newline-delimited JSON over a raw
// TCP socket on port 9000.
package server

// SubscribeMessage is the first message a client must send after connecting.
// It tells the hub which user's progress events this connection should receive.
type SubscribeMessage struct {
	Type   string `json:"type"`    // "subscribe"
	UserID string `json:"user_id"` // hex ObjectID
}

// ProgressUpdate is the broadcast payload pushed to subscribed clients
// whenever the gateway publishes a new reading progress event.
type ProgressUpdate struct {
	Type           string `json:"type"` // "progress_update"
	UserID         string `json:"user_id"`
	MangaID        string `json:"manga_id"`
	MangaTitle     string `json:"manga_title,omitempty"`
	Chapter        int    `json:"chapter"`
	Status         string `json:"status"`
	Timestamp      int64  `json:"timestamp"`
}

// AckMessage is sent by the server to confirm subscription or ping.
type AckMessage struct {
	Type    string `json:"type"`
	Content string `json:"content,omitempty"`
}
