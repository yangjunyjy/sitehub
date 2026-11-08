package services

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	errorx "sitehub/common/errorx"
	settingscontract "sitehub/common/settings/contract"
	settingsmodel "sitehub/common/settings/model"
	settingsrepository "sitehub/common/settings/repository"
	settingsservice "sitehub/common/settings/service"
	fileservice "sitehub/service/file/business"
	filecontract "sitehub/service/file/contract"
	productcontract "sitehub/service/product/contract"
)

func TestWebsiteValidationIsDistinctFromStorageFailures(t *testing.T) {
	cause := errors.New("private database/storage failure")
	service := NewWebsiteContentService(&downloadConfigStore{err: cause}, nil, fileservice.NewStorageSettings(&downloadConfigStore{err: cause}))
	for _, err := range []error{
		service.SaveContent(DefaultWebsiteContent()),
		service.SaveStorage(fileservice.DefaultWebsiteStorage()),
	} {
		var public *errorx.PublicError
		if !errors.Is(err, cause) || errors.As(err, &public) {
			t.Fatalf("dependency error misclassified as validation: %v", err)
		}
	}
	for _, err := range []error{
		service.SaveContent(productcontract.WebsiteContent{}),
		service.SaveStorage(filecontract.WebsiteStorage{Provider: "unsupported"}),
	} {
		var public *errorx.PublicError
		if !errors.As(err, &public) || public.Kind != errorx.InvalidInput {
			t.Fatalf("expected actionable validation error, got %v", err)
		}
	}
}

// The existing system-config service writes one JSON document in one transaction.
// This fake checks the service contract without writing to the user's database.
type downloadConfigStore struct {
	settingsservice.SystemConfigService

	value        string
	writes       int
	err          error
	storage      string
	beforeCommit func()
}

func (s *downloadConfigStore) UpdateJSONConfigIfUnchanged(key, expected, value string) error {
	if s.err != nil {
		return s.err
	}
	if s.beforeCommit != nil {
		s.beforeCommit()
	}
	if key != websiteDownloadsKey || s.value != expected {
		return errorx.Conflicting("配置已被其他操作更新，请重新加载后重试")
	}
	s.writes++
	s.value = value
	return nil
}

func (s *downloadConfigStore) GetConfig(key string) (*settingsmodel.SystemConfig, error) {
	if s.err != nil {
		return nil, s.err
	}
	if key == websiteStorageKey {
		return &settingsmodel.SystemConfig{Key: key, Value: s.storage}, nil
	}
	return &settingsmodel.SystemConfig{Key: key, Value: s.value}, nil
}
func (s *downloadConfigStore) BatchUpdateConfigs(values []settingscontract.BatchUpdateConfig) error {
	if s.err != nil {
		return s.err
	}
	s.writes++
	s.value = values[0].Value
	return nil
}

func TestDownloadReleaseLifecycle(t *testing.T) {
	store := &downloadConfigStore{value: `{"version":"v1.0.0","releaseDate":"2026-09-20","packages":[{"id":"old","status":"published","platform":"Linux","url":"https://files.example.test/old.tgz"},{"id":"placeholder","status":"draft","platform":"Windows"}]}`}
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "releases"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "releases", "app.dmg"), []byte("installer"), 0644); err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(filecontract.WebsiteStorage{Provider: "local", LocalPath: root})
	store.storage = string(encoded)
	service := NewWebsiteContentService(store, nil, fileservice.NewStorageSettings(store))
	admin, err := service.GetDownloads()
	if err != nil {
		t.Fatal(err)
	}
	if len(admin.Packages) != 2 || admin.Packages[0].Status != "published" || admin.Packages[1].Status != "draft" {
		t.Fatalf("legacy packages: %+v", admin.Packages)
	}
	admin.Packages[0].Status = "archived"
	newRelease := productcontract.WebsiteDownloadPackage{ID: "new", Product: "Deploy Pro", Version: "v2.0.0", Platform: "macOS", Arch: "ARM64", URL: "/v1/file/local/releases/app.dmg", Format: "DMG", FileName: "app.dmg", Size: 12345, ReleaseDate: "2026-09-21", Status: "published", Description: "New release"}
	admin.Packages = append(admin.Packages, newRelease)
	if err := service.SaveDownloads(admin); err != nil {
		t.Fatal(err)
	}
	reloaded, err := service.GetDownloads()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(newRelease, reloaded.Packages[2]) {
		t.Fatalf("metadata lost: %+v", reloaded.Packages[2])
	}
	public, err := service.GetPublicDownloads()
	if err != nil {
		t.Fatal(err)
	}
	if len(public.Packages) != 1 || public.Packages[0].ID != "new" {
		t.Fatalf("public packages expose unpublished releases: %+v", public.Packages)
	}
	if public.Packages[0].URL != "/v1/file/download?id=new" {
		t.Fatalf("public manifest exposed raw URL: %s", public.Packages[0].URL)
	}
	resolved, err := service.GetPublishedDownload("new")
	if err != nil || resolved.URL != newRelease.URL || resolved.FileName != newRelease.FileName {
		t.Fatalf("published package resolution: %+v %v", resolved, err)
	}
	if store.writes != 1 {
		t.Fatalf("expected single document write, got %d", store.writes)
	}
}

type publicDownloadConfigRepo struct {
	settingsrepository.
		SystemConfigRepository
}

func (*publicDownloadConfigRepo) GetPublicConfigs() ([]settingsmodel.SystemConfig, error) {
	return []settingsmodel.SystemConfig{
		{Key: websiteContentKey, IsPublic: true},
		{Key: websiteDownloadsKey, Value: "private artifact URL", IsPublic: true},
		{Key: websiteStorageKey, Value: "storage secret", IsPublic: true},
	}, nil
}

func TestPublicConfigCannotExposeRawDownloadManifest(t *testing.T) {
	configs, err := settingsservice.NewSystemConfigService(&publicDownloadConfigRepo{}).GetPublicConfigs()
	if err != nil || len(configs) != 1 || configs[0].Key != websiteContentKey {
		t.Fatalf("legacy public flags leaked restricted configuration: %+v %v", configs, err)
	}
}

func TestLocalDownloadURLValidationPreservesOriginalFile(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "releases"), 0755)
	os.WriteFile(filepath.Join(root, "releases", "stored-id.tgz"), []byte("package"), 0644)
	encoded, _ := json.Marshal(filecontract.WebsiteStorage{Provider: "local", LocalPath: root})
	store := &downloadConfigStore{value: `{"version":"1.0"}`, storage: string(encoded)}
	service := NewWebsiteContentService(store, nil, fileservice.NewStorageSettings(store))
	input := productcontract.WebsiteDownloads{Version: "1.0", Packages: []productcontract.WebsiteDownloadPackage{{ID: "arm", Platform: "Linux", Status: "published", URL: "/v1/file/local/releases/stored-id.tgz", FileName: "deploy_arm.tgz"}}}
	if err := service.SaveDownloads(input); err != nil {
		t.Fatal(err)
	}
	saved, err := service.GetDownloads()
	if err != nil {
		t.Fatal(err)
	}
	if saved.Packages[0].URL != "/v1/file/local/releases/stored-id.tgz" || saved.Packages[0].FileName != "deploy_arm.tgz" {
		t.Fatalf("wrong metadata: %+v", saved.Packages[0])
	}
	before := store.value
	input.Packages[0].URL = "/v1/file/local/releases/deploy_arm.tgz"
	if err = service.SaveDownloads(input); !errors.Is(err, ErrInvalidWebsiteDownloads) {
		t.Fatalf("missing file accepted: %v", err)
	}
	if store.value != before || store.writes != 1 {
		t.Fatal("invalid URL changed the manifest")
	}
	input.Packages[0].Status = "archived"
	if err = service.SaveDownloads(input); err != nil {
		t.Fatal("missing file must not block unpublishing", err)
	}
	input.Packages[0].Status = "published"
	input.Packages[0].URL = "/v1/file/local/releases/../releases/stored-id.tgz"
	if err = service.SaveDownloads(input); !errors.Is(err, ErrInvalidWebsiteDownloads) {
		t.Fatalf("traversal accepted: %v", err)
	}
	if content, err := os.ReadFile(filepath.Join(root, "releases", "stored-id.tgz")); err != nil || string(content) != "package" {
		t.Fatal("renaming download changed stored data")
	}
	for _, bad := range []string{"../a.tgz", "a\\b.tgz", "a\r\nX-Test: injected", "a:stream", ".."} {
		input.Packages[0].URL = "https://cdn.example.com/package.tgz"
		input.Packages[0].FileName = bad
		if err = service.SaveDownloads(input); !errors.Is(err, ErrInvalidWebsiteDownloads) {
			t.Fatalf("unsafe download name accepted: %q", bad)
		}
	}
	input.Packages[0].URL = "https://cdn.example.com/package.tgz?signature=unchanged"
	input.Packages[0].FileName = "deploy_arm.tgz"
	if err = service.SaveDownloads(input); err != nil {
		t.Fatal("external URL must not be checked on the local disk", err)
	}
}

func TestNativeURLsAreNotRewritten(t *testing.T) {
	for _, location := range []string{"http://localhost:8081/uploads/releases/a.tgz", "https://cdn.example.com/a.tgz", "/v1/file/local/releases/a.tgz"} {
		if got := normalizeReleaseURL(location); got != location {
			t.Fatalf("URL rewritten: %s", got)
		}
	}
}

func TestInvalidDownloadBatchDoesNotSaveAnyChanges(t *testing.T) {
	for _, item := range []productcontract.WebsiteDownloadPackage{
		{ID: "bad", Platform: "Linux", Status: "unknown"},
		{ID: "bad", Platform: "Linux", Status: "published"},
		{ID: "bad", Platform: "Linux", Status: "draft", URL: "javascript:alert(1)"},
		{ID: "bad", Platform: "Linux", Status: "draft", URL: "//untrusted.test/file"},
		{ID: "bad", Platform: "Linux", Status: "draft", Size: -1},
		{ID: "bad", Platform: "Linux", Status: "draft", Checksum: "invalid"},
		{ID: " good ", Platform: "Linux", Status: "draft"},
	} {
		store := &downloadConfigStore{value: `{"version":"v1"}`}
		service := NewWebsiteContentService(store, nil, fileservice.NewStorageSettings(store))
		input := productcontract.WebsiteDownloads{Version: "v2", Packages: []productcontract.WebsiteDownloadPackage{{ID: "good", Platform: "Linux", URL: "https://files.example.test/app.zip", Status: "published"}, item}}
		if err := service.SaveDownloads(input); !errors.Is(err, ErrInvalidWebsiteDownloads) {
			t.Fatalf("expected validation error for %+v; got %v", item, err)
		}
		if store.writes != 0 || store.value != `{"version":"v1"}` {
			t.Fatal("invalid input changed persisted content")
		}
	}
}

func TestDownloadStorageFailureRemainsDependencyError(t *testing.T) {
	failure := errors.New("database unavailable")
	service := NewWebsiteContentService(&downloadConfigStore{err: failure}, nil, fileservice.NewStorageSettings(&downloadConfigStore{err: failure}))
	if _, err := service.GetPublicDownloads(); !errors.Is(err, failure) {
		t.Fatalf("read: %v", err)
	}
	if err := service.SaveDownloads(productcontract.WebsiteDownloads{Version: "v1"}); !errors.Is(err, failure) || errors.Is(err, ErrInvalidWebsiteDownloads) {
		t.Fatalf("save: %v", err)
	}
}
