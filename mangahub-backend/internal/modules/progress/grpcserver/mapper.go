package grpcserver

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/types/known/timestamppb"

	progressModel "mangahub-backend/internal/modules/progress/model"
	progresspb "mangahub-backend/proto/progresspb"
)

func ProgressToProto(p *progressModel.ReadingProgress) *progresspb.Progress {
	if p == nil {
		return nil
	}
	out := &progresspb.Progress{
		Id:             hexOrEmpty(p.ID),
		UserId:         hexOrEmpty(p.UserID),
		MangaId:        hexOrEmpty(p.MangaID),
		Status:         p.Status,
		CurrentChapter: int32(p.CurrentChapter),
		Rating:         p.Rating,
	}
	if !p.LastReadAt.IsZero() {
		out.LastReadAt = timestamppb.New(p.LastReadAt)
	}
	return out
}

func ProgressesToProto(items []*progressModel.ReadingProgress) []*progresspb.Progress {
	out := make([]*progresspb.Progress, 0, len(items))
	for _, p := range items {
		out = append(out, ProgressToProto(p))
	}
	return out
}

func hexOrEmpty(id primitive.ObjectID) string {
	if id.IsZero() {
		return ""
	}
	return id.Hex()
}

// ValidateHex re-exports primitive.ObjectIDFromHex so callers in the gateway
// can convert ids without importing the mongo driver directly.
func ValidateHex(s string) (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(s)
}
