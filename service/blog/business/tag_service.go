package business

import (
	"errors"
	"time"

	blogdomain "sitehub/service/blog/domain"
	blogrepository "sitehub/service/blog/model"
	blogentity "sitehub/service/blog/model/entity"
)

type TagService interface {
	CreateTag(tag *blogentity.Tag) error
	UpdateTag(tag *blogentity.Tag) error
	DeleteTag(id uint) error
	GetTag(id uint) (*blogentity.Tag, error)
	GetAllTags() ([]*blogentity.Tag, error)
	GetTagBySlug(slug string) (*blogentity.Tag, error)
	ListTags(page, pageSize int, name string) ([]*blogentity.Tag, int64, error)
	GetArticleTags(articleID uint) ([]*blogentity.Tag, error)
	GetTagArticles(tagID uint, page, pageSize int) ([]*blogentity.Article, int64, error)
}

type tagService struct {
	tagRepo blogrepository.TagRepository
}

func NewTagService(tagRepo blogrepository.TagRepository) TagService {
	return &tagService{
		tagRepo: tagRepo,
	}
}

func (s *tagService) CreateTag(tag *blogentity.Tag) error {
	now := time.Now()
	tag.CreatedAt = now
	return s.tagRepo.Create(tag)
}

func (s *tagService) GetAllTags() ([]*blogentity.Tag, error) {
	return s.tagRepo.GetAllTags()
}

func (s *tagService) UpdateTag(tag *blogentity.Tag) error {
	existing, err := s.tagRepo.GetByID(uint(tag.ID))
	if err != nil {
		return errors.New("tag not found")
	}

	// Tag的ID是uint64，不需要转换
	tag.CreatedAt = existing.CreatedAt
	return s.tagRepo.Update(tag)
}

func (s *tagService) DeleteTag(id uint) error {
	if id == 0 {
		return blogdomain.ErrInvalidArticle
	}
	return s.tagRepo.WithTransaction(func(repo blogrepository.TagRepository) error {
		if err := repo.LockByID(id); err != nil {
			return err
		}
		count, err := repo.ReferenceCount(id)
		if err != nil {
			return err
		}
		if count > 0 {
			return blogdomain.ErrArticleReferenced
		}
		return repo.Delete(id)
	})
}

func (s *tagService) GetTag(id uint) (*blogentity.Tag, error) {
	return s.tagRepo.GetByID(id)
}

func (s *tagService) GetTagBySlug(slug string) (*blogentity.Tag, error) {
	return s.tagRepo.GetBySlug(slug)
}

func (s *tagService) ListTags(page, pageSize int, name string) ([]*blogentity.Tag, int64, error) {
	return s.tagRepo.List(page, pageSize, name)
}

func (s *tagService) GetArticleTags(articleID uint) ([]*blogentity.Tag, error) {
	return s.tagRepo.GetByArticleID(articleID)
}

func (s *tagService) GetTagArticles(tagID uint, page, pageSize int) ([]*blogentity.Article, int64, error) {
	return s.tagRepo.GetTagArticles(tagID, page, pageSize)
}
