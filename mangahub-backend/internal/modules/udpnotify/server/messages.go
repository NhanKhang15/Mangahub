// Package server implements the UDP notification broadcaster. Clients send a
// REGISTER datagram (carrying an optional client_id) and from then on receive
// every chapter notification the gateway publishes, until they either send
// UNREGISTER or stop heartbeating within the TTL window.
package server

// ControlMessage is the inbound datagram a client sends to the UDP server.
// Type is one of: "REGISTER", "HEARTBEAT", "UNREGISTER", "ACK".
type ControlMessage struct {
	Type     string `json:"type"`
	ClientID string `json:"client_id,omitempty"`
	MsgID    string `json:"msg_id,omitempty"`
}

// Notification is the outbound datagram the server broadcasts to every
// registered client whenever a new chapter event is published.
type Notification struct {
	Type       string `json:"type"`       // "new_chapter"
	MsgID      string `json:"msg_id"`     // unique per broadcast (for ACK)
	MangaID    string `json:"manga_id"`
	MangaTitle string `json:"manga_title,omitempty"`
	Chapter    int    `json:"chapter"`
	Message    string `json:"message,omitempty"`
	Timestamp  int64  `json:"timestamp"`
}
