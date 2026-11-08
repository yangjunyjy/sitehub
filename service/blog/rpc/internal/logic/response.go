package logic

import (
	"sitehub/common/protoconv"
	"sitehub/common/rpcx"
	pb "sitehub/service/blog/rpc/pb/blog"

	"google.golang.org/protobuf/proto"
)

func reply[T any](value any, err error) (*T, error) {
	if err != nil {
		return nil, classify(err)
	}
	out := new(T)
	message := any(out).(proto.Message)
	if err = protoconv.Encode(value, message); err != nil {
		return nil, rpcx.Error(err)
	}
	return out, nil
}
func done(err error) (*pb.Empty, error) {
	if err != nil {
		return nil, classify(err)
	}
	return &pb.Empty{}, nil
}
