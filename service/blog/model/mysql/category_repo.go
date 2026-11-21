package mysql

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	blogrepository "sitehub/service/blog/model"
	blogentity "sitehub/service/blog/model/entity"
)

// CategoryRepositoryImpl 分类仓储实现
type CategoryRepositoryImpl struct {
	db *gorm.DB
}

// NewCategoryRepositoryImpl 创建分类仓储实例
func NewCategoryRepositoryImpl(db *gorm.DB) blogrepository.CategoryRepository {
	return &CategoryRepositoryImpl{db: db}
}

// Create 创建分类
func (r *CategoryRepositoryImpl) Create(category *blogentity.Category) error {
	return r.db.Create(category).Error
}

// Update 更新分类
func (r *CategoryRepositoryImpl) Update(category *blogentity.Category) error {
	return r.db.Save(category).Error
}

// Delete 删除分类
func (r *CategoryRepositoryImpl) Delete(id uint) error {
	return r.db.Delete(&blogentity.Category{}, id).Error
}

// GetByID 根据ID获取分类
func (r *CategoryRepositoryImpl) GetByID(id uint) (*blogentity.Category, error) {
	var category blogentity.Category
	err := r.db.First(&category, id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// GetBySlug 根据slug获取分类
func (r *CategoryRepositoryImpl) GetBySlug(slug string) (*blogentity.Category, error) {
	var category blogentity.Category
	err := r.db.Where("slug = ?", slug).First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// List 分页获取分类列表
func (r *CategoryRepositoryImpl) List(page, pageSize int, name string) ([]*blogentity.Category, int64, error) {
	var categories []*blogentity.Category
	var total int64

	offset := (page - 1) * pageSize
	query := r.db.Model(&blogentity.Category{})
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 预加载文章数量
	err := query.Order("id ASC").Offset(offset).Limit(pageSize).Find(&categories).Error
	if err != nil {
		return nil, 0, err
	}

	if len(categories) == 0 {
		return categories, total, nil
	}
	ids := make([]uint, len(categories))
	for i, category := range categories {
		ids[i] = category.ID
	}
	var counts []struct {
		CategoryID uint
		Total      int64
	}
	const publishedCategory = "COALESCE(ap.category_id,a.category_id)"
	err = r.db.Table("articles a").Joins("LEFT JOIN article_publications ap ON ap.article_id = a.id").
		Select(publishedCategory+" AS category_id, COUNT(*) AS total").
		Where("a.status = 1 AND a.deleted_at IS NULL AND "+publishedCategory+" IN ?", ids).
		Group(publishedCategory).Scan(&counts).Error
	if err != nil {
		return nil, 0, err
	}
	totals := make(map[uint]int64, len(counts))
	for _, row := range counts {
		totals[row.CategoryID] = row.Total
	}
	for _, category := range categories {
		category.ArticleCount = totals[category.ID]
	}

	return categories, total, err
}

// GetTree 获取分类树
func (r *CategoryRepositoryImpl) GetTree() ([]*blogentity.Category, error) {
	var categories []*blogentity.Category
	err := r.db.Where("parent_id IS NULL OR parent_id = 0").Find(&categories).Error
	return categories, err
}

// GetChildren 获取子分类
func (r *CategoryRepositoryImpl) GetChildren(parentID uint) ([]*blogentity.Category, error) {
	var categories []*blogentity.Category
	err := r.db.Where("parent_id = ?", parentID).Find(&categories).Error
	return categories, err
}

// GetCategoryArticles 获取分类下的文章
func (r *CategoryRepositoryImpl) GetCategoryArticles(categoryID uint, page, pageSize int) ([]*blogentity.Article, int64, error) {
	repo := &ArticleRepositoryImpl{db: r.db}
	rows, total, err := repo.WorkflowList(page, pageSize, "", "latest", nil, uint64(categoryID), nil, false, true)
	if err != nil {
		return nil, 0, err
	}
	err = repo.PublicationModels(rows)
	return rows, total, err
}

func (r *CategoryRepositoryImpl) WithTransaction(fn func(blogrepository.CategoryRepository) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error { return fn(&CategoryRepositoryImpl{db: tx}) })
}
func (r *CategoryRepositoryImpl) LockByID(id uint) error {
	var row blogentity.Category
	return r.db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&row, id).Error
}
func (r *CategoryRepositoryImpl) ReferenceCount(id uint) (int64, error) {
	var a, b int64
	if err := r.db.Model(&blogentity.Article{}).Where("category_id = ?", id).Count(&a).Error; err != nil {
		return 0, err
	}
	err := r.db.Model(&blogentity.ArticlePublication{}).Where("category_id = ?", id).Count(&b).Error
	if err != nil {
		return 0, err
	}
	var children int64
	err = r.db.Model(&blogentity.Category{}).Where("parent_id = ?", id).Count(&children).Error
	return a + b + children, err
}
