package grpcserver

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	progressModel "mangahub-backend/internal/modules/progress/model"
	progressService "mangahub-backend/internal/modules/progress/service"
	"mangahub-backend/internal/platform/grpcinterceptor"
	progresspb "mangahub-backend/proto/progresspb"
)

type Server struct {
	progresspb.UnimplementedReadingProgressServer
	svc *progressService.Service
}

func New(svc *progressService.Service) *Server {
	return &Server{svc: svc}
}

func (s *Server) UpsertProgress(ctx context.Context, req *progresspb.UpsertProgressRequest) (*progresspb.Progress, error) {
	userID, err := resolveUserID(ctx, req.GetUserId())
	if err != nil {
		return nil, err
	}
	mangaID, err := primitive.ObjectIDFromHex(req.GetMangaId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid manga id")
	}
	if req.GetStatus() == "" {
		return nil, status.Error(codes.InvalidArgument, "status required")
	}
	out, err := s.svc.Upsert(ctx, &progressModel.ReadingProgress{
		UserID:         userID,
		MangaID:        mangaID,
		Status:         req.GetStatus(),
		CurrentChapter: int(req.GetCurrentChapter()),
		Rating:         req.GetRating(),
	})
	if err != nil {
		return nil, mapServiceErr(err)
	}
	return ProgressToProto(out), nil
}

func (s *Server) GetProgress(ctx context.Context, req *progresspb.GetProgressRequest) (*progresspb.Progress, error) {
	userID, err := resolveUserID(ctx, req.GetUserId())
	if err != nil {
		return nil, err
	}
	mangaID, err := primitive.ObjectIDFromHex(req.GetMangaId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid manga id")
	}
	out, err := s.svc.Get(ctx, userID, mangaID)
	if err != nil {
		return nil, mapServiceErr(err)
	}
	return ProgressToProto(out), nil
}

func (s *Server) ListUserProgress(ctx context.Context, req *progresspb.ListUserProgressRequest) (*progresspb.ListUserProgressResponse, error) {
	userID, err := resolveUserID(ctx, req.GetUserId())
	if err != nil {
		return nil, err
	}
	page := normalizePage(req.GetPage())
	limit := normalizeLimit(req.GetLimit())
	items, total, err := s.svc.List(ctx, userID, req.GetStatus(), page, limit)
	if err != nil {
		return nil, mapServiceErr(err)
	}
	return &progresspb.ListUserProgressResponse{
		Items: ProgressesToProto(items),
		Page:  int32(page),
		Limit: int32(limit),
		Total: total,
	}, nil
}

func (s *Server) DeleteProgress(ctx context.Context, req *progresspb.DeleteProgressRequest) (*emptypb.Empty, error) {
	userID, err := resolveUserID(ctx, req.GetUserId())
	if err != nil {
		return nil, err
	}
	mangaID, err := primitive.ObjectIDFromHex(req.GetMangaId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid manga id")
	}
	if err := s.svc.Delete(ctx, userID, mangaID); err != nil {
		return nil, mapServiceErr(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) Stats(ctx context.Context, req *progresspb.StatsRequest) (*progresspb.StatsResponse, error) {
	userID, err := resolveUserID(ctx, req.GetUserId())
	if err != nil {
		return nil, err
	}
	stats, err := s.svc.Stats(ctx, userID)
	if err != nil {
		return nil, mapServiceErr(err)
	}
	buckets := make(map[string]int32, len(stats))
	for k, v := range stats {
		buckets[k] = int32(v)
	}
	return &progresspb.StatsResponse{Buckets: buckets}, nil
}

// resolveUserID prefers the metadata-forwarded user id from the gateway, but
// falls back to the value in the request body — useful for grpcurl debugging.
func resolveUserID(ctx context.Context, fromReq string) (primitive.ObjectID, error) {
	id := fromReq
	if mdID, ok := grpcinterceptor.UserIDFromContext(ctx); ok {
		id = mdID
	}
	if id == "" {
		return primitive.NilObjectID, status.Error(codes.Unauthenticated, "user id missing (set x-user-id metadata)")
	}
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return primitive.NilObjectID, status.Error(codes.InvalidArgument, "invalid user id")
	}
	return oid, nil
}

func mapServiceErr(err error) error {
	switch {
	case errors.Is(err, progressService.ErrNotFound):
		return status.Error(codes.NotFound, "reading progress not found")
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
