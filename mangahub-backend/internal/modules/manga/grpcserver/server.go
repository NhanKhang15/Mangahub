package grpcserver

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	mangaModel "mangahub-backend/internal/modules/manga/model"
	mangaService "mangahub-backend/internal/modules/manga/service"
	catalogpb "mangahub-backend/proto/catalogpb"
)

type Server struct {
	catalogpb.UnimplementedMangaCatalogServer
	svc *mangaService.Service
}

func New(svc *mangaService.Service) *Server {
	return &Server{svc: svc}
}

func (s *Server) GetManga(ctx context.Context, req *catalogpb.GetMangaRequest) (*catalogpb.Manga, error) {
	id, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid manga id")
	}
	m, err := s.svc.Get(ctx, id)
	if err != nil {
		return nil, mapServiceErr(err, "manga")
	}
	return MangaToProto(m), nil
}

func (s *Server) CreateManga(ctx context.Context, req *catalogpb.CreateMangaRequest) (*catalogpb.Manga, error) {
	if req.GetManga() == nil {
		return nil, status.Error(codes.InvalidArgument, "manga payload required")
	}
	m, err := MangaFromProto(req.GetManga())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid manga payload: "+err.Error())
	}
	out, err := s.svc.Create(ctx, m)
	if err != nil {
		return nil, mapServiceErr(err, "manga")
	}
	return MangaToProto(out), nil
}

func (s *Server) UpdateManga(ctx context.Context, req *catalogpb.UpdateMangaRequest) (*catalogpb.Manga, error) {
	id, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid manga id")
	}
	set, err := PatchToBSON(req.GetPatch())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid patch: "+err.Error())
	}
	if set == nil {
		return nil, status.Error(codes.InvalidArgument, "patch is empty")
	}
	out, err := s.svc.Update(ctx, id, set)
	if err != nil {
		return nil, mapServiceErr(err, "manga")
	}
	return MangaToProto(out), nil
}

func (s *Server) DeleteManga(ctx context.Context, req *catalogpb.DeleteMangaRequest) (*emptypb.Empty, error) {
	id, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid manga id")
	}
	if err := s.svc.Delete(ctx, id); err != nil {
		return nil, mapServiceErr(err, "manga")
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) ListManga(ctx context.Context, req *catalogpb.ListMangaRequest) (*catalogpb.ListMangaResponse, error) {
	page := normalizePage(req.GetPage())
	limit := normalizeLimit(req.GetLimit())
	q := mangaModel.MangaListQuery{
		Page:  page,
		Limit: limit,
		Genre: req.GetGenre(),
		Tags:  req.GetTags(),
		Q:     req.GetQ(),
		Sort:  req.GetSort(),
	}
	items, total, err := s.svc.List(ctx, q)
	if err != nil {
		return nil, mapServiceErr(err, "manga")
	}
	return &catalogpb.ListMangaResponse{
		Items: MangasToProto(items),
		Page:  int32(page),
		Limit: int32(limit),
		Total: total,
	}, nil
}

func (s *Server) SearchManga(ctx context.Context, req *catalogpb.SearchMangaRequest) (*catalogpb.ListMangaResponse, error) {
	page := normalizePage(req.GetPage())
	limit := normalizeLimit(req.GetLimit())
	q := mangaModel.MangaListQuery{
		Page:  page,
		Limit: limit,
		Q:     req.GetQuery(),
	}
	items, total, err := s.svc.List(ctx, q)
	if err != nil {
		return nil, mapServiceErr(err, "manga")
	}
	return &catalogpb.ListMangaResponse{
		Items: MangasToProto(items),
		Page:  int32(page),
		Limit: int32(limit),
		Total: total,
	}, nil
}

func (s *Server) GetPopularManga(ctx context.Context, req *catalogpb.GetPopularMangaRequest) (*catalogpb.ListMangaResponse, error) {
	limit := normalizeLimit(req.GetLimit())
	items, err := s.svc.Popular(ctx, limit)
	if err != nil {
		return nil, mapServiceErr(err, "manga")
	}
	return &catalogpb.ListMangaResponse{
		Items: MangasToProto(items),
		Limit: int32(limit),
		Total: int64(len(items)),
	}, nil
}

func (s *Server) GetTrendingManga(ctx context.Context, req *catalogpb.GetTrendingMangaRequest) (*catalogpb.ListMangaResponse, error) {
	limit := normalizeLimit(req.GetLimit())
	items, err := s.svc.Trending(ctx, limit)
	if err != nil {
		return nil, mapServiceErr(err, "manga")
	}
	return &catalogpb.ListMangaResponse{
		Items: MangasToProto(items),
		Limit: int32(limit),
		Total: int64(len(items)),
	}, nil
}

func (s *Server) ListMangaByArtist(ctx context.Context, req *catalogpb.ListMangaByArtistRequest) (*catalogpb.ListMangaResponse, error) {
	id, err := primitive.ObjectIDFromHex(req.GetArtistId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid artist id")
	}
	page := normalizePage(req.GetPage())
	limit := normalizeLimit(req.GetLimit())
	items, total, err := s.svc.ListByArtist(ctx, id, page, limit)
	if err != nil {
		return nil, mapServiceErr(err, "manga")
	}
	return &catalogpb.ListMangaResponse{
		Items: MangasToProto(items),
		Page:  int32(page),
		Limit: int32(limit),
		Total: total,
	}, nil
}

func (s *Server) UpsertMangaByExternalIDs(ctx context.Context, req *catalogpb.UpsertMangaByExternalIDsRequest) (*catalogpb.UpsertMangaByExternalIDsResponse, error) {
	if req.GetManga() == nil {
		return nil, status.Error(codes.InvalidArgument, "manga payload required")
	}
	m, err := MangaFromProto(req.GetManga())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid manga payload: "+err.Error())
	}
	action, out, err := s.svc.UpsertExternal(ctx, m)
	if err != nil {
		return nil, mapServiceErr(err, "manga")
	}
	return &catalogpb.UpsertMangaByExternalIDsResponse{
		Manga:  MangaToProto(out),
		Action: string(action),
	}, nil
}

func mapServiceErr(err error, kind string) error {
	switch {
	case errors.Is(err, mangaService.ErrNotFound):
		return status.Error(codes.NotFound, kind+" not found")
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
