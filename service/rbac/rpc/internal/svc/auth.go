package svc

import (
	"context"
	"errors"
	"sitehub/common/pkg/session"
	"sitehub/common/protoconv"
	"sitehub/common/rpcx"
	domain "sitehub/service/rbac/domain"
	pb "sitehub/service/rbac/rpc/pb/rbac"
	"slices"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ServiceContext) Require(ctx context.Context, permission, audience string, secure bool) (*pb.Principal, error) {
	return s.Authorize(ctx, rpcx.Incoming(ctx, "authorization"), permission, audience, secure, rpcx.Incoming(ctx, "x-totp-token"))
}

// Authorize resolves live account state and permissions; no actor ID is accepted from HTTP.
func (s *ServiceContext) Authorize(ctx context.Context, credential, permission, audience string, secure bool, totp string) (*pb.Principal, error) {
	if credential == "" {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}
	if audience == "member" {
		user, err := s.Core.Members.Profile(ctx, credential)
		if errors.Is(err, domain.ErrMemberSession) || errors.Is(err, domain.ErrMemberMissing) {
			return nil, status.Error(codes.Unauthenticated, "member session expired")
		}
		if err != nil {
			return nil, rpcx.Error(err)
		}
		return &pb.Principal{Id: user.ID, Email: user.Email, Member: true}, nil
	}
	if audience != "admin" {
		return nil, status.Error(codes.InvalidArgument, "unknown authentication audience")
	}
	info, err := s.Core.Sessions.Verify(credential)
	if errors.Is(err, session.ErrSessionExpired) || errors.Is(err, session.ErrSessionNotFound) {
		return nil, status.Error(codes.Unauthenticated, "session expired")
	}
	if err != nil {
		return nil, rpcx.Error(err)
	}
	user, err := s.Core.Users.GetUserInfo(uint64(info.UserID))
	if err != nil {
		return nil, rpcx.Error(err)
	}
	if user.Status != 1 || user.Deleted != 0 {
		return nil, status.Error(codes.PermissionDenied, "account disabled")
	}
	permissions, err := s.Core.Permissions.GetUserPermissions(user.ID)
	if err != nil {
		return nil, rpcx.Error(err)
	}
	super := slices.Contains(permissions, "*")
	if permission != "" && !super && !slices.Contains(permissions, permission) {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}
	if secure && s.Config.TOTP.Enabled {
		ok, err := s.TOTP.Verify(totp)
		if err != nil || !ok {
			return nil, status.Error(codes.PermissionDenied, "invalid TOTP")
		}
	}
	return &pb.Principal{Id: user.ID, Name: user.Username, Email: user.Email, SuperAdmin: super, Permissions: permissions}, nil
}

func (s *ServiceContext) Login(ctx context.Context, in *pb.SysLoginRequest) (*pb.AdminSession, error) {
	if strings.TrimSpace(in.Username) == "" || in.Password == "" || in.Email == "" || len(in.Code) != 6 {
		return nil, status.Error(codes.InvalidArgument, "username, password, email and verification code are required")
	}
	// Bound repeated credential attempts independently of email delivery throttling.
	key := "native:login:" + rpcx.Incoming(ctx, "x-client-ip")
	count, err := s.Cache.Incr(ctx, key).Result()
	if err != nil {
		return nil, rpcx.Error(err)
	}
	if count == 1 {
		if err = s.Cache.Expire(ctx, key, 15*time.Minute).Err(); err != nil {
			return nil, rpcx.Error(err)
		}
	}
	if count > 30 {
		return nil, status.Error(codes.ResourceExhausted, "too many login attempts")
	}
	user, _, err := s.Core.Users.Login(in.Username, in.Password)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}
	if !strings.EqualFold(user.Email, strings.TrimSpace(in.Email)) {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}
	// Compare-and-delete is atomic: a verification code permits only one login.
	used, err := s.Cache.Eval(ctx, `if redis.call('GET',KEYS[1])==ARGV[1] then redis.call('DEL',KEYS[1]);return 1 end;return 0`, []string{"sysCaptcha:" + in.Email}, in.Code).Int()
	if err != nil {
		return nil, rpcx.Error(err)
	}
	if used != 1 {
		return nil, status.Error(codes.Unauthenticated, "invalid verification code")
	}
	if err = s.Core.Sessions.DeleteAllByUserID(uint(user.ID)); err != nil {
		return nil, rpcx.Error(err)
	}
	info := &session.SessionInfo{UserID: uint(user.ID), Username: user.Username, Email: user.Email, Nickname: user.Nickname, RemoteAddr: rpcx.Incoming(ctx, "x-client-ip"), UserAgent: rpcx.Incoming(ctx, "x-client-agent")}
	credential, err := s.Core.Sessions.Create(info)
	if err != nil {
		return nil, rpcx.Error(err)
	}
	out := &pb.AdminSession{Token: credential, ExpiresAt: info.ExpiresAt.Unix(), User: &pb.EntitySysUser{}}
	if err = protoconv.Encode(user, out.User); err != nil {
		_ = s.Core.Sessions.Delete(credential)
		return nil, rpcx.Error(err)
	}
	return out, nil
}
