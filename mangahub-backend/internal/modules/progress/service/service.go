package service

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"

			progressModel "mangahub-backend/internal/modules/progress/model"
)

type Service struct {
	repo Repo
}

func NewService(r Repo) *Service { return &Service{repo: r} }

func (s *Service) Upsert(ctx context.Context, p *progressModel.ReadingProgress) (*progressModel.ReadingProgress, error) {
	return s.repo.Upsert(ctx, p)
}

func (s *Service) Get(ctx context.Context, userID, mangaID primitive.ObjectID) (*progressModel.ReadingProgress, error) {
	return s.repo.Get(ctx, userID, mangaID)
}

func (s *Service) List(ctx context.Context, userID primitive.ObjectID, status string, page, limit int) ([]*progressModel.ReadingProgress, int64, error) {
	return s.repo.List(ctx, userID, status, page, limit)
}

func (s *Service) Delete(ctx context.Context, userID, mangaID primitive.ObjectID) error {
	return s.repo.Delete(ctx, userID, mangaID)
}

func (s *Service) Stats(ctx context.Context, userID primitive.ObjectID) (map[string]int, error) {
	return s.repo.Stats(ctx, userID)
}
