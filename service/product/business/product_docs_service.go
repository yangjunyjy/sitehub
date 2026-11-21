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

type ManualService struct {
	docs productrepository.ProductDocsRepository
}

func NewManualService(repo productrepository.ProductDocsRepository) *ManualService {
	return &ManualService{docs: repo}
}

var ErrWebsiteDocumentNotFound = productdomain.ErrDocumentNotFound
var ErrWebsiteDocsConflict = productdomain.ErrDocumentConflict
var ErrInvalidWebsiteDocs = productdomain.ErrInvalidDocument
var docVersionID = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]{0,79}$`)
var documentSlug = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,199}$`)

func invalidDoc(message string) error { return fmt.Errorf("%w: %s", ErrInvalidWebsiteDocs, message) }
func documentSummary(d productdomain.ProductDocument) productcontract.ManualDocumentSummary {
	status := "draft"
	if d.PublishedRevision > 0 {
		status = "published"
		if d.PublishedRevision != d.Revision {
			status = "modified"
		}
	}
	return productcontract.ManualDocumentSummary{Status: status, PublishedRevision: d.PublishedRevision, ID: d.DocKey, Slug: d.Slug, Title: d.Title, Summary: d.Summary, Revision: d.Revision, GroupID: d.GroupID, UpdatedAt: d.UpdatedAt}
}

func documentView(d productdomain.ProductDocument) productcontract.WebsiteDocDocument {
	return productcontract.WebsiteDocDocument{ManualDocumentSummary: documentSummary(d), Content: d.ContentMarkdown}
}

// Workspace reads only the selected version's navigation. Opening an editor or
// a history row issues a separate request for that document's Markdown.
func (s *ManualService) Workspace(modeID, versionID string) (result productcontract.ManualWorkspace, err error) {
	err = s.docs.Transaction(func(r productrepository.ProductDocsRepository) error {
		site, err := r.Site(false)
		if err != nil {
			return err
		}
		versions, err := r.Versions()
		if err != nil {
			return err
		}
		modes, err := r.Modes()
		if err != nil {
			return err
		}
		result = productcontract.ManualWorkspace{Title: site.Title, Description: site.Description, Revision: site.Revision, DefaultVersion: site.DefaultVersion, DefaultMode: defaultMode(versions, site.DefaultVersion), Modes: []productcontract.WebsiteDocMode{}, Versions: []productcontract.WebsiteDocVersion{}, Groups: []productcontract.WebsiteDocGroup{}}
		if modeID == "" && versionID != "" {
			modeID = defaultMode(versions, versionID)
		}
		if modeID == "" {
			modeID = result.DefaultMode
		}
		if modeID == "" && len(modes) > 0 {
			modeID = modes[0].ID
		}
		for _, m := range modes {
			result.Modes = append(result.Modes, modeView(m))
			if m.ID == modeID {
				result.ModeID = m.ID
				if versionID == "" {
					versionID = m.DefaultVersion
				}
			}
		}
		if modeID != "" && result.ModeID == "" {
			return ErrWebsiteDocumentNotFound
		}
		for _, v := range versions {
			if v.ModeID != modeID {
				continue
			}
			result.Versions = append(result.Versions, productcontract.WebsiteDocVersion{ID: v.ID, Name: v.Name, ModeID: v.ModeID, Status: v.Status, Revision: v.Revision})
		}
		if versionID == "" && len(result.Versions) > 0 {
			versionID = result.Versions[0].ID
		}
		if versionID == "" {
			return nil
		}
		for _, v := range result.Versions {
			if v.ID == versionID {
				result.VersionID = v.ID
				break
			}
		}
		if result.VersionID == "" {
			return ErrWebsiteDocumentNotFound
		}
		groups, err := r.Groups(versionID)
		if err != nil {
			return err
		}
		docs, err := r.Documents(versionID, false)
		if err != nil {
			return err
		}
		items := make(map[uint64][]productcontract.ManualDocumentSummary)
		modified := make(map[uint64]productdomain.ProductDocument)
		for _, d := range docs {
			summary := documentSummary(d)
			items[d.GroupID] = append(items[d.GroupID], summary)
			if summary.Status == "modified" {
				modified[d.ID] = d
			}
		}
		// A moved draft can leave a published revision pointing to an apparently empty group.
		moved := make(map[uint64][]string)
		if len(modified) > 0 {
			live, err := r.PublishedDocuments(versionID)
			if err != nil {
				return err
			}
			for _, published := range live {
				if draft, ok := modified[published.ID]; ok && draft.GroupID != published.GroupID {
					moved[published.GroupID] = append(moved[published.GroupID], draft.Title)
				}
			}
		}
		for _, g := range groups {
			children := items[g.ID]
			if children == nil {
				children = []productcontract.ManualDocumentSummary{}
			}
			result.Groups = append(result.Groups, productcontract.WebsiteDocGroup{ID: g.ID, Title: g.Title, Revision: g.Revision, Items: children, MovedDraftTitles: moved[g.ID]})
		}
		return nil
	})
	return
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
func (s *ManualService) PublicDocument(id, version, mode string) (result productcontract.WebsiteDocDocument, err error) {
	err = s.docs.Transaction(func(r productrepository.ProductDocsRepository) error {
		if version == "" && mode == "" {
			site, err := r.Site(false)
			if err != nil {
				return err
			}
			version = site.DefaultVersion
		}
		selected, err := publicDocVersion(r, version, mode)
		if err != nil {
			return err
		}
		doc, err := r.PublishedDocument(selected.ID, id)
		if err != nil {
			return err
		}
		result = documentView(doc)
		return nil
	})
	return
}
func (s *ManualService) Draft(id, version string) (productcontract.WebsiteDocDocument, error) {
	d, err := s.docs.Document(version, id)
	return documentView(d), err
}

// 所有关系写入先锁文档站行，保护版本/分组删除与新增引用。乐观版本号按资源检查，
// 修改另一篇文档不会导致当前文档保存冲突。事务回调仅使用同一连接的仓储。
func (s *ManualService) mutateDocs(fn func(productrepository.ProductDocsRepository, *productdomain.DocSite) error) error {
	return s.docs.Transaction(func(r productrepository.ProductDocsRepository) error {
		site, err := r.Site(true)
		if err != nil {
			return err
		}
		return fn(r, &site)
	})
}
func (s *ManualService) SaveDocSettings(in productdomain.DocSettingsInput) error {
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
	return r.AddRevision(&productdomain.DocumentRevision{DocumentID: d.ID, Revision: d.Revision, ActorID: actor, VersionID: d.VersionID, DocKey: d.DocKey, Slug: d.Slug, Title: d.Title, Summary: d.Summary, ContentMarkdown: d.ContentMarkdown, GroupTitle: group, GroupID: d.GroupID, SortOrder: d.SortOrder, Operation: operation, CreatedAt: time.Now()})
}
func (s *ManualService) SaveDocument(in productdomain.DocumentInput, actor uint64) (saved productcontract.WebsiteDocDocument, err error) {
	in.Slug = strings.TrimSpace(in.Slug)
	if !documentSlug.MatchString(in.ID) || !documentSlug.MatchString(in.Slug) || strings.TrimSpace(in.Title) == "" || len([]rune(in.Title)) > 200 || len([]rune(in.Summary)) > 2000 || len(in.Content) > 2*1024*1024 || in.GroupID == 0 {
		return saved, invalidDoc("请检查文档标题、标识、分组及正文（最多 2 MB）")
	}
	err = s.mutateDocs(func(r productrepository.ProductDocsRepository, site *productdomain.DocSite) error {
		_, e := findVersion(r, in.Version)
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
		saved = documentView(d)
		return nil
	})
	return
}
func (s *ManualService) DeleteDocument(in productdomain.DocumentDeleteInput, actor uint64) error {
	return s.mutateDocs(func(r productrepository.ProductDocsRepository, site *productdomain.DocSite) error {
		_, e := findVersion(r, in.Version)
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
		if d.PublishedRevision > 0 {
			return fmt.Errorf("%w: 请先撤回已发布文章", ErrWebsiteDocsConflict)
		}
		d.Revision++
		if e = snapshotDocument(r, d, "", "delete", actor); e != nil {
			return e
		}
		if e = r.DeleteDocument(d.ID); e != nil {
			return e
		}
		return nil
	})
}
func (s *ManualService) History(version, key string) (items []productdomain.DocumentRevision, err error) {
	err = s.docs.Transaction(func(r productrepository.ProductDocsRepository) error {
		id, err := r.DocumentID(version, key)
		if err != nil {
			return err
		}
		items, err = r.History(id)
		if items == nil {
			items = []productdomain.DocumentRevision{}
		}
		return err
	})
	return
}

func (s *ManualService) Historical(version, key string, historyID uint64) (row productdomain.DocumentRevision, err error) {
	err = s.docs.Transaction(func(r productrepository.ProductDocsRepository) error {
		id, err := r.DocumentID(version, key)
		if err != nil {
			return err
		}
		row, err = r.Historical(id, historyID)
		return err
	})
	return
}

// SetDocumentPublished changes visibility and records an immutable revision in the
// same transaction. Later saves only update the draft and its revision history.
func (s *ManualService) SetDocumentPublished(in productdomain.DocumentDeleteInput, actor uint64, publish bool) (result productcontract.WebsiteDocDocument, err error) {
	err = s.mutateDocs(func(r productrepository.ProductDocsRepository, site *productdomain.DocSite) error {
		d, err := r.Document(in.Version, in.ID)
		if err != nil {
			return err
		}
		if d.Revision != in.Revision {
			return ErrWebsiteDocsConflict
		}
		operation := "withdraw"
		if publish {
			if strings.TrimSpace(d.ContentMarkdown) == "" {
				return invalidDoc("发布前请填写正文")
			}
			active, err := r.PublishedDocuments(in.Version)
			if err != nil {
				return err
			}
			for _, other := range active {
				if other.ID != d.ID && strings.EqualFold(other.Slug, d.Slug) {
					return invalidDoc("访问标识已被其他线上文章使用")
				}
			}
			d.PublishedRevision = d.Revision + 1
			operation = "publish"
		} else {
			if d.PublishedRevision == 0 {
				return invalidDoc("文章尚未发布")
			}
			d.PublishedRevision = 0
		}
		groups, err := r.Groups(in.Version)
		if err != nil {
			return err
		}
		groupTitle := ""
		for _, g := range groups {
			if g.ID == d.GroupID {
				groupTitle = g.Title
				break
			}
		}
		if groupTitle == "" {
			return ErrWebsiteDocumentNotFound
		}
		d.Revision++
		d.UpdatedBy = actor
		if err = r.PutDocument(&d); err != nil {
			return err
		}
		if err = snapshotDocument(r, d, groupTitle, operation, actor); err != nil {
			return err
		}
		result = documentView(d)
		return nil
	})
	return
}
