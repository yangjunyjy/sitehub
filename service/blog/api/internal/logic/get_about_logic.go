package logic

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
	"sitehub/service/blog/model/entity"
)

type GetAboutLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAboutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAboutLogic {
	return &GetAboutLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAboutLogic) GetAbout() (*types.Response, error) {
	profile, err := l.svcCtx.Core.Community.Profile(l.ctx)
	if err != nil {
		return nil, err
	}
	if profile.Skills == nil {
		profile.Skills = []string{}
	}
	if profile.Socials == nil {
		profile.Socials = []entity.SocialLink{}
	}
	return reply(map[string]any{"name": profile.Name, "title": profile.Title, "bio": profile.Bio, "content": profile.Content, "avatar": profile.Avatar, "location": profile.Location, "email": profile.Email, "skills": profile.Skills, "socials": profile.Socials}, nil)
}
