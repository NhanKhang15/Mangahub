package grpcclient

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"mangahub-backend/internal/core/response"
)

// RespondGRPCError converts a gRPC status into the gateway's standard JSON
// error envelope. Returns false if err == nil so callers can use it as a
// short-circuit in handler bodies.
func RespondGRPCError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	st, _ := status.FromError(err)
	httpStatus, code := codeToHTTP(st.Code())
	response.RespondError(c, httpStatus, code, st.Message(), nil)
	return true
}

func codeToHTTP(c codes.Code) (int, string) {
	switch c {
	case codes.OK:
		return http.StatusOK, "OK"
	case codes.InvalidArgument, codes.FailedPrecondition, codes.OutOfRange:
		return http.StatusBadRequest, "INVALID_ARGUMENT"
	case codes.Unauthenticated:
		return http.StatusUnauthorized, "UNAUTHORIZED"
	case codes.PermissionDenied:
		return http.StatusForbidden, "FORBIDDEN"
	case codes.NotFound:
		return http.StatusNotFound, "NOT_FOUND"
	case codes.AlreadyExists, codes.Aborted:
		return http.StatusConflict, "CONFLICT"
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests, "RATE_LIMITED"
	case codes.Unavailable:
		return http.StatusServiceUnavailable, "UNAVAILABLE"
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout, "TIMEOUT"
	case codes.Unimplemented:
		return http.StatusNotImplemented, "UNIMPLEMENTED"
	default:
		return http.StatusInternalServerError, "INTERNAL"
	}
}
