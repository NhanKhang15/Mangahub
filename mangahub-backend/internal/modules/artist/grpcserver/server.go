package grpcserver

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	artistService "mangahub-backend/internal/modules/artist/service"
	artistpb "mangahub-backend/proto/artistpb"
	catalogpb "mangahub-backend/proto/catalogpb"
)

type Server struct {
	artistpb.UnimplementedArtistServer
	svc           *artistService.Service
	catalogClient catalogpb.MangaCatalogClient
}

// New constructs an artist gRPC server. catalogClient is required for the
// ListArtistManga RPC, which delegates to catalog-svc to fetch the manga
// catalogued under the artist (artist-svc never queries the manga collection
// directly).
func New(svc *artistService.Service, catalogClient catalogpb.MangaCatalogClient) *Server {
	return &Server{svc: svc, catalogClient: catalogClient}
}

func (s *Server) GetArtist(ctx context.Context, req *artistpb.GetArtistRequest) (*artistpb.ArtistEntity, error) {
	id, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid artist id")
	}
	a, err := s.svc.Get(ctx, id)
	if err != nil {
		return nil, mapServiceErr(err)
	}
	return ArtistToProto(a), nil
}

func (s *Server) CreateArtist(ctx context.Context, req *artistpb.CreateArtistRequest) (*artistpb.ArtistEntity, error) {
	if req.GetArtist() == nil {
		return nil, status.Error(codes.InvalidArgument, "artist payload required")
	}
	a, err := ArtistFromProto(req.GetArtist())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid artist payload: "+err.Error())
	}
	out, err := s.svc.Create(ctx, a)
	if err != nil {
		return nil, mapServiceErr(err)
	}
	return ArtistToProto(out), nil
}

func (s *Server) UpdateArtist(ctx context.Context, req *artistpb.UpdateArtistRequest) (*artistpb.ArtistEntity, error) {
	id, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid artist id")
	}
	set := PatchToBSON(req.GetPatch())
	if set == nil {
		return nil, status.Error(codes.InvalidArgument, "patch is empty")
	}
	out, err := s.svc.Update(ctx, id, set)
	if err != nil {
		return nil, mapServiceErr(err)
	}
	return ArtistToProto(out), nil
}

func (s *Server) DeleteArtist(ctx context.Context, req *artistpb.DeleteArtistRequest) (*emptypb.Empty, error) {
	id, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid artist id")
	}
	if err := s.svc.Delete(ctx, id); err != nil {
		return nil, mapServiceErr(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) ListArtist(ctx context.Context, req *artistpb.ListArtistRequest) (*artistpb.ListArtistResponse, error) {
	page := normalizePage(req.GetPage())
	limit := normalizeLimit(req.GetLimit())
	items, total, err := s.svc.List(ctx, req.GetQ(), page, limit)
	if err != nil {
		return nil, mapServiceErr(err)
	}
	return &artistpb.ListArtistResponse{
		Items: ArtistsToProto(items),
		Page:  int32(page),
		Limit: int32(limit),
		Total: total,
	}, nil
}

func (s *Server) SearchArtist(ctx context.Context, req *artistpb.SearchArtistRequest) (*artistpb.ListArtistResponse, error) {
	page := normalizePage(req.GetPage())
	limit := normalizeLimit(req.GetLimit())
	items, total, err := s.svc.List(ctx, req.GetQuery(), page, limit)
	if err != nil {
		return nil, mapServiceErr(err)
	}
	return &artistpb.ListArtistResponse{
		Items: ArtistsToProto(items),
		Page:  int32(page),
		Limit: int32(limit),
		Total: total,
	}, nil
}

// ListArtistManga delegates to catalog-svc instead of touching the manga
// collection directly. This keeps artist-svc constrained to its own collection
// and gives us a real cross-service gRPC hop in the architecture.
func (s *Server) ListArtistManga(ctx context.Context, req *artistpb.ListArtistMangaRequest) (*artistpb.ListArtistMangaResponse, error) {
	if s.catalogClient == nil {
		return nil, status.Error(codes.FailedPrecondition, "catalog client not configured")
	}
	if _, err := primitive.ObjectIDFromHex(req.GetArtistId()); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid artist id")
	}
	page := normalizePage(req.GetPage())
	limit := normalizeLimit(req.GetLimit())
	resp, err := s.catalogClient.ListMangaByArtist(ctx, &catalogpb.ListMangaByArtistRequest{
		ArtistId: req.GetArtistId(),
		Page:     int32(page),
		Limit:    int32(limit),
	})
	if err != nil {
		return nil, err
	}
	summaries := make([]*artistpb.MangaSummary, 0, len(resp.GetItems()))
	for _, m := range resp.GetItems() {
		summaries = append(summaries, &artistpb.MangaSummary{
			Id:         m.GetId(),
			Title:      m.GetTitle(),
			CoverUrl:   m.GetCoverUrl(),
			Status:     m.GetStatus(),
			Chapters:   m.GetChapters(),
			Rating:     m.GetRating(),
			Popularity: m.GetPopularity(),
		})
	}
	return &artistpb.ListArtistMangaResponse{
		Items: summaries,
		Page:  resp.GetPage(),
		Limit: resp.GetLimit(),
		Total: resp.GetTotal(),
	}, nil
}

func mapServiceErr(err error) error {
	switch {
	case errors.Is(err, artistService.ErrNotFound):
		return status.Error(codes.NotFound, "artist not found")
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func normalizePage(p int32) int {
	if p <= 0 {
		return 1
	}
	return int(p)
}

func normalizeLimit(l int32) int {
	if l <= 0 {
		return 20
	}
	if l > 100 {
		return 100
	}
	return int(l)
}
