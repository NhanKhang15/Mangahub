package grpcinterceptor

import (
	"context"

	"google.golang.org/grpc/metadata"
)

// MetadataUserIDKey is the metadata key the gateway uses to forward the
// authenticated user id to downstream services.
const MetadataUserIDKey = "x-user-id"

type ctxKey int

const userIDCtxKey ctxKey = iota

// UserIDFromContext returns the user id forwarded by the gateway via the
// `x-user-id` metadata header. The boolean is false when no value is present.
func UserIDFromContext(ctx context.Context) (string, bool) {
	if v, ok := ctx.Value(userIDCtxKey).(string); ok && v != "" {
		return v, true
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", false
	}
	values := md.Get(MetadataUserIDKey)
	if len(values) == 0 || values[0] == "" {
		return "", false
	}
	return values[0], true
}

// AppendUserID returns a context with the user id added to the outgoing
// metadata, so a gateway client call carries it to the service.
func AppendUserID(ctx context.Context, userID string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, MetadataUserIDKey, userID)
}
