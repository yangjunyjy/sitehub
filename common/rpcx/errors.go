package rpcx

import (
	"context"
	"errors"
	"log/slog"
	"sitehub/common/errorx"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// Error exposes only classified business failures. Infrastructure details stay in logs.
func Error(err error) error {
	if err == nil {
		return nil
	}
	if _, ok := status.FromError(err); ok {
		return err
	}
	var public *errorx.PublicError
	if errors.As(err, &public) {
		code := codes.InvalidArgument
		switch public.Kind {
		case errorx.Conflict:
			code = codes.Aborted
		case errorx.NotFound:
			code = codes.NotFound
		}
		return status.Error(code, public.Message)
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return status.Error(codes.NotFound, "resource not found")
	}
	if errors.Is(err, context.Canceled) {
		return status.Error(codes.Canceled, "request canceled")
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return status.Error(codes.DeadlineExceeded, "request timed out")
	}
	slog.Error("RPC operation failed", "error", err)
	return status.Error(codes.Internal, "service operation failed")
}
