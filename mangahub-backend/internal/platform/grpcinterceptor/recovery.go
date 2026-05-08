package grpcinterceptor

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Recovery converts panics inside a handler into a codes.Internal error so a
// single bad request does not bring the whole service down.
func Recovery() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("grpc panic method=%s err=%v stack=%s", info.FullMethod, r, debug.Stack())
				err = status.Error(codes.Internal, fmt.Sprintf("internal panic: %v", r))
			}
		}()
		return handler(ctx, req)
	}
}
