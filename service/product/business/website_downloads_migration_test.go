package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	errorx "sitehub/common/errorx"
	settingsrepository "sitehub/common/settings/repository"
	settingsservice "sitehub/common/settings/service"
	fileservice "sitehub/service/file/business"
	filecontract "sitehub/service/file/contract"
	productcontract "sitehub/service/product/contract"
)

// Missing release files must not prevent archiving or deleting existing entries.
const migratedDownloadManifest = `{"version":"1.0","packages":[{"id":"linux-arm64","status":"published","platform":"Linux","url":"/v1/file/local/releases/old-arm.tgz"},{"id":"linux-amd64","platform":"Linux","status":"published","url":"/v1/file/local/releases/old-amd.tgz"},{"id":"draft","platform":"Windows","status":"draft"}]}`

func migratedDownloads(t *testing.T) (WebsiteContentService, *downloadConfigStore, string) {
	t.Helper()
	root := t.TempDir()
	storage, err := json.Marshal(filecontract.WebsiteStorage{Provider: "local", LocalPath: root})
	if err != nil {
		t.Fatal(err)
	}
	store := &downloadConfigStore{value: migratedDownloadManifest, storage: string(storage)}
	return NewWebsiteContentService(store, nil, fileservice.NewStorageSettings(store)), store, root
}

func TestMissingMigratedFilesDoNotBlockDownloadMaintenance(t *testing.T) {
	for _, action := range []string{"delete-published", "delete-draft", "archive", "metadata"} {
		t.Run(action, func(t *testing.T) {
			service, store, _ := migratedDownloads(t)
			input, err := service.GetDownloads()
			if err != nil {
				t.Fatal(err)
			}
			switch action {
			case "delete-published":
				input.Packages = input.Packages[1:]
			case "delete-draft":
				input.Packages = input.Packages[:2]
			case "archive":
				input.Packages[0].Status = "archived"
			case "metadata":
				input.Packages[0].Description = "Migrated release"
				input.Version = "2.0"
			}
			if err := service.SaveDownloads(input); err != nil {
				t.Fatalf("unrelated missing file blocked %s: %v", action, err)
			}
			saved, err := service.GetDownloads()
			if err != nil || store.writes != 1 || len(saved.Packages) != len(input.Packages) {
				t.Fatalf("maintenance was not saved: %+v, %v", saved, err)
			}
			for i, item := range saved.Packages {
				if item.ID != input.Packages[i].ID || item.Status != input.Packages[i].Status || item.Description != input.Packages[i].Description {
					t.Fatalf("maintenance changed the wrong record: %+v", item)
				}
			}
		})
	}
}

func TestFreshUploadCanBePublishedBesideMissingMigratedFiles(t *testing.T) {
	for _, replace := range []bool{false, true} {
		name := "new-record"
		if replace {
			name = "replace-migrated-record"
		}
		t.Run(name, func(t *testing.T) {
			service, store, root := migratedDownloads(t)
			var body bytes.Buffer
			writer := multipart.NewWriter(&body)
			part, err := writer.CreateFormFile("file", "deploy.tgz")
			if err != nil {
				t.Fatal(err)
			}
			if _, err = part.Write([]byte("fresh uploaded package")); err != nil {
				t.Fatal(err)
			}
			if err = writer.Close(); err != nil {
				t.Fatal(err)
			}
			req := httptest.NewRequest("POST", "/upload", &body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			stream, header, err := req.FormFile("file")
			if err != nil {
				t.Fatal(err)
			}
			defer stream.Close()
			defer req.MultipartForm.RemoveAll()
			uploaded, err := fileservice.NewUploadService(nil, service).UploadRelease(header.Filename, header.Size, stream)
			if err != nil {
				t.Fatal(err)
			}
			input, err := service.GetDownloads()
			if err != nil {
				t.Fatal(err)
			}
			if replace {
				input.Packages[0].URL = uploaded.URL
			} else {
				input.Packages = append(input.Packages, productcontract.WebsiteDownloadPackage{ID: "new", Platform: "Linux", Status: "published", URL: uploaded.URL})
			}
			if err := service.SaveDownloads(input); err != nil || store.writes != 1 {
				t.Fatalf("valid upload blocked by migrated record: %v", err)
			}
			target := filepath.Join(root, filepath.FromSlash(strings.TrimPrefix(uploaded.URL, "/v1/file/local/")))
			content, err := os.ReadFile(target)
			if err != nil || string(content) != "fresh uploaded package" {
				t.Fatalf("uploaded file was lost: %q, %v", content, err)
			}
		})
	}
}

func TestMissingFileStillBlocksNewOrChangedPublication(t *testing.T) {
	for _, action := range []string{"new", "replace-url", "republish-archived", "publish-draft"} {
		t.Run(action, func(t *testing.T) {
			service, store, _ := migratedDownloads(t)
			input, err := service.GetDownloads()
			if err != nil {
				t.Fatal(err)
			}
			switch action {
			case "new":
				input.Packages = append(input.Packages, productcontract.WebsiteDownloadPackage{ID: "new", Platform: "Linux", Status: "published", URL: "/v1/file/local/releases/missing.tgz"})
			case "replace-url":
				input.Packages[0].URL = "/v1/file/local/releases/replacement-missing.tgz"
			case "republish-archived", "publish-draft":
				input.Packages[0].Status = "archived"
				if action == "publish-draft" {
					input.Packages[0].Status = "draft"
				}
				if err := service.SaveDownloads(input); err != nil {
					t.Fatal(err)
				}
				input.Packages[0].Status = "published"
			}
			before, writes := store.value, store.writes
			if err := service.SaveDownloads(input); !errors.Is(err, ErrInvalidWebsiteDownloads) {
				t.Fatalf("missing new publication accepted: %v", err)
			}
			if store.value != before || store.writes != writes {
				t.Fatal("failed publication changed the manifest")
			}
		})
	}
}

func TestDownloadSaveRejectsChangedSnapshot(t *testing.T) {
	service, store, _ := migratedDownloads(t)
	input, err := service.GetDownloads()
	if err != nil {
		t.Fatal(err)
	}
	input.Packages = input.Packages[:2]
	concurrent := strings.Replace(store.value, `"status":"published"`, `"status":"archived"`, 1)
	store.beforeCommit = func() { store.value = concurrent }
	err = service.SaveDownloads(input)
	var public *errorx.PublicError
	if !errors.As(err, &public) || public.Kind != errorx.Conflict || store.writes != 0 || store.value != concurrent {
		t.Fatalf("concurrent unpublish overwritten by stale validation: %v", err)
	}
}

type snapshotConfigRepo struct {
	settingsrepository.
		SystemConfigRepository
	matched bool
	err     error
}

func (r *snapshotConfigRepo) CompareAndSwapValue(string, string, string) (bool, error) {
	return r.matched, r.err
}

func TestConfigSnapshotErrorsKeepTheirClassification(t *testing.T) {
	cause := errors.New("private database failure")
	for _, tc := range []struct {
		name    string
		matched bool
		err     error
	}{
		{name: "success", matched: true},
		{name: "conflict"},
		{name: "dependency-failure", err: cause},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := settingsservice.NewSystemConfigService(&snapshotConfigRepo{matched: tc.matched, err: tc.err}).UpdateJSONConfigIfUnchanged(websiteDownloadsKey, "{}", "{}")
			var public *errorx.PublicError
			switch {
			case tc.err != nil:
				if !errors.Is(err, cause) || errors.As(err, &public) {
					t.Fatalf("dependency misclassified: %v", err)
				}
			case tc.matched:
				if err != nil {
					t.Fatal(err)
				}
			default:
				if !errors.As(err, &public) || public.Kind != errorx.Conflict {
					t.Fatalf("missing conflict: %v", err)
				}
			}
		})
	}
}
