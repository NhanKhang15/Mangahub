package artist

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"mangahub-backend/internal/domain"
)

type Service struct {
	repo Repo
}

func NewService(r Repo) *Service { return &Service{repo: r} }

func (s *Service) Create(ctx context.Context, a *domain.Artist) (*domain.Artist, error) {
	id, err := s.repo.Create(ctx, a)
	if err != nil {
		return nil, err
	}
	return s.repo.Get(ctx, id)
}

func (s *Service) Get(ctx context.Context, id primitive.ObjectID) (*domain.Artist, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) Update(ctx context.Context, id primitive.ObjectID, set bson.M) (*domain.Artist, error) {
	return s.repo.Update(ctx, id, set)
}

func (s *Service) Delete(ctx context.Context, id primitive.ObjectID) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context, q string, page, limit int) ([]*domain.Artist, int64, error) {
	return s.repo.List(ctx, q, page, limit)
}
