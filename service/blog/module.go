package blog

import (
	"gorm.io/gorm"

	blogservice "sitehub/service/blog/business"
	mysql "sitehub/service/blog/model/mysql"
	filecontract "sitehub/service/file/contract"
)

type Module struct {
	Articles   blogservice.ArticleService
	Categories blogservice.CategoryService
	Tags       blogservice.TagService
	Statistics blogservice.StatisticsService
	Assets     blogservice.ArticleAssetService
}

func New(db *gorm.DB, storage filecontract.PrivateStorage) *Module {
	articles := mysql.NewArticleRepositoryImpl(db)
	return &Module{
		Articles:   blogservice.NewArticleService(articles),
		Categories: blogservice.NewCategoryService(mysql.NewCategoryRepositoryImpl(db)),
		Tags:       blogservice.NewTagService(mysql.NewTagRepositoryImpl(db)),
		Statistics: blogservice.NewStatisticsService(mysql.NewStatisticsRepository(db)),
		Assets:     blogservice.NewArticleAssetService(articles, storage),
	}
}
