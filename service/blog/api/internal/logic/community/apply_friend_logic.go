package community

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/url"
	logicshared "sitehub/service/blog/api/internal/logic"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
	"sitehub/service/blog/model/entity"
	"strings"
	"unicode/utf8"
)

type ApplyFriendLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewApplyFriendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApplyFriendLogic {
	return &ApplyFriendLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ApplyFriendLogic) ApplyFriend(req *types.FriendApplication) (*types.Response, error) {

	member, err := l.svcCtx.Member(l.ctx)
	if err != nil {
		return nil, err
	}
	link := entity.FriendLink{MemberID: member.Id, Name: strings.TrimSpace(req.Name), URL: strings.TrimSpace(req.URL), Description: strings.TrimSpace(req.Description), Avatar: strings.TrimSpace(req.Avatar), Status: "pending"}
	if link.Name == "" || utf8.RuneCountInString(link.Name) > 80 || link.Description == "" || utf8.RuneCountInString(link.Description) > 240 || !webAddress(link.URL) || (link.Avatar != "" && !webAddress(link.Avatar)) {
		return nil, status.Error(codes.InvalidArgument, "请检查站点名称、介绍和网址")
	}
	if err := l.svcCtx.Core.Community.Apply(l.ctx, &link); err != nil {
		return nil, err
	}
	return logicshared.Reply(map[string]string{"status": link.Status}, nil)
}

func webAddress(value string) bool {
	parsed, err := url.ParseRequestURI(value)
	return err == nil && utf8.RuneCountInString(value) <= 500 && (parsed.Scheme == "https" || parsed.Scheme == "http") && parsed.Hostname() != "" && parsed.User == nil
}
