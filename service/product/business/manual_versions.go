package services

import (
	"fmt"
	"strings"
	"time"

	productdomain "sitehub/service/product/domain"
	productrepository "sitehub/service/product/model"
)

func (s *ManualService) SaveVersion(in productdomain.DocVersionInput, actor uint64) error {
	if !docVersionID.MatchString(in.ID) {
		return invalidDoc("版本标识无效")
	}
	return s.mutateDocs(func(r productrepository.ProductDocsRepository, site *productdomain.DocSite) error {
		versions, err := r.Versions()
		if err != nil {
			return err
		}
		var current *productdomain.DocVersion
		for i := range versions {
			v := &versions[i]
			if v.ID == in.ID {
				current = v
			}
			if v.ID != in.ID && v.ModeID == in.ModeID && v.Name == strings.TrimSpace(in.Name) {
				return invalidDoc("版本名称已存在")
			}
		}
		if current != nil && in.ModeID != "" && in.ModeID != current.ModeID {
			return invalidDoc("版本所属模式不可直接变更，请复制为新版本")
		}
		if strings.TrimSpace(in.Name) == "" || len([]rune(in.Name)) > 60 {
			return invalidDoc("请填写有效的版本名称")
		}
		if current != nil {
			if current.Revision != in.Revision {
				return ErrWebsiteDocsConflict
			}
			current.Name = strings.TrimSpace(in.Name)
			current.Revision++
			return r.PutVersion(*current)
		}
		if in.Revision != 0 {
			return ErrWebsiteDocumentNotFound
		}
		if _, err = findMode(r, in.ModeID); err != nil {
			return err
		}
		if len(versions) >= 50 {
			return invalidDoc("最多保留 50 个产品版本")
		}
		if in.Source != "" {
			if _, err = findVersion(r, in.Source); err != nil {
				return err
			}
		}
		v := productdomain.DocVersion{ID: in.ID, Name: strings.TrimSpace(in.Name), Status: "draft", Revision: 1, ModeID: in.ModeID}
		if err = r.PutVersion(v); err != nil {
			return err
		}
		if in.Source == "" {
			return r.PutGroup(&productdomain.DocGroup{VersionID: in.ID, Title: "快速开始", Revision: 1})
		}
		groups, err := r.Groups(in.Source)
		if err != nil {
			return err
		}
		docs, err := r.Documents(in.Source, true)
		if err != nil {
			return err
		}
		for _, g := range groups {
			old := g.ID
			g.ID = 0
			g.VersionID = in.ID
			g.Revision = 1
			if err = r.PutGroup(&g); err != nil {
				return err
			}
			for _, d := range docs {
				if d.GroupID != old {
					continue
				}
				d.ID = 0
				d.VersionID = in.ID
				d.GroupID = g.ID
				d.Revision = 1
				d.PublishedRevision = 0
				d.UpdatedBy = actor
				d.CreatedAt = time.Time{}
				d.UpdatedAt = time.Time{}
				if err = r.PutDocument(&d); err != nil {
					return err
				}
				if err = snapshotDocument(r, d, g.Title, "copy", actor); err != nil {
					return err
				}
			}
		}
		return nil
	})
}
func (s *ManualService) changeVersion(id string, revision uint64, mutate func(productrepository.ProductDocsRepository, *productdomain.DocSite, productdomain.DocMode, productdomain.DocVersion) error) error {
	return s.mutateDocs(func(r productrepository.ProductDocsRepository, site *productdomain.DocSite) error {
		v, err := findVersion(r, id)
		if err != nil {
			return err
		}
		if v.Revision != revision {
			return ErrWebsiteDocsConflict
		}
		mode, err := findMode(r, v.ModeID)
		if err != nil {
			return err
		}
		return mutate(r, site, mode, v)
	})
}
func (s *ManualService) SetVersionPublished(id string, revision uint64, publish bool) error {
	return s.changeVersion(id, revision, func(r productrepository.ProductDocsRepository, site *productdomain.DocSite, mode productdomain.DocMode, v productdomain.DocVersion) error {
		if publish {
			docs, err := r.PublishedDocuments(id)
			if err != nil {
				return err
			}
			if len(docs) == 0 {
				return invalidDoc("请先发布至少一篇文档，再发布版本")
			}
			v.Status = "published"
		} else {
			if id == site.DefaultVersion || id == mode.DefaultVersion {
				return invalidDoc("请先切换默认版本")
			}
			v.Status = "draft"
		}
		v.Revision++
		return r.PutVersion(v)
	})
}
func (s *ManualService) DefaultVersion(id string, revision uint64) error {
	return s.changeVersion(id, revision, func(r productrepository.ProductDocsRepository, site *productdomain.DocSite, mode productdomain.DocMode, v productdomain.DocVersion) error {
		if v.Status != "published" {
			return invalidDoc("默认版本必须已发布")
		}
		if mode.DefaultVersion == id {
			return nil
		}
		v.Revision++
		if err := r.PutVersion(v); err != nil {
			return err
		}
		old, err := findVersion(r, site.DefaultVersion)
		if err != nil {
			return err
		}
		mode.DefaultVersion = id
		mode.Revision++
		if err = r.PutMode(mode); err != nil {
			return err
		}
		if old.ModeID == mode.ID {
			site.DefaultVersion = id
			site.Revision++
			return r.PutSite(*site)
		}
		return nil
	})
}
func (s *ManualService) DeleteVersion(id string, revision uint64) error {
	return s.changeVersion(id, revision, func(r productrepository.ProductDocsRepository, site *productdomain.DocSite, mode productdomain.DocMode, v productdomain.DocVersion) error {
		if v.Status != "draft" || id == site.DefaultVersion || id == mode.DefaultVersion {
			return invalidDoc("只能删除非默认的草稿版本")
		}
		groups, err := r.Groups(id)
		if err != nil {
			return err
		}
		if len(groups) > 0 {
			return fmt.Errorf("%w: 版本内仍有分组", ErrWebsiteDocsConflict)
		}
		docs, err := r.Documents(id, false)
		if err != nil {
			return err
		}
		if len(docs) > 0 {
			return fmt.Errorf("%w: 版本内仍有文章", ErrWebsiteDocsConflict)
		}
		return r.DeleteVersion(id)
	})
}
