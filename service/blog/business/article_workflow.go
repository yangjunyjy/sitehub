package business

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	blogdomain "sitehub/service/blog/domain"
	blogrepository "sitehub/service/blog/model"
	blogentity "sitehub/service/blog/model/entity"
)

var assetPattern = regexp.MustCompile(`/v1/blog/assets\?id=([a-fA-F0-9-]{36})`)

func assetIDs(input blogdomain.ArticleInput) []string {
	result := []string{}
	seen := map[string]bool{}
	for _, m := range assetPattern.FindAllStringSubmatch(input.Content+"\n"+input.Cover, -1) {
		if !seen[m[1]] {
			result = append(result, m[1])
			seen[m[1]] = true
		}
	}
	return result
}
func articleError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return blogdomain.ErrArticleNotFound
	}
	var duplicate *mysqlDriver.MySQLError
	if errors.As(err, &duplicate) && duplicate.Number == 1062 {
		if strings.Contains(strings.ToLower(duplicate.Message), "slug") {
			return blogdomain.ErrArticleSlug
		}
		return blogdomain.ErrArticleConflict
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return blogdomain.ErrArticleConflict
	}
	// SQLite is used only by the isolated regression suite.
	if err != nil && strings.HasPrefix(err.Error(), "UNIQUE constraint failed: articles.slug") {
		return blogdomain.ErrArticleSlug
	}
	if err != nil && strings.HasPrefix(err.Error(), "UNIQUE constraint failed: article_operations.") {
		return blogdomain.ErrArticleConflict
	}
	return err
}
func normalizeArticle(input *blogdomain.ArticleInput) error {
	input.Title = strings.TrimSpace(input.Title)
	input.Slug = strings.TrimSpace(input.Slug)
	if input.ContentFormat == "" {
		input.ContentFormat = "markdown"
	}
	if input.ContentFormat != "markdown" || !utf8.ValidString(input.Content) || strings.ContainsRune(input.Content, 0) || utf8.RuneCountInString(input.Content) > 500000 || len(input.Content) > 2000000 {
		return blogdomain.ErrInvalidArticle
	}
	if utf8.RuneCountInString(input.Title) > 200 || utf8.RuneCountInString(input.Summary) > 500 || len(input.Cover) > 500 || len(input.TagIDs) > 100 {
		return blogdomain.ErrInvalidArticle
	}
	if input.Slug == "" {
		var b strings.Builder
		for _, r := range strings.ToLower(input.Title) {
			if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' {
				b.WriteRune(r)
			} else {
				b.WriteByte('-')
			}
		}
		input.Slug = strings.Trim(b.String(), "-")
		if input.Slug == "" {
			input.Slug = "draft"
		}
	}
	if utf8.RuneCountInString(input.Slug) > 200 || strings.ContainsAny(input.Slug, " /?#\\\t\n") {
		return blogdomain.ErrInvalidArticle
	}
	if input.Cover != "" {
		u, err := url.Parse(input.Cover)
		if err != nil || (!strings.HasPrefix(input.Cover, "/") && u.Scheme != "https" && u.Scheme != "http") || strings.HasPrefix(input.Cover, "//") {
			return blogdomain.ErrInvalidArticle
		}
	}
	ids := []uint{}
	seen := map[uint]bool{}
	for _, id := range input.TagIDs {
		if id == 0 {
			return blogdomain.ErrArticleReference
		}
		if !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	input.TagIDs = ids
	return nil
}
func snapshot(row *blogentity.Article) blogdomain.ArticleView {
	v := blogdomain.ArticleView{ID: row.ID, Version: row.Version, ArticleInput: blogdomain.ArticleInput{Title: row.Title, Slug: row.Slug, Summary: row.Summary, Content: row.Content, ContentFormat: row.ContentFormat, Cover: row.Cover, CategoryID: row.CategoryID, TagIDs: []uint{}}, Status: row.Status, IsTop: row.IsTop, Author: blogdomain.ArticleLabel{ID: row.CreatedBy, Name: row.AuthorName}, Tags: []blogdomain.ArticleLabel{}, CreatedBy: row.CreatedBy, UpdatedBy: row.UpdatedBy, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt, PublishedAt: row.PublishedAt, ScheduledAt: row.ScheduledAt, DeletedAt: row.DeletedAt, Views: row.Views, HasUnpublishedChanges: row.Status == 1 && row.Version > row.PublishedVersion}
	if row.Category != nil {
		v.Category = &blogdomain.ArticleLabel{ID: uint64(row.Category.ID), Name: row.Category.Name}
	}
	for _, tag := range row.Tags {
		v.Tags = append(v.Tags, blogdomain.ArticleLabel{ID: tag.ID, Name: tag.Name})
		v.TagIDs = append(v.TagIDs, uint(tag.ID))
	}
	return v
}
func applyInput(row *blogentity.Article, in blogdomain.ArticleInput) {
	row.Title = in.Title
	row.Slug = in.Slug
	row.Summary = in.Summary
	row.Content = in.Content
	row.ContentFormat = in.ContentFormat
	row.Cover = in.Cover
	row.CategoryID = in.CategoryID
}
func (s *articleService) EditorDetail(actor blogdomain.ArticleActor, id uint64) (blogdomain.ArticleView, error) {
	if actor.ID == 0 || !(actor.CanRead || actor.CanUpdate) {
		return blogdomain.ArticleView{}, blogdomain.ErrArticleForbidden
	}
	row, err := s.articleRepo.WorkflowGet(id, false)
	if err != nil {
		return blogdomain.ArticleView{}, articleError(err)
	}
	return editorView(s.articleRepo, row)
}
func (s *articleService) PublicDetail(id uint64) (blogdomain.ArticleView, error) {
	row, err := s.articleRepo.WorkflowGet(id, false)
	if err != nil {
		return blogdomain.ArticleView{}, articleError(err)
	}
	return publishedView(s.articleRepo, row)
}
func publishedView(repo blogrepository.ArticleRepository, row *blogentity.Article) (blogdomain.ArticleView, error) {
	if row.DeletedAt != nil || row.Status != 1 {
		return blogdomain.ArticleView{}, blogdomain.ErrArticleNotFound
	}
	v := snapshot(row)
	publication, err := repo.Publication(row.ID)
	if err == nil {
		if err = json.Unmarshal([]byte(publication.Payload), &v); err != nil {
			return v, err
		}
	} else {
		return v, articleError(err)
	}
	v.CreatedBy = 0
	v.UpdatedBy = 0
	v.ScheduledAt = nil
	v.DeletedAt = nil
	v.HasUnpublishedChanges = false
	v.Status = 1
	v.IsTop = row.IsTop
	v.Views = row.Views
	return v, nil
}

// Public browsing always uses published snapshots and ignores editor-only filters.
func (s *articleService) BrowsePublic(q blogdomain.ArticleListQuery) (blogdomain.ArticlePage, error) {
	q.Trash = false
	q.Status = nil
	return s.browse(blogdomain.ArticleActor{}, q, true)
}
func (s *articleService) BrowseEditor(actor blogdomain.ArticleActor, q blogdomain.ArticleListQuery) (blogdomain.ArticlePage, error) {
	return s.browse(actor, q, false)
}
func (s *articleService) browse(actor blogdomain.ArticleActor, q blogdomain.ArticleListQuery, public bool) (blogdomain.ArticlePage, error) {
	if q.Page == 0 {
		q.Page = 1
	}
	if q.PageSize == 0 {
		q.PageSize = 20
	}
	result := blogdomain.ArticlePage{Items: []blogdomain.ArticleView{}, Page: q.Page, PageSize: q.PageSize}
	if !public && (actor.ID == 0 || !actor.CanRead) {
		return result, blogdomain.ErrArticleForbidden
	}
	if q.Page < 1 || q.PageSize < 1 || q.PageSize > 50 {
		return result, blogdomain.ErrInvalidArticle
	}
	rows, total, err := s.articleRepo.WorkflowList(q.Page, q.PageSize, q.Title, q.Sort, q.Status, q.CategoryID, q.TagIDs, q.Trash, public)
	if err != nil {
		return result, err
	}
	result.Total = total
	for _, row := range rows {
		v := snapshot(row)
		if public {
			v, err = publishedView(s.articleRepo, row)
		} else {
			v, err = editorView(s.articleRepo, row)
		}
		if err != nil {
			return result, err
		}
		v.Content = ""
		result.Items = append(result.Items, v)
	}
	return result, nil
}
func (s *articleService) EditorOptions(actor blogdomain.ArticleActor) (blogdomain.ArticleOptions, error) {
	out := blogdomain.ArticleOptions{Categories: []blogdomain.ArticleLabel{}, Tags: []blogdomain.ArticleLabel{}}
	if actor.ID == 0 || !(actor.CanCreate || actor.CanUpdate || actor.CanRead) {
		return out, blogdomain.ErrArticleForbidden
	}
	categories, tags, err := s.articleRepo.Options()
	if err != nil {
		return out, err
	}
	for _, v := range categories {
		out.Categories = append(out.Categories, blogdomain.ArticleLabel{ID: uint64(v.ID), Name: v.Name})
	}
	for _, v := range tags {
		out.Tags = append(out.Tags, blogdomain.ArticleLabel{ID: v.ID, Name: v.Name})
	}
	return out, nil
}

// A saved request result shares the same transaction as content and audit. Retrying an uncertain response is safe.
func (s *articleService) editorMutation(actor blogdomain.ArticleActor, key string, input any, fn func(blogrepository.ArticleRepository) (blogdomain.ArticleView, error)) (blogdomain.ArticleView, error) {
	var out blogdomain.ArticleView
	if actor.ID == 0 || len(key) < 8 || len(key) > 80 {
		return out, blogdomain.ErrInvalidArticle
	}
	raw, err := json.Marshal(input)
	if err != nil {
		return out, err
	}
	digest := fmt.Sprintf("%x", sha256.Sum256(raw))
	err = s.articleRepo.WorkflowTransaction(func(repo blogrepository.ArticleRepository) error {
		previous, e := repo.FindOperation(actor.ID, key)
		if e == nil {
			if previous.Digest != digest {
				return blogdomain.ErrArticleConflict
			}
			return json.Unmarshal([]byte(previous.Result), &out)
		}
		if !errors.Is(e, gorm.ErrRecordNotFound) {
			return e
		}
		out, e = fn(repo)
		if e != nil {
			return e
		}
		data, e := json.Marshal(out)
		if e != nil {
			return e
		}
		return repo.StoreOperation(&blogentity.ArticleOperation{ActorID: actor.ID, RequestID: key, Digest: digest, Result: string(data), CreatedAt: time.Now()})
	})
	if err != nil { // A concurrent identical request may have committed while this transaction waited.
		previous, e := s.articleRepo.FindOperation(actor.ID, key)
		if e == nil && previous.Digest == digest {
			if e = json.Unmarshal([]byte(previous.Result), &out); e == nil {
				return out, nil
			}
		}
	}
	return out, articleError(err)
}
func (s *articleService) SaveDraft(actor blogdomain.ArticleActor, cmd blogdomain.SaveArticle) (blogdomain.ArticleView, error) {
	if actor.ID == 0 || (cmd.ID == 0 && !actor.CanCreate) || (cmd.ID != 0 && !actor.CanUpdate) {
		return blogdomain.ArticleView{}, blogdomain.ErrArticleForbidden
	}
	automaticSlug := strings.TrimSpace(cmd.Input.Slug) == ""
	if err := normalizeArticle(&cmd.Input); err != nil {
		return blogdomain.ArticleView{}, err
	}
	if automaticSlug {
		hash := fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%d:%s", actor.ID, cmd.RequestID))))
		base := []rune(cmd.Input.Slug)
		if len(base) > 180 {
			base = base[:180]
		}
		cmd.Input.Slug = string(base) + "-" + hash[:8]
	}
	return s.editorMutation(actor, cmd.RequestID, cmd, func(repo blogrepository.ArticleRepository) (blogdomain.ArticleView, error) {
		now := time.Now()
		row := &blogentity.Article{Version: 0, CreatedAt: now, CreatedBy: actor.ID, AuthorName: actor.Name}
		if cmd.ID != 0 {
			existing, err := repo.WorkflowGet(cmd.ID, true)
			if err != nil {
				return blogdomain.ArticleView{}, err
			}
			row = existing
			if row.DeletedAt != nil {
				return blogdomain.ArticleView{}, blogdomain.ErrArticleState
			}
			if row.Version != cmd.Version {
				return blogdomain.ArticleView{}, blogdomain.ErrArticleConflict
			}
		}
		assets := assetIDs(cmd.Input)
		if err := repo.ValidateReferences(cmd.Input.CategoryID, cmd.Input.TagIDs, assets, actor.ID, row.ID); err != nil {
			return blogdomain.ArticleView{}, err
		}
		if row.ID != 0 {
			if err := ensureBaseline(repo, row, actor.ID); err != nil {
				return blogdomain.ArticleView{}, err
			}
			if row.Status == 2 && row.ScheduledVersion == 0 {
				return blogdomain.ArticleView{}, blogdomain.ErrArticleState
			}
		}
		applyInput(row, cmd.Input)
		row.Version++
		row.UpdatedBy = actor.ID
		row.UpdatedAt = now
		if err := repo.WorkflowSave(row); err != nil {
			return blogdomain.ArticleView{}, err
		}
		if err := repo.DeleteArticleTags(uint(row.ID)); err != nil {
			return blogdomain.ArticleView{}, err
		}
		if err := repo.SaveArticleTags(uint(row.ID), cmd.Input.TagIDs); err != nil {
			return blogdomain.ArticleView{}, err
		}
		if err := repo.SaveAssetReferences(row.ID, row.Version, assets); err != nil {
			return blogdomain.ArticleView{}, err
		}
		loaded, err := repo.WorkflowGet(row.ID, false)
		if err != nil {
			return blogdomain.ArticleView{}, err
		}
		return recordRevision(repo, loaded, actor.ID, "save")
	})
}
func ensureBaseline(repo blogrepository.ArticleRepository, row *blogentity.Article, actor uint64) error {
	history, err := repo.Revisions(row.ID)
	if err != nil {
		return err
	}
	if len(history) > 0 {
		return nil
	}
	return blogdomain.ErrArticleState
}
func recordRevision(repo blogrepository.ArticleRepository, row *blogentity.Article, actor uint64, action string) (blogdomain.ArticleView, error) {
	view := snapshot(row)
	raw, err := json.Marshal(view)
	if err != nil {
		return view, err
	}
	err = repo.SaveRevision(&blogentity.ArticleRevision{ArticleID: row.ID, Version: row.Version, Payload: string(raw), ActorID: actor, Action: action, CreatedAt: time.Now()})
	if err != nil {
		return view, err
	}
	return editorView(repo, row)
}
func writePublication(repo blogrepository.ArticleRepository, row *blogentity.Article) error {
	view := snapshot(row)
	view.Status = 1
	view.HasUnpublishedChanges = false
	raw, err := json.Marshal(view)
	if err != nil {
		return err
	}
	when := time.Now()
	if row.PublishedAt != nil {
		when = *row.PublishedAt
	}
	return repo.Publish(&blogentity.ArticlePublication{ArticleID: row.ID, Title: row.Title, Slug: row.Slug, Content: row.Content, CategoryID: row.CategoryID, Payload: string(raw), PublishedAt: when}, view.TagIDs)
}
func (s *articleService) ActOnArticle(actor blogdomain.ArticleActor, cmd blogdomain.ArticleAction) (blogdomain.ArticleView, error) {
	allowed := false
	switch cmd.Action {
	case "publish", "schedule", "cancel-schedule", "unpublish":
		allowed = actor.CanPublish
	case "top":
		allowed = actor.CanAudit
	case "delete", "restore":
		allowed = actor.CanDelete
	case "restore-revision":
		allowed = actor.CanUpdate
	}
	if !allowed || actor.ID == 0 {
		return blogdomain.ArticleView{}, blogdomain.ErrArticleForbidden
	}
	if cmd.ID == 0 || cmd.Version == 0 {
		return blogdomain.ArticleView{}, blogdomain.ErrInvalidArticle
	}
	return s.editorMutation(actor, cmd.RequestID, cmd, func(repo blogrepository.ArticleRepository) (blogdomain.ArticleView, error) {
		row, err := repo.WorkflowGet(cmd.ID, true)
		if err != nil {
			return blogdomain.ArticleView{}, err
		}
		if row.Version != cmd.Version {
			return blogdomain.ArticleView{}, blogdomain.ErrArticleConflict
		}
		if row.DeletedAt != nil && cmd.Action != "restore" {
			return blogdomain.ArticleView{}, blogdomain.ErrArticleState
		}
		if err = ensureBaseline(repo, row, actor.ID); err != nil {
			return blogdomain.ArticleView{}, err
		}
		now := time.Now()
		if cmd.Action == "publish" || cmd.Action == "schedule" {
			if row.ContentFormat != "markdown" || strings.TrimSpace(row.Title) == "" || !hasArticleContent(row.Content) {
				return blogdomain.ArticleView{}, blogdomain.ErrInvalidArticle
			}
			v := snapshot(row)
			if err = repo.ValidateReferences(row.CategoryID, v.TagIDs, assetIDs(v.ArticleInput), actor.ID, row.ID); err != nil {
				return blogdomain.ArticleView{}, err
			}
		}
		switch cmd.Action {
		case "publish":
			row.Status = 1
			row.PublishedAt = &now
			row.ScheduledAt = nil
			row.ScheduledVersion = 0
			row.PublishedVersion = row.Version + 1
		case "schedule":
			if row.Status == 1 || cmd.ScheduledAt == nil || !cmd.ScheduledAt.After(now.Add(time.Minute)) || cmd.ScheduledAt.After(now.AddDate(1, 0, 0)) {
				return blogdomain.ArticleView{}, blogdomain.ErrInvalidArticle
			}
			row.Status = 2
			row.ScheduledAt = cmd.ScheduledAt
			row.ScheduledVersion = row.Version + 1
		case "cancel-schedule":
			if row.Status != 2 {
				return blogdomain.ArticleView{}, blogdomain.ErrArticleState
			}
			row.Status = 0
			row.ScheduledAt = nil
			row.ScheduledVersion = 0
		case "unpublish":
			if row.Status != 1 {
				return blogdomain.ArticleView{}, blogdomain.ErrArticleState
			}
			row.Status = 0
		case "top":
			if cmd.IsTop == nil {
				return blogdomain.ArticleView{}, blogdomain.ErrInvalidArticle
			}
			row.IsTop = *cmd.IsTop
		case "delete":
			row.DeletedAt = &now
			row.Status = 0
			row.ScheduledAt = nil
			row.ScheduledVersion = 0
		case "restore":
			if row.DeletedAt == nil {
				return blogdomain.ArticleView{}, blogdomain.ErrArticleState
			}
			row.DeletedAt = nil
			row.Status = 0
		case "restore-revision":
			rev, e := repo.Revision(row.ID, cmd.RevisionID)
			if e != nil {
				return blogdomain.ArticleView{}, e
			}
			var old blogdomain.ArticleView
			if e = json.Unmarshal([]byte(rev.Payload), &old); e != nil {
				return blogdomain.ArticleView{}, e
			}
			if old.ContentFormat != "markdown" {
				return blogdomain.ArticleView{}, blogdomain.ErrInvalidArticle
			}
			if e = repo.ValidateReferences(old.CategoryID, old.TagIDs, assetIDs(old.ArticleInput), actor.ID, row.ID); e != nil {
				return blogdomain.ArticleView{}, e
			}
			applyInput(row, old.ArticleInput)
			if e = repo.DeleteArticleTags(uint(row.ID)); e != nil {
				return blogdomain.ArticleView{}, e
			}
			if e = repo.SaveArticleTags(uint(row.ID), old.TagIDs); e != nil {
				return blogdomain.ArticleView{}, e
			}
		}
		row.Version++
		row.UpdatedBy = actor.ID
		row.UpdatedAt = now
		if err = repo.WorkflowSave(row); err != nil {
			return blogdomain.ArticleView{}, err
		}
		row, err = repo.WorkflowGet(row.ID, false)
		if err != nil {
			return blogdomain.ArticleView{}, err
		}
		if cmd.Action == "publish" {
			if err = writePublication(repo, row); err != nil {
				return blogdomain.ArticleView{}, err
			}
		}
		if err = repo.SaveAssetReferences(row.ID, row.Version, assetIDs(snapshot(row).ArticleInput)); err != nil {
			return blogdomain.ArticleView{}, err
		}
		return recordRevision(repo, row, actor.ID, cmd.Action)
	})
}
func (s *articleService) History(actor blogdomain.ArticleActor, id uint64) ([]blogdomain.ArticleHistory, error) {
	if _, err := s.EditorDetail(actor, id); err != nil {
		return nil, err
	}
	rows, err := s.articleRepo.Revisions(id)
	out := []blogdomain.ArticleHistory{}
	for _, r := range rows {
		out = append(out, blogdomain.ArticleHistory{ID: r.ID, Version: r.Version, ActorID: r.ActorID, Action: r.Action, CreatedAt: r.CreatedAt})
	}
	return out, err
}
func (s *articleService) RevisionDetail(actor blogdomain.ArticleActor, id, revisionID uint64) (blogdomain.ArticleView, error) {
	if _, err := s.EditorDetail(actor, id); err != nil {
		return blogdomain.ArticleView{}, err
	}
	row, err := s.articleRepo.Revision(id, revisionID)
	if err != nil {
		return blogdomain.ArticleView{}, articleError(err)
	}
	var out blogdomain.ArticleView
	err = json.Unmarshal([]byte(row.Payload), &out)
	return out, err
}

func editorView(repo blogrepository.ArticleRepository, row *blogentity.Article) (blogdomain.ArticleView, error) {
	view := snapshot(row)
	if row.Status == 1 {
		publication, err := repo.Publication(row.ID)
		if err == nil {
			var published blogdomain.ArticleView
			if err = json.Unmarshal([]byte(publication.Payload), &published); err != nil {
				return view, err
			}
			view.HasUnpublishedChanges = !reflect.DeepEqual(view.ArticleInput, published.ArticleInput)
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return view, err
		}
	}
	return view, nil
}

var htmlOnlyTags = regexp.MustCompile(`<[^>]*>`)

func hasArticleContent(source string) bool {
	return strings.Trim(htmlOnlyTags.ReplaceAllString(source, ""), "#*_~`>-| \r\n\t") != ""
}
