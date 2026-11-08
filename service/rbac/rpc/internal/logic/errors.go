package logic

import (
	"errors"
	"sitehub/common/pkg/e"
	"sitehub/common/rpcx"
	domain "sitehub/service/rbac/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func classify(err error) error {
	switch {
	case errors.Is(err, domain.ErrMemberCredentials), errors.Is(err, domain.ErrMemberSession):
		return status.Error(codes.Unauthenticated, "invalid credentials or expired session")
	case errors.Is(err, domain.ErrMemberRateLimit), errors.Is(err, e.ERROR_OVER_LIMIT):
		return status.Error(codes.ResourceExhausted, "too many requests")
	case errors.Is(err, e.ERROR_PARAMS):
		return status.Error(codes.InvalidArgument, "invalid parameters")
	}
	return rpcx.Error(err)
}
func memberError(err error) error {
	if err == nil {
		return nil
	}
	return classify(err)
}
