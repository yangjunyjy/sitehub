package business

import (
	"errors"
	"time"

	blogdomain "sitehub/service/blog/domain"
	blogrepository "sitehub/service/blog/model"
	blogentity "sitehub/service/blog/model/entity"
)

type CategoryService interface {
	CreateCategory(category *blogentity.Category) error
	UpdateCategory(id uint, category *blogentity.Category) error
	DeleteCategory(id uint) error
	GetCategory(id uint) (*blogentity.Category, error)
	GetCategoryBySlug(slug string) (*blogentity.Category, error)
	ListCategories(page, pageSize int, name string) ([]*blogentity.Category, int64, error)
	GetCategoryTree() ([]*blogentity.Category, error)
	GetCategoryChildren(parentID uint) ([]*blogentity.Category, error)
	GetCategoryArticles(categoryID uint, page, pageSize int) ([]*blogentity.Article, int64, error)
}

type categoryService struct {
	categoryRepo blogrepository.CategoryRepository
}

func NewCategoryService(categoryRepo blogrepository.CategoryRepository) CategoryService {
	return &categoryService{
		categoryRepo: categoryRepo,
	}
}

func (s *categoryService) CreateCategory(category *blogentity.Category) error {
	now := time.Now()
	category.CreatedAt = now
	category.UpdatedAt = now
	return s.categoryRepo.Create(category)
}

func (s *categoryService) UpdateCategory(id uint, category *blogentity.Category) error {
	existing, err := s.categoryRepo.GetByID(id)
	if err != nil {
		return errors.New("category not found")
	}

	category.ID = id
	category.CreatedAt = existing.CreatedAt
	category.UpdatedAt = time.Now()
	return s.categoryRepo.Update(category)
}

func (s *categoryService) DeleteCategory(id uint) error {
	if id == 0 {
		return blogdomain.ErrInvalidArticle
	}
	return s.categoryRepo.WithTransaction(func(repo blogrepository.CategoryRepository) error {
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

func (s *categoryService) GetCategory(id uint) (*blogentity.Category, error) {
	return s.categoryRepo.GetByID(id)
}

func (s *categoryService) GetCategoryBySlug(slug string) (*blogentity.Category, error) {
	return s.categoryRepo.GetBySlug(slug)
}

func (s *categoryService) ListCategories(page, pageSize int, name string) ([]*blogentity.Category, int64, error) {
	return s.categoryRepo.List(page, pageSize, name)
}

func (s *categoryService) GetCategoryTree() ([]*blogentity.Category, error) {
	return s.categoryRepo.GetTree()
}

func (s *categoryService) GetCategoryChildren(parentID uint) ([]*blogentity.Category, error) {
	return s.categoryRepo.GetChildren(parentID)
}

func (s *categoryService) GetCategoryArticles(categoryID uint, page, pageSize int) ([]*blogentity.Article, int64, error) {
	return s.categoryRepo.GetCategoryArticles(categoryID, page, pageSize)
}
