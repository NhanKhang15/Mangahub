package catalog

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

func (s *Service) Create(ctx context.Context, m *domain.Manga) (*domain.Manga, error) {
	id, err := s.repo.Create(ctx, m)
	if err != nil {
		return nil, err
	}
	return s.repo.Get(ctx, id)
}

func (s *Service) Get(ctx context.Context, id primitive.ObjectID) (*domain.Manga, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) Update(ctx context.Context, id primitive.ObjectID, set bson.M) (*domain.Manga, error) {
	return s.repo.Update(ctx, id, set)
}

func (s *Service) Delete(ctx context.Context, id primitive.ObjectID) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context, q domain.MangaListQuery) ([]*domain.Manga, int64, error) {
	return s.repo.List(ctx, q)
}

func (s *Service) ListByArtist(ctx context.Context, artistID primitive.ObjectID, page, limit int) ([]*domain.Manga, int64, error) {
	return s.repo.ListByArtist(ctx, artistID, page, limit)
}

func (s *Service) Popular(ctx context.Context, limit int) ([]*domain.Manga, error) {
	return s.repo.Popular(ctx, limit)
}

func (s *Service) Trending(ctx context.Context, limit int) ([]*domain.Manga, error) {
	return s.repo.Trending(ctx, limit)
}
