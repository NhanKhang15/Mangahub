package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	ProgressReading     = "reading"
	ProgressCompleted   = "completed"
	ProgressPlanToRead  = "plan_to_read"
	ProgressDropped     = "dropped"
)

type ReadingProgress struct {
	ID             primitive.ObjectID `json:"id"               bson:"_id,omitempty"`
	UserID         primitive.ObjectID `json:"user_id"          bson:"user_id"`
	MangaID        primitive.ObjectID `json:"manga_id"         bson:"manga_id"`
	Status         string             `json:"status"           bson:"status"`
	CurrentChapter int                `json:"current_chapter"  bson:"current_chapter"`
	Rating         float64            `json:"rating,omitempty" bson:"rating,omitempty"`
	LastReadAt     time.Time          `json:"last_read_at"     bson:"last_read_at"`
}
