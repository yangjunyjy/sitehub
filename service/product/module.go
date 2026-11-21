package product

import (
	"gorm.io/gorm"

	settings "sitehub/common/settings/contract"
	filecontract "sitehub/service/blog/contract"
	productservice "sitehub/service/product/business"
	mysql "sitehub/service/product/model/mysql"
)

type Module struct {
	Content  productservice.WebsiteContentService
	Manuals  *productservice.ManualService
	Releases *productservice.ReleaseService
}

func New(db *gorm.DB, store settings.Store, files filecontract.Settings) *Module {
	return &Module{
		Content:  productservice.NewWebsiteContentService(store),
		Manuals:  productservice.NewManualService(mysql.NewProductDocsRepository(db)),
		Releases: productservice.NewReleaseService(db, files),
	}
}
