package logic

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
	"sitehub/service/blog/model/entity"
	"strings"
	"unicode/utf8"
)

type SendMessageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSendMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendMessageLogic {
	return &SendMessageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendMessageLogic) SendMessage(req *types.GuestPost) (*types.Response, error) {

	message := entity.GuestMessage{Nickname: strings.TrimSpace(req.Nickname), Content: strings.TrimSpace(req.Content), Status: "pending"}
	if message.Content == "" || utf8.RuneCountInString(message.Content) > 1000 || utf8.RuneCountInString(message.Nickname) > 40 {
		return nil, status.Error(codes.InvalidArgument, "留言须为 1 至 1000 字，昵称最多 40 字")
	}
	if err := l.svcCtx.Core.Community.Submit(l.ctx, &message); err != nil {
		return nil, err
	}
	return reply(map[string]string{"status": message.Status}, nil)
}
