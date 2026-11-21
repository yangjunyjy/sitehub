package services

import (
	"fmt"
	"strings"

	productdomain "sitehub/service/product/domain"
	productrepository "sitehub/service/product/model"
)

func (s *ManualService) SaveGroup(in productdomain.DocGroupInput) error {
	return s.mutateDocs(func(r productrepository.ProductDocsRepository, site *productdomain.DocSite) error {
		if _, err := findVersion(r, in.Version); err != nil {
			return err
		}
		title := strings.TrimSpace(in.Title)
		if title == "" || len([]rune(title)) > 60 {
			return invalidDoc("请填写有效的分组名称")
		}
		groups, err := r.Groups(in.Version)
		if err != nil {
			return err
		}
		group := productdomain.DocGroup{VersionID: in.Version}
		for _, g := range groups {
			if g.ID == in.ID {
				group = g
			} else if g.Title == title {
				return invalidDoc("分组名称已存在")
			}
		}
		if in.ID != 0 && group.ID == 0 {
			return ErrWebsiteDocumentNotFound
		}
		if group.Revision != in.Revision {
			return ErrWebsiteDocsConflict
		}
		if group.ID == 0 && len(groups) > 0 {
			group.SortOrder = groups[len(groups)-1].SortOrder + 1
		}
		group.Title = title
		group.Revision++
		return r.PutGroup(&group)
	})
}
func (s *ManualService) MoveGroup(version string, id, revision uint64, direction int) error {
	if direction != 1 && direction != -1 {
		return invalidDoc("排序方向无效")
	}
	return s.mutateDocs(func(r productrepository.ProductDocsRepository, site *productdomain.DocSite) error {
		groups, err := r.Groups(version)
		if err != nil {
			return err
		}
		for i, g := range groups {
			if g.ID != id {
				continue
			}
			if g.Revision != revision {
				return ErrWebsiteDocsConflict
			}
			other := i + direction
			if other < 0 || other >= len(groups) {
				return invalidDoc("分组已经位于边界")
			}
			groups[i], groups[other] = groups[other], groups[i]
			for n := range groups {
				if groups[n].SortOrder != n {
					groups[n].SortOrder = n
					groups[n].Revision++
					if err = r.PutGroup(&groups[n]); err != nil {
						return err
					}
				}
			}
			return nil
		}
		return ErrWebsiteDocumentNotFound
	})
}
func (s *ManualService) DeleteGroup(version string, id, revision uint64) error {
	return s.mutateDocs(func(r productrepository.ProductDocsRepository, site *productdomain.DocSite) error {
		groups, err := r.Groups(version)
		if err != nil {
			return err
		}
		for _, g := range groups {
			if g.ID != id {
				continue
			}
			if g.Revision != revision {
				return ErrWebsiteDocsConflict
			}
			drafts, err := r.Documents(version, false)
			if err != nil {
				return err
			}
			published, err := r.PublishedDocuments(version)
			if err != nil {
				return err
			}
			for _, d := range append(drafts, published...) {
				if d.GroupID == id {
					return fmt.Errorf("%w: 分组内仍有草稿或线上文章", ErrWebsiteDocsConflict)
				}
			}
			return r.DeleteGroup(id)
		}
		return ErrWebsiteDocumentNotFound
	})
}
