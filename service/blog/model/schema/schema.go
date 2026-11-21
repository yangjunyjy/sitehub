package schema

import (
	entity "sitehub/service/blog/model/entity"

	"gorm.io/gorm"
)

// Apply creates only this service's schema. It is invoked by the explicit bootstrap tool.
func Apply(db *gorm.DB) error {
	return db.AutoMigrate(&entity.Article{}, &entity.Category{}, &entity.Tag{}, &entity.ArticleTag{}, &entity.ArticleView{}, &entity.ArticleRevision{}, &entity.ArticleOperation{}, &entity.ArticlePublication{}, &entity.ArticlePublicationTag{}, &entity.ArticleAsset{}, &entity.ArticleAssetReference{}, &entity.FriendLink{}, &entity.GuestMessage{}, &entity.AuthorProfile{}, &entity.ArticleLike{}, &entity.ArticleFavorite{}, &entity.FavoriteFolder{})
}
