package services

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"sitehub/service/product/domain"
	"sitehub/service/product/model/mysql"
	"sitehub/service/product/model/schema"
)

func TestManualDraftPublicationAndVisibility(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "manuals.db")), &gorm.Config{})
	check := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	check(err)
	sqlDB, err := db.DB()
	check(err)
	t.Cleanup(func() { sqlDB.Close() })
	check(schema.Apply(db))
	manuals := NewManualService(mysql.NewProductDocsRepository(db))
	check(manuals.SaveGroup(domain.DocGroupInput{Version: "v1", Title: "Start"}))
	check(manuals.SaveGroup(domain.DocGroupInput{Version: "v1", Title: "Advanced"}))
	workspace, err := manuals.Workspace("guide", "v1")
	check(err)
	before, after := workspace.Groups[0], workspace.Groups[1]
	input := domain.DocumentInput{Version: "v1", ID: "install", Slug: "start", Title: "Install", GroupID: before.ID}
	saved, err := manuals.SaveDocument(input, 17)
	check(err)
	visibility := domain.DocumentDeleteInput{Version: "v1", ID: "install", Revision: saved.Revision}
	if _, err = manuals.SetDocumentPublished(visibility, 17, true); !errors.Is(err, domain.ErrInvalidDocument) {
		t.Fatalf("published blank document: %v", err)
	}
	if _, err = manuals.PublicDocument("install", "v1", "guide"); !errors.Is(err, domain.ErrDocumentNotFound) {
		t.Fatalf("draft was public: %v", err)
	}
	input.Content = "Public Markdown"
	input.Revision = saved.Revision
	saved, err = manuals.SaveDocument(input, 17)
	check(err)
	visibility.Revision = saved.Revision
	live, err := manuals.SetDocumentPublished(visibility, 17, true)
	check(err)
	input.Revision = live.Revision
	input.Content = "Unpublished changes"
	input.Slug = "edited"
	input.Title = "Private title"
	input.GroupID = after.ID
	modified, err := manuals.SaveDocument(input, 23)
	check(err)
	if modified.Status != "modified" || modified.PublishedRevision != live.Revision {
		t.Fatal(modified)
	}
	if _, err = manuals.SaveDocument(input, 17); !errors.Is(err, domain.ErrDocumentConflict) {
		t.Fatalf("stale save: %v", err)
	}
	public, err := manuals.PublicDocument("install", "v1", "guide")
	check(err)
	if public.Content != "Public Markdown" || public.Slug != "start" || public.GroupID != before.ID {
		t.Fatalf("draft leaked into live article: %+v", public)
	}
	catalog, err := manuals.GetDocCatalog("v1", "guide")
	check(err)
	if catalog.Navigation.Groups[0].Items[0].Title != "Install" || len(catalog.Navigation.Groups[1].Items) != 0 {
		t.Fatalf("draft leaked into sidebar: %+v", catalog.Navigation.Groups)
	}
	if err = manuals.DeleteGroup("v1", before.ID, before.Revision); !errors.Is(err, domain.ErrDocumentConflict) {
		t.Fatalf("removed referenced group: %v", err)
	}
	workspace, err = manuals.Workspace("guide", "v1")
	check(err)
	raw, err := json.Marshal(workspace)
	check(err)
	if strings.Contains(string(raw), "Unpublished changes") || strings.Contains(string(raw), "\"content\"") {
		t.Fatal("workspace loaded document bodies")
	}
	history, err := manuals.History("v1", "install")
	check(err)
	if len(history) != 4 || history[0].ContentMarkdown != "" {
		t.Fatalf("unexpected history metadata: %+v", history)
	}
	snapshot, err := manuals.Historical("v1", "install", history[1].ID)
	check(err)
	if snapshot.Operation != "publish" || snapshot.ContentMarkdown != "Public Markdown" || snapshot.ActorID != 17 {
		t.Fatalf("bad historical snapshot: %+v", snapshot)
	}

	visibility.Revision = modified.Revision
	if err = manuals.DeleteDocument(visibility, 23); !errors.Is(err, domain.ErrDocumentConflict) {
		t.Fatalf("deleted published article: %v", err)
	}
	// A failed history insertion must roll back the visibility pointer and revision.
	check(db.Exec("CREATE TRIGGER reject_withdraw BEFORE INSERT ON product_document_revisions WHEN NEW.operation = 'withdraw' BEGIN SELECT RAISE(ABORT,'history unavailable'); END").Error)
	if _, err = manuals.SetDocumentPublished(visibility, 23, false); err == nil {
		t.Fatal("expected history error")
	}
	public, err = manuals.PublicDocument("install", "v1", "guide")
	check(err)
	if public.Revision != live.Revision {
		t.Fatal(public)
	}
	check(db.Exec("DROP TRIGGER reject_withdraw").Error)
	hidden, err := manuals.SetDocumentPublished(visibility, 23, false)
	check(err)
	if hidden.Status != "draft" || hidden.PublishedRevision != 0 {
		t.Fatal(hidden)
	}
	if _, err = manuals.PublicDocument("install", "v1", "guide"); !errors.Is(err, domain.ErrDocumentNotFound) {
		t.Fatalf("withdrawn article: %v", err)
	}
	check(manuals.DeleteGroup("v1", before.ID, before.Revision))
	visibility.Revision = hidden.Revision
	check(manuals.DeleteDocument(visibility, 23))
	check(manuals.DeleteGroup("v1", after.ID, after.Revision))

	check(manuals.SaveVersion(domain.DocVersionInput{ID: "next", ModeID: "guide", Name: "Next"}, 17))
	next, err := manuals.Workspace("guide", "next")
	check(err)
	if len(next.Groups) != 1 || next.Groups[0].Title != "快速开始" {
		t.Fatal(next)
	}
	if err = manuals.SetVersionPublished("next", 1, true); !errors.Is(err, domain.ErrInvalidDocument) {
		t.Fatalf("published empty version: %v", err)
	}
	input.Version = "next"
	input.GroupID = next.Groups[0].ID
	input.Revision = 0
	draft, err := manuals.SaveDocument(input, 17)
	check(err)
	_, err = manuals.SetDocumentPublished(domain.DocumentDeleteInput{ID: "install", Version: "next", Revision: draft.Revision}, 17, true)
	check(err)
	if _, err = manuals.PublicDocument("install", "next", "guide"); !errors.Is(err, domain.ErrDocumentNotFound) {
		t.Fatalf("draft version leak: %v", err)
	}
	check(manuals.SetVersionPublished("next", 1, true))
	_, err = manuals.PublicDocument("install", "next", "guide")
	check(err)
	check(manuals.SetVersionPublished("next", 2, false))
	if _, err = manuals.PublicDocument("install", "next", "guide"); !errors.Is(err, domain.ErrDocumentNotFound) {
		t.Fatalf("withdrawn version leak: %v", err)
	}
	if err = manuals.DeleteVersion("next", 3); !errors.Is(err, domain.ErrDocumentConflict) {
		t.Fatalf("deleted populated version: %v", err)
	}
}
