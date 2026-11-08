package repository

import (
	blogentity "sitehub/service/blog/model/entity"
)

// CategoryRepository 分类仓储接口
type CategoryRepository interface {
	WithTransaction(func(CategoryRepository) error) error
	LockByID(uint) error
	ReferenceCount(uint) (int64, error)

	Create(category *blogentity.Category) error
	Update(category *blogentity.Category) error
	Delete(id uint) error
	GetByID(id uint) (*blogentity.Category, error)
	GetBySlug(slug string) (*blogentity.Category, error)
	List(page, pageSize int, name string) ([]*blogentity.Category, int64, error)
	GetTree() ([]*blogentity.Category, error)
	GetChildren(parentID uint) ([]*blogentity.Category, error)
	GetCategoryArticles(categoryID uint, page, pageSize int) ([]*blogentity.Article, int64, error)
}
