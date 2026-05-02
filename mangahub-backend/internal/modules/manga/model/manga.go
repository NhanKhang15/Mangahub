package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	MangaStatusOngoing   = "ongoing"
	MangaStatusCompleted = "completed"
	MangaStatusHiatus    = "hiatus"
)

type Manga struct {
	ID          primitive.ObjectID   `json:"id"                     bson:"_id,omitempty"`
	ExternalIDs map[string]string    `json:"external_ids,omitempty" bson:"external_ids,omitempty"`
	Title       string               `json:"title"                  bson:"title"`
	AltTitles   []string             `json:"alt_titles,omitempty"   bson:"alt_titles,omitempty"`
	ArtistIDs   []primitive.ObjectID `json:"artist_ids,omitempty"   bson:"artist_ids,omitempty"`
	AuthorIDs   []primitive.ObjectID `json:"author_ids,omitempty"   bson:"author_ids,omitempty"`
	Description string               `json:"description,omitempty"  bson:"description,omitempty"`
	Status      string               `json:"status"                 bson:"status"`
	Genres      []string             `json:"genres"                 bson:"genres"`
	Tags        []string             `json:"tags,omitempty"         bson:"tags,omitempty"`
	Chapters    int                  `json:"chapters"               bson:"chapters"`
	Rating      float64              `json:"rating"                 bson:"rating"`
	CoverURL    string               `json:"cover_url,omitempty"    bson:"cover_url,omitempty"`
	Popularity  int                  `json:"popularity"             bson:"popularity"`
	CreatedAt   time.Time            `json:"created_at"             bson:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"             bson:"updated_at"`
}

type MangaListQuery struct {
	Page  int
	Limit int
	Genre string
	Tags  []string
	Q     string
	Sort  string // popularity_desc | rating_desc | updated_at_desc | title_asc
}
