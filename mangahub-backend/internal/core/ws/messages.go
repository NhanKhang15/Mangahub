package ws

// BaseMessage represents the incoming JSON message from client
type BaseMessage struct {
	Type string `json:"type"`
	Room string `json:"room,omitempty"`
}

// RoomMessage represents a message to be broadcasted to a specific room
type RoomMessage struct {
	Room    string      `json:"-"` // internal use only
	Type    string      `json:"type"`
	Manga   string      `json:"manga,omitempty"`
	Content string      `json:"content"`
	Chapter int         `json:"chapter,omitempty"`
	TS      string      `json:"ts,omitempty"`
}

// DirectMessage represents a direct notification to a specific user
type DirectMessage struct {
	UserID  string `json:"-"` // internal use only
	Type    string `json:"type"`
	Content string `json:"content"`
}

// SystemMessage represents a system level message (e.g. connected)
type SystemMessage struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}
