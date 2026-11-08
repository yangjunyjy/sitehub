package rpcx

import (
	"context"

	"google.golang.org/grpc"
)

func ClientPolicy(check func(context.Context) error) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, in any, info *grpc.UnaryServerInfo, next grpc.UnaryHandler) (any, error) {
		if err := check(ctx); err != nil {
			return nil, err
		}
		return next(ctx, in)
	}
}
func StreamClientPolicy(check func(context.Context) error) grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, next grpc.StreamHandler) error {
		if err := check(stream.Context()); err != nil {
			return err
		}
		return next(srv, stream)
	}
}
