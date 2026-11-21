package console

import (
	"bytes"
	"context"
	"encoding/base64"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"image"
	"image/png"
	pb "sitehub/service/rbac/rpc/pb/rbac"
	"sync/atomic"
	"time"
)

// Browser fixture for the RPC boundary. No SMTP or production credentials are used.
type fixtureAuth struct {
	pb.RbacClient
	verified, mailed, revoked atomic.Bool
}

func (f *fixtureAuth) Captcha(ctx context.Context, _ *pb.Empty, _ ...grpc.CallOption) (*pb.SliderCaptchaResult, error) {
	thumbnail := func(w, h int) string {
		var buf bytes.Buffer
		_ = png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, w, h)))
		return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
	}
	return &pb.SliderCaptchaResult{CaptchaKey: "test-challenge", BackgroundImage: thumbnail(300, 150), SliderImage: thumbnail(52, 52), Y: 20, ImgWidth: 300, ImgHeight: 150, PieceSize: 44, PiecePadding: 4}, nil
}
func (f *fixtureAuth) VerifyCaptcha(ctx context.Context, in *pb.CaptchaVerifyRequest, _ ...grpc.CallOption) (*pb.BoolReply, error) {
	ok := in.CaptchaKey == "test-challenge" && in.Duration >= 300 && in.MoveCount >= 3
	f.verified.Store(ok)
	return &pb.BoolReply{Value: ok}, nil
}
func (f *fixtureAuth) SendAdminCode(ctx context.Context, in *pb.SendCodeRequest, _ ...grpc.CallOption) (*pb.Empty, error) {
	if in.CaptchaKey != "test-challenge" || in.Email != "ops@example.test" || !f.verified.Swap(false) {
		return nil, status.Error(codes.InvalidArgument, "verify captcha")
	}
	f.mailed.Store(true)
	return &pb.Empty{}, nil
}
func (f *fixtureAuth) Login(ctx context.Context, in *pb.SysLoginRequest, _ ...grpc.CallOption) (*pb.AdminSession, error) {
	md, _ := metadata.FromOutgoingContext(ctx)
	if len(md.Get("x-client-ip")) != 1 || md.Get("x-client-ip")[0] == "" {
		return nil, status.Error(codes.Internal, "missing client identity")
	}
	if in.Username != "operator" || in.Password != "fixture-password" || in.Email != "ops@example.test" || in.Code != "654321" || !f.mailed.Swap(false) {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}
	f.revoked.Store(false)
	return &pb.AdminSession{Token: "operator", ExpiresAt: time.Now().Add(time.Hour).Unix()}, nil
}
func (f *fixtureAuth) Logout(ctx context.Context, _ *pb.Empty, _ ...grpc.CallOption) (*pb.Empty, error) {
	md, _ := metadata.FromOutgoingContext(ctx)
	if tokens := md.Get("authorization"); len(tokens) != 1 || tokens[0] != "operator" {
		return nil, status.Error(codes.Internal, "wrong session")
	}
	f.revoked.Store(true)
	return &pb.Empty{}, nil
}
