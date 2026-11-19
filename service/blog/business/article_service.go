package business

import (
	"context"
	"fmt"
	"time"

	blogdomain "sitehub/service/blog/domain"
	blogresponse "sitehub/service/blog/dto/response"
	blogrepository "sitehub/service/blog/model"
	blogentity "sitehub/service/blog/model/entity"
)

type ArticleService interface {
	PublishScheduledArticles(context.Context) (int64, error)
	EditorDetail(blogdomain.ArticleActor, uint64) (blogdomain.ArticleView, error)
	PublicDetail(uint64) (blogdomain.ArticleView, error)
	BrowsePublic(blogdomain.ArticleListQuery) (blogdomain.ArticlePage, error)
	BrowseEditor(blogdomain.ArticleActor, blogdomain.ArticleListQuery) (blogdomain.ArticlePage, error)
	EditorOptions(blogdomain.ArticleActor) (blogdomain.ArticleOptions, error)
	SaveDraft(blogdomain.ArticleActor, blogdomain.SaveArticle) (blogdomain.ArticleView, error)
	ActOnArticle(blogdomain.ArticleActor, blogdomain.ArticleAction) (blogdomain.ArticleView, error)
	History(blogdomain.ArticleActor, uint64) ([]blogdomain.ArticleHistory, error)
	RevisionDetail(blogdomain.ArticleActor, uint64, uint64) (blogdomain.ArticleView, error)

	CreateArticle(article *blogentity.Article, tagIDs []uint) error
	UpdateArticle(article *blogentity.Article, tagIDs []uint) error
	DeleteArticle(id uint) error
	GetArticle(id uint) (*blogentity.Article, error)
	ListArticles(page, pageSize int, title string, status *int, categoryID *uint, tags []uint) ([]*blogresponse.ArticleListItem, int64, error)
	SearchArticles(keyword string, page, pageSize int) ([]*blogentity.Article, int64, error)
	ViewArticle(id uint) error
	RecordViewWithInfo(id uint, userID uint, ipAddress, userAgent string) error
	GetHotArticles(limit int) ([]*blogentity.Article, error)
	GetRecentArticles(limit int) ([]*blogentity.Article, error)
	ToggleTop(id uint, isTop bool) error
}

type articleService struct {
	articleRepo blogrepository.ArticleRepository
}

func NewArticleService(articleRepo blogrepository.ArticleRepository) ArticleService {
	return &articleService{articleRepo: articleRepo}
}

func (s *articleService) GetArticle(id uint) (*blogentity.Article, error) {
	return s.articleRepo.GetByID(id)
}

func (s *articleService) ListArticles(page, pageSize int, title string, status *int, categoryID *uint, tags []uint) ([]*blogresponse.ArticleListItem, int64, error) {
	articles, count, err := s.articleRepo.List(page, pageSize, title, status, categoryID, tags)
	if err != nil {
		return nil, 0, err
	}

	// 转换为响应格式
	articleList := make([]*blogresponse.ArticleListItem, 0, len(articles))
	articleIDs := make([]uint64, len(articles))

	for i, article := range articles {
		articleIDs[i] = article.ID

		item := &blogresponse.ArticleListItem{
			ID:        uint(article.ID),
			Title:     article.Title,
			Summary:   article.Summary,
			Status:    article.Status,
			IsTop:     article.IsTop,
			Views:     int64(article.Views),
			Likes:     int64(article.Likes),
			Favorites: int64(article.Favorites),
			Comments:  int64(article.CommentCount),
			CreatedAt: article.CreatedAt,
			UpdatedAt: article.UpdatedAt,
		}

		// 封面图片
		if article.Cover != "" {
			item.Cover = &article.Cover
		}

		// 分类信息
		if article.Category != nil {
			item.Category = blogresponse.CategoryInfo{
				ID:   uint(article.Category.ID),
				Name: article.Category.Name,
			}
		}

		// 标签信息
		if len(article.Tags) > 0 {
			tags := make([]blogresponse.TagInfo, len(article.Tags))
			for j, tag := range article.Tags {
				tags[j] = blogresponse.TagInfo{
					Name:         tag.Name,
					ArticleCount: int64(tag.ArticleCount),
				}
			}
			item.Tags = tags
		}

		// 作者信息
		if article.Author != nil {
			item.Author = blogresponse.AuthorInfo{
				ID:     uint(article.Author.ID),
				Name:   article.Author.Username,
				Avatar: &article.Author.Avatar,
				Bio:    &article.Author.Bio,
			}
		}

		articleList = append(articleList, item)
	}

	return articleList, count, nil
}

func (s *articleService) SearchArticles(keyword string, page, pageSize int) ([]*blogentity.Article, int64, error) {
	return s.articleRepo.Search(keyword, page, pageSize)
}

func (s *articleService) ViewArticle(id uint) error {
	return s.articleRepo.RecordView(id, 0, "", "")
}

// RecordViewWithInfo 记录浏览（带用户和请求信息，用于前台调用）
func (s *articleService) RecordViewWithInfo(id uint, userID uint, ipAddress, userAgent string) error {
	return s.articleRepo.RecordView(id, userID, ipAddress, userAgent)
}

func (s *articleService) GetHotArticles(limit int) ([]*blogentity.Article, error) {
	return s.articleRepo.GetHotArticles(limit)
}

func (s *articleService) GetRecentArticles(limit int) ([]*blogentity.Article, error) {
	return s.articleRepo.GetRecentArticles(limit)
}

// ToggleTop 切换置顶状态（委托给 Repository，使用 UpdateColumn 避免 GORM 零值跳过）
func (s *articleService) ToggleTop(id uint, isTop bool) error {
	return s.articleRepo.ToggleTop(id, isTop)
}

// 文章与标签在服务拥有的同一事务中提交，避免删除或更新留下部分关系。
func (s *articleService) CreateArticle(article *blogentity.Article, tagIDs []uint) error {
	now := time.Now()
	article.CreatedAt, article.UpdatedAt = now, now
	if article.Status == 1 {
		article.PublishedAt = &now
	}
	return s.articleRepo.WithTransaction(func(repo blogrepository.ArticleRepository) error {
		if err := repo.ValidateReferences(article.CategoryID, tagIDs, nil, article.CreatedBy, article.ID); err != nil {
			return err
		}
		if err := repo.Create(article); err != nil {
			return err
		}
		return repo.SaveArticleTags(uint(article.ID), tagIDs)
	})
}

func (s *articleService) UpdateArticle(article *blogentity.Article, tagIDs []uint) error {
	return s.articleRepo.WithTransaction(func(repo blogrepository.ArticleRepository) error {
		existing, err := repo.LockByID(uint(article.ID))
		if err != nil {
			return err
		}
		if err := repo.ValidateReferences(article.CategoryID, tagIDs, nil, article.CreatedBy, article.ID); err != nil {
			return err
		}
		article.CreatedAt = existing.CreatedAt
		article.UpdatedAt = time.Now()
		article.PublishedAt = existing.PublishedAt
		if article.Status == 1 && existing.Status != 1 {
			now := time.Now()
			article.PublishedAt = &now
		}
		if err := repo.Update(article); err != nil {
			return err
		}
		if tagIDs == nil {
			return nil
		}
		if err := repo.DeleteArticleTags(uint(article.ID)); err != nil {
			return err
		}
		return repo.SaveArticleTags(uint(article.ID), tagIDs)
	})
}

func (s *articleService) DeleteArticle(id uint) error {
	if id == 0 {
		return fmt.Errorf("文章 ID 必须大于 0")
	}
	return s.articleRepo.WithTransaction(func(repo blogrepository.ArticleRepository) error {
		if _, err := repo.LockByID(id); err != nil {
			return err
		}
		if err := repo.DeleteArticleTags(id); err != nil {
			return err
		}
		if err := repo.DeleteArticleViews(id); err != nil {
			return err
		}
		return repo.Delete(id)
	})
}
