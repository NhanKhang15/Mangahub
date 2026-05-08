package grpcserver

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	prefsService "mangahub-backend/internal/modules/prefs/service"
	"mangahub-backend/internal/platform/grpcinterceptor"
	prefspb "mangahub-backend/proto/prefspb"
)

type Server struct {
	prefspb.UnimplementedUserPreferencesServer
	svc *prefsService.Service
}

func New(svc *prefsService.Service) *Server { return &Server{svc: svc} }

func (s *Server) GetPreferences(ctx context.Context, req *prefspb.GetPreferencesRequest) (*prefspb.Preferences, error) {
	userID, err := resolveUserID(ctx, req.GetUserId())
	if err != nil {
		return nil, err
	}
	p, err := s.svc.GetPreferences(ctx, userID)
	if err != nil {
		return nil, mapServiceErr(err)
	}
	return PreferencesToProto(p), nil
}

func (s *Server) UpdatePreferences(ctx context.Context, req *prefspb.UpdatePreferencesRequest) (*prefspb.Preferences, error) {
	userID, err := resolveUserID(ctx, req.GetUserId())
	if err != nil {
		return nil, err
	}
	patch := req.GetPatch()
	if patch == nil {
		return nil, status.Error(codes.InvalidArgument, "patch required")
	}
	servicePatch := prefsService.PreferencesPatch{}
	if patch.GetFavoriteGenresSet() {
		genres := patch.GetFavoriteGenres()
		servicePatch.FavoriteGenres = &genres
	}
	if patch.GetLanguageSet() {
		lang := patch.GetLanguage()
		servicePatch.Language = &lang
	}
	if patch.GetNsfwSet() {
		nsfw := patch.GetNsfw()
		servicePatch.NSFW = &nsfw
	}
	p, err := s.svc.UpdatePreferences(ctx, userID, servicePatch)
	if err != nil {
		return nil, mapServiceErr(err)
	}
	return PreferencesToProto(p), nil
}

func (s *Server) Subscribe(ctx context.Context, req *prefspb.SubscribeRequest) (*prefspb.Subscription, error) {
	userID, err := resolveUserID(ctx, req.GetUserId())
	if err != nil {
		return nil, err
	}
	if req.GetRoom() == "" {
		return nil, status.Error(codes.InvalidArgument, "room required")
	}
	sub, err := s.svc.Subscribe(ctx, userID, req.GetRoom())
	if err != nil {
		return nil, mapServiceErr(err)
	}
	return SubscriptionToProto(sub), nil
}

func (s *Server) Unsubscribe(ctx context.Context, req *prefspb.UnsubscribeRequest) (*emptypb.Empty, error) {
	userID, err := resolveUserID(ctx, req.GetUserId())
	if err != nil {
		return nil, err
	}
	if req.GetRoom() == "" {
		return nil, status.Error(codes.InvalidArgument, "room required")
	}
	if err := s.svc.Unsubscribe(ctx, userID, req.GetRoom()); err != nil {
		return nil, mapServiceErr(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) ListSubscriptions(ctx context.Context, req *prefspb.ListSubscriptionsRequest) (*prefspb.ListSubscriptionsResponse, error) {
	userID, err := resolveUserID(ctx, req.GetUserId())
	if err != nil {
		return nil, err
	}
	items, err := s.svc.ListSubscriptions(ctx, userID)
	if err != nil {
		return nil, mapServiceErr(err)
	}
	return &prefspb.ListSubscriptionsResponse{Items: SubscriptionsToProto(items)}, nil
}

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
	case errors.Is(err, prefsService.ErrNotFound):
		return status.Error(codes.NotFound, "subscription not found")
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
