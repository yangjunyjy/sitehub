package logic

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"sitehub/common/rpcx"
	domain "sitehub/service/product/domain"
)

func classify(err error) error {
	switch {
	case errors.Is(err, domain.ErrDocumentConflict):
		return status.Error(codes.Aborted, err.Error())
	case errors.Is(err, domain.ErrDocumentNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrInvalidDocument):
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return rpcx.Error(err)
}
