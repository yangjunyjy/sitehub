package services

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	productcontract "sitehub/service/product/contract"
	productdomain "sitehub/service/product/domain"
	productrepository "sitehub/service/product/model"
)

var ErrWebsiteDocumentNotFound = productdomain.ErrDocumentNotFound
var ErrWebsiteDocsConflict = productdomain.ErrDocumentConflict
var ErrInvalidWebsiteDocs = productdomain.ErrInvalidDocument
var docVersionID = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]{0,79}$`)
var documentSlug = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,199}$`)

func invalidDoc(message string) error { return fmt.Errorf("%w: %s", ErrInvalidWebsiteDocs, message) }
func documentView(d productdomain.ProductDocument) productcontract.WebsiteDocDocument {
	return productcontract.WebsiteDocDocument{ID: d.DocKey, Slug: d.Slug, Title: d.Title, Summary: d.Summary, Content: d.ContentMarkdown, Revision: d.Revision, GroupID: d.GroupID, UpdatedAt: d.UpdatedAt}
}

// 管理导航只查询元数据。正文通过单篇接口读取，禁止整树回写。
func (s *websiteContentService) GetDocs() (result productcontract.WebsiteDocs, err error) {
	err = s.docs.Transaction(func(r productrepository.ProductDocsRepository) error {
		site, e := r.Site(false)
		if e != nil {
			return e
		}
		versions, e := r.Versions()
		if e != nil {
			return e
		}
		modes, e := r.Modes()
		if e != nil {
			return e
		}
		groups, e := r.Groups("")
		if e != nil {
			return e
		}
		docs, e := r.Documents("", false)
		if e != nil {
			return e
		}
		result = productcontract.WebsiteDocs{Title: site.Title, Description: site.Description, DefaultVersion: site.DefaultVersion, Revision: site.Revision, Versions: []productcontract.WebsiteDocVersion{}}
		result.Modes = []productcontract.WebsiteDocMode{}
		result.DefaultMode = defaultMode(versions, site.DefaultVersion)
		for _, m := range modes {
			result.Modes = append(result.Modes, modeView(m))
		}
		for _, v := range versions {
			entry := productcontract.WebsiteDocVersion{ID: v.ID, Name: v.Name, Status: v.Status, Revision: v.Revision, Groups: []productcontract.WebsiteDocGroup{}}
			entry.ModeID = v.ModeID
			for _, g := range groups {
				if g.VersionID != v.ID {
					continue
				}
				group := productcontract.WebsiteDocGroup{ID: g.ID, Revision: g.Revision, Title: g.Title, Items: []productcontract.WebsiteDocDocument{}}
				for _, d := range docs {
					if d.GroupID == g.ID {
						group.Items = append(group.Items, documentView(d))
					}
				}
				entry.Groups = append(entry.Groups, group)
			}
			result.Versions = append(result.Versions, entry)
		}
		return nil
	})
	return
}
func (s *websiteContentService) GetDocsNavigation(version string, requestedMode ...string) (productcontract.WebsiteDocsNavigation, error) {
	docs, err := s.GetDocs()
	if err != nil {
		return productcontract.WebsiteDocsNavigation{}, err
	}
	mode := docs.DefaultMode
	if version != "" {
		for _, v := range docs.Versions {
			if v.ID == version {
				mode = v.ModeID
			}
		}
	}
	if len(requestedMode) > 0 && requestedMode[0] != "" {
		mode = requestedMode[0]
	}
	var selected *productcontract.WebsiteDocMode
	for i := range docs.Modes {
		if docs.Modes[i].ID == mode {
			selected = &docs.Modes[i]
		}
	}
	if selected == nil || !selected.Available {
		return productcontract.WebsiteDocsNavigation{}, ErrWebsiteDocumentNotFound
	}
	if version == "" {
		version = selected.DefaultVersion
	}
	nav := productcontract.WebsiteDocsNavigation{Title: docs.Title, Description: docs.Description, VersionID: version, DefaultVersion: docs.DefaultVersion, Versions: []productcontract.WebsiteDocVersionSummary{}, Groups: []productcontract.WebsiteDocNavigationGroup{}}
	nav.DefaultVersion = selected.DefaultVersion
	found := false
	for _, v := range docs.Versions {
		if v.Status != "published" || v.ModeID != mode {
			continue
		}
		nav.Versions = append(nav.Versions, productcontract.WebsiteDocVersionSummary{ID: v.ID, Name: v.Name})
		if v.ID != version {
			continue
		}
		found = true
		nav.Version = v.Name
		for _, g := range v.Groups {
			entry := productcontract.WebsiteDocNavigationGroup{Title: g.Title, Items: []productcontract.WebsiteDocNavigationDocument{}}
			for _, d := range g.Items {
				entry.Items = append(entry.Items, productcontract.WebsiteDocNavigationDocument{ID: d.ID, Slug: d.Slug, Title: d.Title, Summary: d.Summary})
			}
			nav.Groups = append(nav.Groups, entry)
		}
	}
	if !found {
		return nav, ErrWebsiteDocumentNotFound
	}
	return nav, nil
}
func findVersion(r productrepository.ProductDocsRepository, id string) (productdomain.DocVersion, error) {
	versions, err := r.Versions()
	if err != nil {
		return productdomain.DocVersion{}, err
	}
	for _, v := range versions {
		if v.ID == id {
			return v, nil
		}
	}
	return productdomain.DocVersion{}, ErrWebsiteDocumentNotFound
}
func (s *websiteContentService) readDocument(id, version string, public bool, mode string) (result productcontract.WebsiteDocDocument, err error) {
	err = s.docs.Transaction(func(r productrepository.ProductDocsRepository) error {
		site, e := r.Site(false)
		if e != nil {
			return e
		}
		if public {
			v, e := publicDocVersion(r, site, version, mode)
			if e != nil {
				return e
			}
			version = v.ID
		} else if version == "" {
			version = site.DefaultVersion
		}
		v, e := findVersion(r, version)
		if e != nil {
			return e
		}
		if public && v.Status != "published" {
			return ErrWebsiteDocumentNotFound
		}
		d, e := r.Document(version, id)
		if e != nil {
			return e
		}
		result = documentView(d)
		return nil
	})
	return
}
func (s *websiteContentService) GetDocument(id, version string, requestedMode ...string) (productcontract.WebsiteDocDocument, error) {
	mode := ""
	if len(requestedMode) > 0 {
		mode = requestedMode[0]
	}
	return s.readDocument(id, version, true, mode)
}
func (s *websiteContentService) GetAdminDocument(id, version string) (productcontract.WebsiteDocDocument, error) {
	return s.readDocument(id, version, false, "")
}

// 所有关系写入先锁文档站行，保护版本/分组删除与新增引用。乐观版本号按资源检查，
// 修改另一篇文档不会导致当前文档保存冲突。事务回调仅使用同一连接的仓储。
func (s *websiteContentService) mutateDocs(fn func(productrepository.ProductDocsRepository, *productdomain.DocSite) error) error {
	return s.docs.Transaction(func(r productrepository.ProductDocsRepository) error {
		site, err := r.Site(true)
		if err != nil {
			return err
		}
		return fn(r, &site)
	})
}
func (s *websiteContentService) SaveDocSettings(in productdomain.DocSettingsInput) error {
	if strings.TrimSpace(in.Title) == "" || len([]rune(in.Title)) > 100 || len([]rune(in.Description)) > 500 {
		return invalidDoc("请填写有效的文档站标题和介绍")
	}
	return s.mutateDocs(func(r productrepository.ProductDocsRepository, site *productdomain.DocSite) error {
		if in.Revision != site.Revision {
			return ErrWebsiteDocsConflict
		}
		site.Title = strings.TrimSpace(in.Title)
		site.Description = in.Description
		site.Revision++
		return r.PutSite(*site)
	})
}
func snapshotDocument(r productrepository.ProductDocsRepository, d productdomain.ProductDocument, group, operation string, actor uint64) error {
	return r.AddRevision(&productdomain.DocumentRevision{DocumentID: d.ID, Revision: d.Revision, ActorID: actor, VersionID: d.VersionID, DocKey: d.DocKey, Slug: d.Slug, Title: d.Title, Summary: d.Summary, ContentMarkdown: d.ContentMarkdown, GroupTitle: group, Operation: operation, CreatedAt: time.Now()})
}
func (s *websiteContentService) ChangeDocVersion(in productdomain.DocVersionInput, actor uint64) error {
	if !docVersionID.MatchString(in.ID) {
		return invalidDoc("版本标识无效")
	}
	return s.mutateDocs(func(r productrepository.ProductDocsRepository, site *productdomain.DocSite) error {
		versions, err := r.Versions()
		if err != nil {
			return err
		}
		var current *productdomain.DocVersion
		targetMode := in.ModeID
		if targetMode == "" {
			for _, v := range versions {
				if v.ID == in.ID {
					targetMode = v.ModeID
				}
			}
			if targetMode == "" {
				targetMode = defaultMode(versions, site.DefaultVersion)
			}
		}
		for i := range versions {
			v := &versions[i]
			if v.ID == in.ID {
				current = v
			}
			if in.Action == "save" && v.ID != in.ID && v.ModeID == targetMode && v.Name == strings.TrimSpace(in.Name) {
				return invalidDoc("版本名称已存在")
			}
		}
		if in.Action == "save" {
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
			if in.ModeID == "" {
				in.ModeID = defaultMode(versions, site.DefaultVersion)
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
		}
		if current == nil {
			return ErrWebsiteDocumentNotFound
		}
		if current.Revision != in.Revision {
			return ErrWebsiteDocsConflict
		}
		mode, err := findMode(r, current.ModeID)
		if err != nil {
			return err
		}
		switch in.Action {
		case "status":
			if in.Status != "draft" && in.Status != "published" {
				return invalidDoc("版本状态无效")
			}
			if in.Status == "draft" && (site.DefaultVersion == in.ID || mode.DefaultVersion == in.ID) {
				return invalidDoc("请先切换默认版本")
			}
			if in.Status == "published" {
				docs, e := r.Documents(in.ID, false)
				if e != nil {
					return e
				}
				if len(docs) == 0 {
					return invalidDoc("空版本不能发布")
				}
			}
			current.Status = in.Status
			current.Revision++
			return r.PutVersion(*current)
		case "default":
			if current.Status != "published" {
				return invalidDoc("默认版本必须已发布")
			}
			mode.DefaultVersion = in.ID
			mode.Revision++
			if err = r.PutMode(mode); err != nil {
				return err
			}
			if defaultMode(versions, site.DefaultVersion) == mode.ID {
				site.DefaultVersion = in.ID
				site.Revision++
				return r.PutSite(*site)
			}
			return nil
		case "delete":
			if current.Status != "draft" || site.DefaultVersion == in.ID || mode.DefaultVersion == in.ID {
				return invalidDoc("只能删除非默认的草稿版本")
			}
			docs, e := r.Documents(in.ID, true)
			if e != nil {
				return e
			}
			groups, e := r.Groups(in.ID)
			if e != nil {
				return e
			}
			for _, d := range docs {
				d.Revision++
				if err = snapshotDocument(r, d, "", "delete", actor); err != nil {
					return err
				}
				if err = r.DeleteDocument(d.ID); err != nil {
					return err
				}
			}
			for _, g := range groups {
				if err = r.DeleteGroup(g.ID); err != nil {
					return err
				}
			}
			return r.DeleteVersion(in.ID)
		}
		return invalidDoc("版本操作无效")
	})
}
func (s *websiteContentService) ChangeDocGroup(in productdomain.DocGroupInput) error {
	return s.mutateDocs(func(r productrepository.ProductDocsRepository, site *productdomain.DocSite) error {
		v, err := findVersion(r, in.Version)
		if err != nil {
			return err
		}
		groups, err := r.Groups(in.Version)
		if err != nil {
			return err
		}
		index := -1
		for i, g := range groups {
			if g.ID == in.ID {
				index = i
			}
			if in.Action == "save" && g.ID != in.ID && g.Title == strings.TrimSpace(in.Title) {
				return invalidDoc("分组名称已存在")
			}
		}
		var g productdomain.DocGroup
		if in.ID != 0 {
			if index < 0 {
				return ErrWebsiteDocumentNotFound
			}
			g = groups[index]
			if g.Revision != in.Revision {
				return ErrWebsiteDocsConflict
			}
		} else if in.Action != "save" || in.Revision != 0 {
			return invalidDoc("缺少分组 ID")
		}
		switch in.Action {
		case "save":
			if strings.TrimSpace(in.Title) == "" || len([]rune(in.Title)) > 60 {
				return invalidDoc("请填写有效的分组名称")
			}
			g.Title = strings.TrimSpace(in.Title)
			g.VersionID = in.Version
			g.Revision++
			if in.ID == 0 {
				g.SortOrder = len(groups)
				if len(groups) > 0 {
					g.SortOrder = groups[len(groups)-1].SortOrder + 1
				}
			}
			err = r.PutGroup(&g)
		case "move":
			if in.Direction != 1 && in.Direction != -1 {
				return invalidDoc("排序方向无效")
			}
			other := index + in.Direction
			if other < 0 || other >= len(groups) {
				return invalidDoc("分组已经位于边界")
			}
			groups[index], groups[other] = groups[other], groups[index]
			for i := range groups {
				groups[i].SortOrder = i
				groups[i].Revision++
				if err = r.PutGroup(&groups[i]); err != nil {
					return err
				}
			}
		case "delete":
			docs, e := r.Documents(in.Version, false)
			if e != nil {
				return e
			}
			for _, d := range docs {
				if d.GroupID == g.ID {
					return fmt.Errorf("%w: 分组内仍有文档", ErrWebsiteDocsConflict)
				}
			}
			err = r.DeleteGroup(g.ID)
		default:
			return invalidDoc("分组操作无效")
		}
		if err != nil {
			return err
		}
		v.Revision++
		return r.PutVersion(v)
	})
}
func (s *websiteContentService) SaveDocument(in productdomain.DocumentInput, actor uint64) (saved productcontract.WebsiteDocDocument, err error) {
	in.Slug = strings.TrimSpace(in.Slug)
	if !documentSlug.MatchString(in.ID) || !documentSlug.MatchString(in.Slug) || strings.TrimSpace(in.Title) == "" || len([]rune(in.Title)) > 200 || len([]rune(in.Summary)) > 2000 || strings.TrimSpace(in.Content) == "" || len(in.Content) > 2*1024*1024 || in.GroupID == 0 {
		return saved, invalidDoc("请检查文档标题、标识、分组及正文（最多 2 MB）")
	}
	err = s.mutateDocs(func(r productrepository.ProductDocsRepository, site *productdomain.DocSite) error {
		v, e := findVersion(r, in.Version)
		if e != nil {
			return e
		}
		groups, e := r.Groups(in.Version)
		if e != nil {
			return e
		}
		groupTitle := ""
		for _, g := range groups {
			if g.ID == in.GroupID {
				groupTitle = g.Title
			}
		}
		if groupTitle == "" {
			return ErrWebsiteDocumentNotFound
		}
		docs, e := r.Documents(in.Version, false)
		if e != nil {
			return e
		}
		var d productdomain.ProductDocument
		for _, item := range docs {
			if item.DocKey == in.ID {
				d = item
				continue
			}
			// 仅同版本内其他现存文档的公开地址占用 slug，内部 ID 与历史修订不参与查重。
			if item.Slug == in.Slug {
				return invalidDoc(fmt.Sprintf("访问标识“%s”已被当前版本的文档“%s”使用", in.Slug, item.Title))
			}
		}
		if d.ID == 0 && in.Revision != 0 {
			return ErrWebsiteDocumentNotFound
		}
		if d.Revision != in.Revision {
			return ErrWebsiteDocsConflict
		}
		operation := "update"
		if d.ID == 0 {
			operation = "create"
			d.SortOrder = len(docs)
		}
		d.VersionID = in.Version
		d.GroupID = in.GroupID
		d.DocKey = in.ID
		d.Slug = in.Slug
		d.Title = strings.TrimSpace(in.Title)
		d.Summary = in.Summary
		d.ContentMarkdown = in.Content
		d.Revision++
		d.UpdatedBy = actor
		if e = r.PutDocument(&d); e != nil {
			return e
		}
		if e = snapshotDocument(r, d, groupTitle, operation, actor); e != nil {
			return e
		}
		v.Revision++
		if e = r.PutVersion(v); e != nil {
			return e
		}
		saved = documentView(d)
		return nil
	})
	return
}
func (s *websiteContentService) DeleteDocument(in productdomain.DocumentDeleteInput, actor uint64) error {
	return s.mutateDocs(func(r productrepository.ProductDocsRepository, site *productdomain.DocSite) error {
		v, e := findVersion(r, in.Version)
		if e != nil {
			return e
		}
		d, e := r.Document(in.Version, in.ID)
		if e != nil {
			return e
		}
		if d.Revision != in.Revision {
			return ErrWebsiteDocsConflict
		}
		if v.Status == "published" {
			docs, e := r.Documents(in.Version, false)
			if e != nil {
				return e
			}
			if len(docs) <= 1 {
				return invalidDoc("已发布版本至少保留一篇文档")
			}
		}
		d.Revision++
		if e = snapshotDocument(r, d, "", "delete", actor); e != nil {
			return e
		}
		if e = r.DeleteDocument(d.ID); e != nil {
			return e
		}
		v.Revision++
		return r.PutVersion(v)
	})
}
func (s *websiteContentService) DocumentHistory(id, version string, historyID uint64) (any, error) {
	var result any
	err := s.docs.Transaction(func(r productrepository.ProductDocsRepository) error {
		d, e := r.Document(version, id)
		if e != nil {
			return e
		}
		if historyID > 0 {
			v, e := r.Historical(d.ID, historyID)
			result = v
			return e
		}
		v, e := r.History(d.ID)
		if v == nil {
			v = []productdomain.DocumentRevision{}
		}
		result = v
		return e
	})
	return result, err
}
