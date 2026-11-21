package mysql

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"sitehub/common/errorx"
	"sitehub/common/pkg/dberror"
	"sitehub/service/blog/domain"
	"sitehub/service/blog/model/entity"
)

type ReaderRepository struct{ db *gorm.DB }

func NewReaderRepository(db *gorm.DB) *ReaderRepository { return &ReaderRepository{db} }

func publishedArticles(db *gorm.DB) *gorm.DB {
	return db.Table("articles").
		Joins("JOIN article_publications AS pub ON pub.article_id = articles.id").
		Where("articles.status = 1 AND articles.deleted_at IS NULL")
}

func interaction(db *gorm.DB, member, article uint64) (domain.ArticleInteraction, error) {
	var state domain.ArticleInteraction
	err := publishedArticles(db).Where("articles.id = ?", article).
		Joins("LEFT JOIN article_likes AS liked ON liked.member_id = ? AND liked.article_id = articles.id", member).
		Joins("LEFT JOIN article_favorites AS saved ON saved.member_id = ? AND saved.article_id = articles.id", member).
		Select("articles.likes, articles.favorites, liked.article_id IS NOT NULL AS liked, saved.article_id IS NOT NULL AS favorited, COALESCE(saved.folder_id, 0) AS folder_id").
		Take(&state).Error
	return state, err
}

func (r *ReaderRepository) Interaction(ctx context.Context, member, article uint64) (domain.ArticleInteraction, error) {
	return interaction(r.db.WithContext(ctx), member, article)
}

// Publication and reader writes serialize on the same article row. Recheck visibility
// under that lock before recording a comment, like or favorite.
func lockPublished(db *gorm.DB, id uint64) error {
	var row entity.Article
	err := publishedArticles(db).Select("articles.id").Where("articles.id = ?", id).
		Clauses(clause.Locking{Strength: "UPDATE"}).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.ErrArticleNotFound
	}
	return err
}

func (r *ReaderRepository) change(ctx context.Context, member, article uint64, apply func(*gorm.DB, domain.ArticleInteraction) error) (domain.ArticleInteraction, error) {
	var result domain.ArticleInteraction
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockPublished(tx, article); err != nil {
			return err
		}
		state, err := interaction(tx, member, article)
		if err != nil {
			return err
		}
		if err = apply(tx, state); err != nil {
			return err
		}
		result, err = interaction(tx, member, article)
		return err
	})
	return result, err
}

func (r *ReaderRepository) SetLike(ctx context.Context, member, article uint64, like bool) (domain.ArticleInteraction, error) {
	return r.change(ctx, member, article, func(tx *gorm.DB, state domain.ArticleInteraction) error {
		if state.Liked == like {
			return nil
		}
		row := entity.ArticleLike{MemberID: member, ArticleID: article}
		articles := tx.Model(&entity.Article{}).Where("id = ?", article)
		if like {
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
			return articles.UpdateColumn("likes", gorm.Expr("likes + 1")).Error
		}
		if err := tx.Delete(&row).Error; err != nil {
			return err
		}
		return articles.Where("likes > 0").UpdateColumn("likes", gorm.Expr("likes - 1")).Error
	})
}

// Zero identifies the owner's virtual default folder, including on the first save.
func ownedFolder(db *gorm.DB, member, folder uint64) error {
	if folder == 0 {
		return nil
	}
	var found entity.FavoriteFolder
	return db.Select("id").Where("id = ? AND member_id = ?", folder, member).Take(&found).Error
}

func (r *ReaderRepository) SetFavorite(ctx context.Context, member, article, folder uint64, save bool) (domain.ArticleInteraction, error) {
	return r.change(ctx, member, article, func(tx *gorm.DB, state domain.ArticleInteraction) error {
		row := entity.ArticleFavorite{MemberID: member, ArticleID: article, FolderID: folder}
		articles := tx.Model(&entity.Article{}).Where("id = ?", article)
		if save {
			if err := ownedFolder(tx, member, folder); err != nil {
				return err
			}
			if state.Favorited {
				if state.FolderID == folder {
					return nil
				}
				return tx.Model(&row).UpdateColumn("folder_id", folder).Error
			}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
			return articles.UpdateColumn("favorites", gorm.Expr("favorites + 1")).Error
		}
		if !state.Favorited {
			return nil
		}
		if err := tx.Delete(&row).Error; err != nil {
			return err
		}
		return articles.Where("favorites > 0").UpdateColumn("favorites", gorm.Expr("favorites - 1")).Error
	})
}

const defaultName = "默认收藏夹"

func (r *ReaderRepository) CreateFolder(ctx context.Context, member uint64, name string) (domain.ReaderFolder, error) {
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > 50 {
		return domain.ReaderFolder{}, errorx.Invalid("收藏夹名称需为 1 至 50 个字符")
	}
	if name == defaultName {
		return domain.ReaderFolder{}, errorx.Conflicting("该收藏夹已存在")
	}
	row := entity.FavoriteFolder{MemberID: member, Name: name}
	err := r.db.WithContext(ctx).Create(&row).Error
	if dberror.IsDuplicate(err) {
		return domain.ReaderFolder{}, errorx.Conflicting("该收藏夹已存在")
	}
	return domain.ReaderFolder{ID: row.ID, Name: name}, err
}

func savedArticles(db *gorm.DB, member uint64) *gorm.DB {
	return publishedArticles(db).Joins("JOIN article_favorites AS saved ON saved.article_id = articles.id AND saved.member_id = ?", member)
}

func (r *ReaderRepository) Folders(ctx context.Context, member uint64) ([]domain.ReaderFolder, error) {
	db := r.db.WithContext(ctx)
	var folders []domain.ReaderFolder
	if err := db.Model(&entity.FavoriteFolder{}).Select("id, name").Where("member_id = ?", member).Order("id ASC").Find(&folders).Error; err != nil {
		return nil, err
	}
	folders = append([]domain.ReaderFolder{{Name: defaultName}}, folders...)
	var counts []struct {
		FolderID uint64
		Count    int64
	}
	err := savedArticles(db, member).Select("saved.folder_id, COUNT(*) AS count").Group("saved.folder_id").Scan(&counts).Error
	if err != nil {
		return nil, err
	}
	totals := make(map[uint64]int64, len(counts))
	for _, entry := range counts {
		totals[entry.FolderID] = entry.Count
	}
	for i := range folders {
		folders[i].Count = totals[folders[i].ID]
	}
	return folders, nil
}

func (r *ReaderRepository) Articles(ctx context.Context, member, folder uint64, page, size int) (domain.SavedArticlePage, error) {
	result := domain.SavedArticlePage{Items: []domain.SavedArticle{}}
	if page < 1 || page > 1000000 || size < 1 || size > 100 {
		return result, errorx.Invalid("分页参数无效")
	}
	db := r.db.WithContext(ctx)
	if err := ownedFolder(db, member, folder); err != nil {
		return result, err
	}
	query := savedArticles(db, member).Where("saved.folder_id = ?", folder)
	if err := query.Count(&result.Total).Error; err != nil {
		return result, err
	}
	err := query.Select("articles.id, pub.title, pub.published_at, saved.created_at AS saved_at").
		Order("saved.created_at DESC, articles.id DESC").Limit(size).Offset((page - 1) * size).Scan(&result.Items).Error
	return result, err
}
