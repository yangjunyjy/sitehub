package community

import (
	"context"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	logicshared "sitehub/service/blog/api/internal/logic"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
	"sitehub/service/blog/model/entity"
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

	member, err := l.svcCtx.Member(l.ctx)
	if err != nil {
		return nil, err
	}
	name := member.Name
	if name == "" {
		name = "读者 " + strconv.FormatUint(member.Id, 10)
	}
	message := entity.GuestMessage{MemberID: member.Id, ArticleID: req.ArticleID, Nickname: name, Content: strings.TrimSpace(req.Content), Status: "pending"}
	if message.Content == "" || utf8.RuneCountInString(message.Content) > 1000 {
		return nil, status.Error(codes.InvalidArgument, "留言须为 1 至 1000 字")
	}
	if err := l.svcCtx.Core.Community.Submit(l.ctx, &message); err != nil {
		return nil, logicshared.Classify(err)
	}
	return logicshared.Reply(map[string]string{"status": message.Status}, nil)
}
