package services

import (
	"context"
	"errors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"sitehub/common/errorx"
	"sitehub/service/product/contract"
	domain "sitehub/service/product/domain"
	"sitehub/service/product/model/schema"
	"testing"
)

type testReleaseValidator struct{ run func(string) error }

func (v *testReleaseValidator) ValidateRelease(url string) error { return v.run(url) }
func releaseDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sql, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sql.SetMaxOpenConns(1)
	t.Cleanup(func() { sql.Close() })
	if err := schema.ApplyReleases(db); err != nil {
		t.Fatal(err)
	}
	return db
}
func checkReleaseError(t *testing.T, err error, kind errorx.ErrorKind) {
	t.Helper()
	var public *errorx.PublicError
	if !errors.As(err, &public) || public.Kind != kind {
		t.Fatalf("expected %d, got %v", kind, err)
	}
}
func TestReleaseLifecycleAndReferences(t *testing.T) {
	db := releaseDB(t)
	ctx := context.Background()
	validator := &testReleaseValidator{run: func(string) error { return errorx.Invalid("文件缺失") }}
	s := NewReleaseService(db, validator)
	group, err := s.SaveGroup(ctx, domain.ReleaseGroup{ID: "tools", Name: "工具"}, true)
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.SaveGroup(ctx, group, true)
	checkReleaseError(t, err, errorx.Conflict)
	invalid := domain.ReleaseFile{ID: "cli", Product: "CLI", Platform: "Linux", Version: "v1", URL: "/v1/file/local/releases/a.zip", GroupID: "absent"}
	_, err = s.SaveFile(ctx, invalid, true)
	checkReleaseError(t, err, errorx.NotFound)
	invalid.GroupID = group.ID
	file, err := s.SaveFile(ctx, invalid, true)
	if err != nil {
		t.Fatal(err)
	}
	checkReleaseError(t, s.DeleteGroup(ctx, group.ID, group.Revision), errorx.Conflict)
	_, err = s.Resolve(ctx, file.ID)
	checkReleaseError(t, err, errorx.NotFound)
	_, err = s.SetPublished(ctx, file.ID, file.Revision, true)
	checkReleaseError(t, err, errorx.InvalidInput)
	// A concurrent editor changes the record while storage is being verified.
	validator.run = func(string) error {
		edit := file
		edit.Description = "修改说明"
		_, e := s.SaveFile(ctx, edit, false)
		return e
	}
	_, err = s.SetPublished(ctx, file.ID, file.Revision, true)
	checkReleaseError(t, err, errorx.Conflict)
	validator.run = func(string) error { return nil }
	rows, err := s.Files(ctx)
	if err != nil || len(rows) != 1 {
		t.Fatal(rows, err)
	}
	live, err := s.SetPublished(ctx, file.ID, rows[0].Revision, true)
	if err != nil {
		t.Fatal(err)
	}
	target, err := s.Resolve(ctx, live.ID)
	if err != nil || target.URL != file.URL {
		t.Fatal(target, err)
	}
	catalog, err := s.Catalog(ctx)
	if err != nil || len(catalog.Packages) != 1 || catalog.Packages[0].URL != "/v1/file/download?id=cli" {
		t.Fatal(catalog, err)
	}
	archive, err := s.SetPublished(ctx, live.ID, live.Revision, false)
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.Resolve(ctx, file.ID)
	checkReleaseError(t, err, errorx.NotFound)
	catalog, err = s.Catalog(ctx)
	if err != nil || len(catalog.Packages) != 0 {
		t.Fatal(catalog, err)
	}
	checkReleaseError(t, s.DeleteFile(ctx, file.ID, file.Revision), errorx.Conflict)
	if err := s.DeleteFile(ctx, file.ID, archive.Revision); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteGroup(ctx, group.ID, group.Revision); err != nil {
		t.Fatal(err)
	}
}
func TestReleaseSettingsAtomicReplacement(t *testing.T) {
	s := NewReleaseService(releaseDB(t), nil)
	ctx := context.Background()
	settings, err := s.Settings(ctx)
	if err != nil {
		t.Fatal(err)
	}
	settings.Title = "技术工具"
	settings.InstallOptions = []contract.WebsiteInstallOption{{ID: "docker", Title: "容器"}, {ID: "manual", Title: "手动"}}
	settings.Requirements = []contract.WebsiteRequirement{{Label: "系统", Value: "Linux"}}
	saved, err := s.SaveSettings(ctx, settings)
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.SaveSettings(ctx, settings)
	checkReleaseError(t, err, errorx.Conflict)
	view, err := s.Catalog(ctx)
	if err != nil || len(view.InstallOptions) != 2 || view.InstallOptions[1].Title != "手动" || len(view.Requirements) != 1 {
		t.Fatal(view, err)
	}
	saved.Requirements = nil
	saved.InstallOptions = nil
	if _, err = s.SaveSettings(ctx, saved); err != nil {
		t.Fatal(err)
	}
	view, err = s.Catalog(ctx)
	if err != nil || len(view.Requirements) != 0 || len(view.InstallOptions) != 0 {
		t.Fatal(view, err)
	}
}
func TestReleaseFileValidation(t *testing.T) {
	for _, url := range []string{"https://cdn.example.test/release.tgz", "/v1/file/local/releases/example.tgz"} {
		if !validTarget(url) {
			t.Errorf("expected valid target %q", url)
		}
	}
	for _, url := range []string{"//cdn.example.test/a.zip", "/other/file", "/v1/file/local/releases/../config", "/v1/file/local/releases/bad%2fname", "https://user:secret@example.test/x", "file:///tmp/file"} {
		if validTarget(url) {
			t.Errorf("unexpected valid target %q", url)
		}
	}
}
