package services

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"sitehub/common/errorx"
	"sitehub/common/pkg/dberror"
	"sitehub/service/product/contract"
	domain "sitehub/service/product/domain"
)

type ReleaseService struct {
	db    *gorm.DB
	files interface{ ValidateRelease(string) error }
}

func NewReleaseService(db *gorm.DB, files interface{ ValidateRelease(string) error }) *ReleaseService {
	return &ReleaseService{db, files}
}
func releaseMissing() error {
	return &errorx.PublicError{Kind: errorx.NotFound, Message: "下载资源不存在或已下架"}
}
func releaseConflict() error { return errorx.Conflicting("下载资源已变更，请刷新后重试") }
func requireRevision(id string, revision uint64) error {
	if !releaseID.MatchString(id) || revision == 0 {
		return errorx.Invalid("资源标识或修订号无效")
	}
	return nil
}
func fetchRelease(tx *gorm.DB, id string, value any) error {
	err := tx.Where("id = ?", id).Take(value).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return releaseMissing()
	}
	return err
}
func affectedRelease(result *gorm.DB) error {
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return releaseConflict()
	}
	return nil
}
func insertRelease(tx *gorm.DB, row any) error {
	if tx.Dialector.Name() == "sqlite" {
		result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(row)
		if result.Error == nil && result.RowsAffected == 0 {
			return errorx.Conflicting("下载资源标识已被使用")
		}
		return result.Error
	}
	err := tx.Create(row).Error
	if dberror.IsDuplicate(err) {
		return errorx.Conflicting("下载资源标识已被使用")
	}
	return err
}
func releaseGroups(tx *gorm.DB) (rows []domain.ReleaseGroup, err error) {
	err = tx.Order("sort_order ASC, id ASC").Find(&rows).Error
	return
}
func releaseFiles(tx *gorm.DB) (rows []domain.ReleaseFile, err error) {
	err = tx.Order("sort_order ASC, id ASC").Find(&rows).Error
	return
}
func (s *ReleaseService) Groups(ctx context.Context) ([]domain.ReleaseGroup, error) {
	return releaseGroups(s.db.WithContext(ctx))
}
func (s *ReleaseService) Files(ctx context.Context) ([]domain.ReleaseFile, error) {
	return releaseFiles(s.db.WithContext(ctx))
}
func (s *ReleaseService) Catalog(ctx context.Context) (domain.ReleaseCatalog, error) {
	var result domain.ReleaseCatalog
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		if result.ReleaseSettings, err = releaseSettings(tx); err != nil {
			return err
		}
		if result.Groups, err = releaseGroups(tx); err != nil {
			return err
		}
		if result.Packages, err = releaseFiles(tx.Where("status = ?", "published")); err != nil {
			return err
		}
		for i := range result.Packages {
			result.Packages[i].URL = "/v1/file/download?id=" + url.QueryEscape(result.Packages[i].ID)
		}
		return nil
	})
	return result, err
}
func (s *ReleaseService) Resolve(ctx context.Context, id string) (contract.ProductDownload, error) {
	var target contract.ProductDownload
	if !releaseID.MatchString(id) {
		return target, errorx.Invalid("下载文件标识无效")
	}
	var item domain.ReleaseFile
	if err := fetchRelease(s.db.WithContext(ctx).Where("status = ?", "published"), id, &item); err != nil {
		return target, err
	}
	return contract.ProductDownload{URL: item.URL, FileName: item.FileName}, nil
}
func (s *ReleaseService) SaveGroup(ctx context.Context, group domain.ReleaseGroup, create bool) (domain.ReleaseGroup, error) {
	group.Name = strings.TrimSpace(group.Name)
	if !releaseID.MatchString(group.ID) || group.ID == "ungrouped" || group.Name == "" || !validText(group.Name, 80) || !validText(group.Description, 2000) || group.SortOrder < 0 || group.SortOrder > 99999 {
		return group, errorx.Invalid("分组标识、名称、介绍或排序无效")
	}
	tx := s.db.WithContext(ctx)
	if create {
		group.Revision = 1
		err := insertRelease(tx, &group)
		return group, err
	}
	rev := group.Revision
	if err := requireRevision(group.ID, rev); err != nil {
		return group, err
	}
	group.Revision++
	err := affectedRelease(tx.Model(&domain.ReleaseGroup{}).Where("id = ? AND revision = ?", group.ID, rev).Select("*").Omit("id").Updates(group))
	return group, err
}
func (s *ReleaseService) DeleteGroup(ctx context.Context, id string, revision uint64) error {
	if err := requireRevision(id, revision); err != nil {
		return err
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// File assignments acquire this same parent lock before writing their reference.
		var group domain.ReleaseGroup
		if err := fetchRelease(tx.Clauses(clause.Locking{Strength: "UPDATE"}), id, &group); err != nil {
			return err
		}
		if group.Revision != revision {
			return releaseConflict()
		}
		var count int64
		if err := tx.Model(&domain.ReleaseFile{}).Where("group_id = ?", id).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errorx.Conflicting("该分组仍有下载文件，请先移动或删除文件")
		}
		return affectedRelease(tx.Where("id = ? AND revision = ?", id, revision).Delete(&domain.ReleaseGroup{}))
	})
}
func (s *ReleaseService) checkPublishedLocation(item domain.ReleaseFile) error {
	if item.URL == "" {
		return errorx.Invalid("上架前请上传文件或填写下载地址")
	}
	if strings.HasPrefix(item.URL, "/") {
		if s.files == nil {
			return errors.New("release storage unavailable")
		}
		return s.files.ValidateRelease(item.URL)
	}
	return nil
}
func (s *ReleaseService) SaveFile(ctx context.Context, file domain.ReleaseFile, create bool) (domain.ReleaseFile, error) {
	file.Product = strings.TrimSpace(file.Product)
	file.Version = strings.TrimSpace(file.Version)
	file.Platform = strings.TrimSpace(file.Platform)
	file.URL = strings.TrimSpace(file.URL)
	if err := validateFile(file); err != nil {
		return file, err
	}
	db := s.db.WithContext(ctx)
	rev := file.Revision
	file.Status = "draft"
	if !create {
		if err := requireRevision(file.ID, rev); err != nil {
			return file, err
		}
		var prev domain.ReleaseFile
		if err := fetchRelease(db, file.ID, &prev); err != nil {
			return file, err
		}
		if prev.Revision != rev {
			return file, releaseConflict()
		}
		file.Status = prev.Status
		// Remote/local file inspection precedes SQL locking. CAS below fences the validated snapshot.
		if file.Status == "published" && prev.URL != file.URL {
			if err := s.checkPublishedLocation(file); err != nil {
				return file, err
			}
		}
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		if file.GroupID != "" {
			var group domain.ReleaseGroup
			if err := fetchRelease(tx.Clauses(clause.Locking{Strength: "UPDATE"}), file.GroupID, &group); err != nil {
				return err
			}
		}
		if create {
			file.Revision = 1
			return insertRelease(tx, &file)
		}
		file.Revision = rev + 1
		return affectedRelease(tx.Model(&domain.ReleaseFile{}).Where("id = ? AND revision = ?", file.ID, rev).Select("*").Omit("id").Updates(file))
	})
	return file, err
}
func (s *ReleaseService) SetPublished(ctx context.Context, id string, rev uint64, published bool) (domain.ReleaseFile, error) {
	var file domain.ReleaseFile
	if err := requireRevision(id, rev); err != nil {
		return file, err
	}
	db := s.db.WithContext(ctx)
	if err := fetchRelease(db, id, &file); err != nil {
		return file, err
	}
	if file.Revision != rev {
		return file, releaseConflict()
	}
	if published {
		if err := s.checkPublishedLocation(file); err != nil {
			return file, err
		}
		file.Status = "published"
		if file.ReleaseDate == "" {
			file.ReleaseDate = time.Now().Format("2006-01-02")
		}
	} else {
		file.Status = "archived"
	}
	file.Revision++
	err := affectedRelease(db.Model(&domain.ReleaseFile{}).Where("id = ? AND revision = ?", id, rev).Select("status", "release_date", "revision").Updates(file))
	return file, err
}
func (s *ReleaseService) DeleteFile(ctx context.Context, id string, revision uint64) error {
	if err := requireRevision(id, revision); err != nil {
		return err
	}
	return affectedRelease(s.db.WithContext(ctx).Where("id = ? AND revision = ?", id, revision).Delete(&domain.ReleaseFile{}))
}
