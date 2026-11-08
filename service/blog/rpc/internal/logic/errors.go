package logic

import (
	"errors"
	"sitehub/common/rpcx"
	domain "sitehub/service/blog/domain"
	rbacpb "sitehub/service/rbac/rpc/pb/rbac"
	"slices"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func classify(err error) error {
	switch {
	case errors.Is(err, domain.ErrArticleForbidden):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, domain.ErrArticleNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrArticleConflict), errors.Is(err, domain.ErrArticleSlug), errors.Is(err, domain.ErrArticleReferenced), errors.Is(err, domain.ErrArticleState):
		return status.Error(codes.Aborted, err.Error())
	case errors.Is(err, domain.ErrInvalidArticle), errors.Is(err, domain.ErrArticleReference):
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return rpcx.Error(err)
}
func articleActor(p *rbacpb.Principal) domain.ArticleActor {
	has := func(permission string) bool { return p.SuperAdmin || slices.Contains(p.Permissions, permission) }
	return domain.ArticleActor{ID: p.Id, Name: p.Name, CanRead: has("blog:article:list"), CanCreate: has("blog:article:create"), CanUpdate: has("blog:article:update"), CanPublish: has("blog:article:publish"), CanAudit: has("blog:article:audit"), CanDelete: has("blog:article:delete")}
}
