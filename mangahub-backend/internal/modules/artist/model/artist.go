package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	ArtistRoleArtist = "artist"
	ArtistRoleAuthor = "author"
	ArtistRoleBoth   = "both"
)

type Artist struct {
	ID          primitive.ObjectID   `json:"id"                     bson:"_id,omitempty"`
	ExternalIDs map[string]string    `json:"external_ids,omitempty" bson:"external_ids,omitempty"`
	Name        string               `json:"name"                   bson:"name"`
	Role        string               `json:"role"                   bson:"role"`
	Bio         string               `json:"bio,omitempty"          bson:"bio,omitempty"`
	MangaIDs    []primitive.ObjectID `json:"manga_ids,omitempty"    bson:"manga_ids,omitempty"`
	CreatedAt   time.Time            `json:"created_at"             bson:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"             bson:"updated_at"`
}
