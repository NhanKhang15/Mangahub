package service

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"

	prefsModel "mangahub-backend/internal/modules/prefs/model"
)

type Service struct {
	repo Repo
}

func NewService(r Repo) *Service { return &Service{repo: r} }

func (s *Service) GetPreferences(ctx context.Context, userID primitive.ObjectID) (*prefsModel.Preferences, error) {
	return s.repo.GetPreferences(ctx, userID)
}

func (s *Service) UpdatePreferences(ctx context.Context, userID primitive.ObjectID, patch PreferencesPatch) (*prefsModel.Preferences, error) {
	return s.repo.UpdatePreferences(ctx, userID, patch)
}

func (s *Service) Subscribe(ctx context.Context, userID primitive.ObjectID, room string) (*prefsModel.Subscription, error) {
	return s.repo.Subscribe(ctx, userID, room)
}

func (s *Service) Unsubscribe(ctx context.Context, userID primitive.ObjectID, room string) error {
	return s.repo.Unsubscribe(ctx, userID, room)
}

func (s *Service) ListSubscriptions(ctx context.Context, userID primitive.ObjectID) ([]*prefsModel.Subscription, error) {
	return s.repo.ListSubscriptions(ctx, userID)
}
