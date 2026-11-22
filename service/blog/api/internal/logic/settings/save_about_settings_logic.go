package settings

import (
	"context"
	"net/mail"
	"net/url"
	"strings"
	"unicode/utf8"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	logicshared "sitehub/service/blog/api/internal/logic"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
	"sitehub/service/blog/model/entity"
)

type SaveAboutSettingsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSaveAboutSettingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveAboutSettingsLogic {
	return &SaveAboutSettingsLogic{ctx, svcCtx}
}

func (l *SaveAboutSettingsLogic) SaveAboutSettings(req *types.AboutSettingsRequest) (*types.Response, error) {
	if _, err := l.svcCtx.Require(l.ctx, "blog:settings:edit", "admin", true); err != nil {
		return nil, err
	}
	profile, err := aboutProfile(req)
	if err != nil {
		return nil, err
	}
	return logicshared.Done(l.svcCtx.Core.Community.SaveAbout(l.ctx, profile))
}

func aboutProfile(req *types.AboutSettingsRequest) (entity.AuthorProfile, error) {
	p := entity.AuthorProfile{Name: strings.TrimSpace(req.Name), Title: strings.TrimSpace(req.Title), Bio: strings.TrimSpace(req.Bio),
		Content: strings.TrimSpace(req.Content), Avatar: strings.TrimSpace(req.Avatar), Location: strings.TrimSpace(req.Location),
		Email: strings.TrimSpace(req.Email), Skills: []string{}, Socials: []entity.SocialLink{}}
	for _, field := range []struct {
		label, value string
		max          int
	}{
		{"博主名称", p.Name, 80}, {"身份或头衔", p.Title, 120}, {"简介", p.Bio, 600}, {"自述", p.Content, 12000}, {"所在地", p.Location, 100}, {"公开邮箱", p.Email, 254},
	} {
		if utf8.RuneCountInString(field.value) > field.max {
			return p, status.Errorf(codes.InvalidArgument, "%s最多 %d 个字符", field.label, field.max)
		}
	}
	if p.Email != "" {
		address, err := mail.ParseAddress(p.Email)
		if err != nil || address.Address != p.Email {
			return p, status.Error(codes.InvalidArgument, "请填写有效的公开邮箱")
		}
	}
	if p.Avatar != "" && !aboutURL(p.Avatar, true) {
		return p, status.Error(codes.InvalidArgument, "头像请使用站内图片路径或 HTTP(S) 图片地址")
	}
	if len(req.Skills) > 32 || len(req.Socials) > 16 {
		return p, status.Error(codes.InvalidArgument, "最多支持 32 个兴趣标签和 16 个社交链接")
	}
	seen := map[string]bool{}
	for _, item := range req.Skills {
		item = strings.TrimSpace(item)
		if item == "" || utf8.RuneCountInString(item) > 40 {
			return p, status.Error(codes.InvalidArgument, "兴趣标签需为 1 至 40 个字符")
		}
		if !seen[item] {
			p.Skills = append(p.Skills, item)
			seen[item] = true
		}
	}
	for _, item := range req.Socials {
		label, link := strings.TrimSpace(item.Label), strings.TrimSpace(item.URL)
		if label == "" || utf8.RuneCountInString(label) > 60 || !aboutURL(link, false) {
			return p, status.Error(codes.InvalidArgument, "请检查社交链接的名称和 HTTP(S) 网址")
		}
		p.Socials = append(p.Socials, entity.SocialLink{Label: label, URL: link})
	}
	return p, nil
}

func aboutURL(value string, allowPath bool) bool {
	if utf8.RuneCountInString(value) > 500 || strings.ContainsAny(value, "\\\r\n\t ") {
		return false
	}
	u, err := url.Parse(value)
	if err != nil || u.User != nil {
		return false
	}
	if allowPath && strings.HasPrefix(value, "/") && !strings.HasPrefix(value, "//") {
		return true
	}
	return (u.Scheme == "http" || u.Scheme == "https") && u.Hostname() != ""
}
