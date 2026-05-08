package grpcinterceptor

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// Logging logs every RPC method, duration and resulting status code.
func Logging() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		code := status.Code(err)
		log.Printf("grpc method=%s code=%s dur=%s", info.FullMethod, code.String(), time.Since(start))
		return resp, err
	}
}
