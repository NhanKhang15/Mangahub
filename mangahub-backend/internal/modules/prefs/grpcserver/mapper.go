package grpcserver

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	prefsModel "mangahub-backend/internal/modules/prefs/model"
	prefspb "mangahub-backend/proto/prefspb"
)

func PreferencesToProto(p *prefsModel.Preferences) *prefspb.Preferences {
	if p == nil {
		return nil
	}
	out := &prefspb.Preferences{
		UserId:         p.UserID.Hex(),
		FavoriteGenres: p.FavoriteGenres,
		Language:       p.Language,
		Nsfw:           p.NSFW,
	}
	if !p.UpdatedAt.IsZero() {
		out.UpdatedAt = timestamppb.New(p.UpdatedAt)
	}
	return out
}

func SubscriptionToProto(s *prefsModel.Subscription) *prefspb.Subscription {
	if s == nil {
		return nil
	}
	out := &prefspb.Subscription{
		Id:     s.ID.Hex(),
		UserId: s.UserID.Hex(),
		Room:   s.Room,
	}
	if !s.CreatedAt.IsZero() {
		out.CreatedAt = timestamppb.New(s.CreatedAt)
	}
	return out
}

func SubscriptionsToProto(items []*prefsModel.Subscription) []*prefspb.Subscription {
	out := make([]*prefspb.Subscription, 0, len(items))
	for _, s := range items {
		out = append(out, SubscriptionToProto(s))
	}
	return out
}
