package product

import (
	"gorm.io/gorm"

	settings "sitehub/common/settings/contract"
	filecontract "sitehub/service/blog/contract"
	productservice "sitehub/service/product/business"
	mysql "sitehub/service/product/model/mysql"
)

type Module struct {
	Content productservice.WebsiteContentService
}

func New(db *gorm.DB, store settings.Store, files filecontract.Settings) *Module {
	return &Module{Content: productservice.NewWebsiteContentService(store, mysql.NewProductDocsRepository(db), files)}
}
