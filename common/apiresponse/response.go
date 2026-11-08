package apiresponse

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Install() {
	httpx.SetErrorHandlerCtx(func(_ context.Context, err error) (int, any) {
		code, message := http.StatusInternalServerError, "服务暂时不可用"
		s := status.Convert(err)
		switch s.Code() {
		case codes.InvalidArgument:
			code, message = http.StatusBadRequest, s.Message()
		case codes.Unauthenticated:
			code, message = http.StatusUnauthorized, s.Message()
		case codes.PermissionDenied:
			code, message = http.StatusForbidden, s.Message()
		case codes.NotFound:
			code, message = http.StatusNotFound, s.Message()
		case codes.AlreadyExists, codes.Aborted, codes.FailedPrecondition:
			code, message = http.StatusConflict, s.Message()
		case codes.ResourceExhausted:
			code, message = http.StatusTooManyRequests, s.Message()
		case codes.Unavailable:
			code = http.StatusServiceUnavailable
		case codes.DeadlineExceeded:
			code = http.StatusGatewayTimeout
		}
		return code, map[string]any{"code": code, "message": message}
	})
}
