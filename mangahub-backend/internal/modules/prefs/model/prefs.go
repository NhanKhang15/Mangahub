package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Preferences holds per-user UI/feed preferences. Stored in collection
// `user_preferences`, keyed by user_id (1:1).
type Preferences struct {
	UserID         primitive.ObjectID `json:"user_id"                  bson:"_id"`
	FavoriteGenres []string           `json:"favorite_genres,omitempty" bson:"favorite_genres,omitempty"`
	Language       string             `json:"language,omitempty"        bson:"language,omitempty"`
	NSFW           bool               `json:"nsfw"                     bson:"nsfw"`
	UpdatedAt      time.Time          `json:"updated_at"               bson:"updated_at"`
}

// Subscription is one user's interest in a single WS room (e.g. "manga:<id>"
// or "genre:<name>"). Stored in collection `subscriptions`, unique by
// (user_id, room).
type Subscription struct {
	ID        primitive.ObjectID `json:"id"         bson:"_id,omitempty"`
	UserID    primitive.ObjectID `json:"user_id"    bson:"user_id"`
	Room      string             `json:"room"       bson:"room"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
}
