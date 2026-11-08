package repository

import (
	blogentity "sitehub/service/blog/model/entity"
)

// TagRepository 标签仓储接口
type TagRepository interface {
	WithTransaction(func(TagRepository) error) error
	LockByID(uint) error
	ReferenceCount(uint) (int64, error)

	Create(tag *blogentity.Tag) error
	Update(tag *blogentity.Tag) error
	Delete(id uint) error
	GetByID(id uint) (*blogentity.Tag, error)
	GetBySlug(slug string) (*blogentity.Tag, error)
	List(page, pageSize int, name string) ([]*blogentity.Tag, int64, error)
	GetByArticleID(articleID uint) ([]*blogentity.Tag, error)
	GetAllTags() ([]*blogentity.Tag, error)
	GetTagArticles(tagID uint, page, pageSize int) ([]*blogentity.Article, int64, error)
}
