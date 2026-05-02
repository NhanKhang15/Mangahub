package service

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	mangaModel "mangahub-backend/internal/modules/manga/model"
		)

type Service struct {
	repo Repo
}

func NewService(r Repo) *Service { return &Service{repo: r} }

func (s *Service) Create(ctx context.Context, m *mangaModel.Manga) (*mangaModel.Manga, error) {
	id, err := s.repo.Create(ctx, m)
	if err != nil {
		return nil, err
	}
	return s.repo.Get(ctx, id)
}

func (s *Service) Get(ctx context.Context, id primitive.ObjectID) (*mangaModel.Manga, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) Update(ctx context.Context, id primitive.ObjectID, set bson.M) (*mangaModel.Manga, error) {
	return s.repo.Update(ctx, id, set)
}

func (s *Service) Delete(ctx context.Context, id primitive.ObjectID) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context, q mangaModel.MangaListQuery) ([]*mangaModel.Manga, int64, error) {
	return s.repo.List(ctx, q)
}

func (s *Service) ListByArtist(ctx context.Context, artistID primitive.ObjectID, page, limit int) ([]*mangaModel.Manga, int64, error) {
	return s.repo.ListByArtist(ctx, artistID, page, limit)
}

func (s *Service) Popular(ctx context.Context, limit int) ([]*mangaModel.Manga, error) {
	return s.repo.Popular(ctx, limit)
}

func (s *Service) Trending(ctx context.Context, limit int) ([]*mangaModel.Manga, error) {
	return s.repo.Trending(ctx, limit)
}

// UpsertExternal upserts a manga keyed by its external_ids — used by the
// admin import endpoint and the seed script. Returns whether the call
// inserted or updated.
func (s *Service) UpsertExternal(ctx context.Context, m *mangaModel.Manga) (UpsertAction, *mangaModel.Manga, error) {
	return s.repo.UpsertByExternalIDs(ctx, m)
}
