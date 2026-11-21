package blog

import (
	coreconfig "sitehub/common/config"
	settings "sitehub/common/settings/contract"
	blogservice "sitehub/service/blog/business"
	"sitehub/service/blog/contract"
	mysql "sitehub/service/blog/model/mysql"

	"gorm.io/gorm"
)

type Module struct {
	Community       *mysql.CommunityRepository
	Reader          *mysql.ReaderRepository
	Articles        blogservice.ArticleService
	Categories      blogservice.CategoryService
	Tags            blogservice.TagService
	Statistics      blogservice.StatisticsService
	Assets          blogservice.ArticleAssetService
	StorageSettings contract.Settings
	Uploads         blogservice.UploadService
	PrivateStorage  contract.PrivateStorage
}

func New(db *gorm.DB, uploadConfig coreconfig.UploadConfig, settingsStore settings.Store, storageSecret string) *Module {
	defaults := blogservice.DefaultWebsiteStorage()
	if uploadConfig.Local.UploadPath != "" {
		defaults.LocalPath = uploadConfig.Local.UploadPath
	}
	defaults.URLPrefix = uploadConfig.URLPrefix
	storageSettings := blogservice.NewStorageSettingsWithDefaults(settingsStore, defaults)
	privateStorage := blogservice.NewPrivateStorage(&uploadConfig, storageSettings, storageSecret)
	articles := mysql.NewArticleRepositoryImpl(db)
	return &Module{
		Community:       mysql.NewCommunityRepository(db),
		Reader:          mysql.NewReaderRepository(db),
		Articles:        blogservice.NewArticleService(articles),
		Categories:      blogservice.NewCategoryService(mysql.NewCategoryRepositoryImpl(db)),
		Tags:            blogservice.NewTagService(mysql.NewTagRepositoryImpl(db)),
		Statistics:      blogservice.NewStatisticsService(mysql.NewStatisticsRepository(db)),
		Assets:          blogservice.NewArticleAssetService(articles, privateStorage),
		StorageSettings: storageSettings,
		Uploads:         blogservice.NewUploadService(&uploadConfig, storageSettings),
		PrivateStorage:  privateStorage,
	}
}
