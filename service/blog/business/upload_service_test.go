package business

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"sitehub/common/config"
	filecontract "sitehub/service/blog/contract"
)

func TestLocalFileErrorsUseFileContract(t *testing.T) {
	cfg := config.UploadConfig{Type: "local"}
	cfg.Local.UploadPath = t.TempDir()
	uploads := NewUploadService(&cfg, nil)
	if _, err := uploads.ServeLocalPath("../outside.txt"); !errors.Is(err, filecontract.ErrInvalidPath) {
		t.Fatalf("expected a file path error, got %v", err)
	}
	if _, err := uploads.ServeLocalPath("images/missing.png"); !errors.Is(err, filecontract.ErrNotFound) {
		t.Fatalf("expected a file not-found error, got %v", err)
	}
}

func TestUploadToLocalStreamsContentAndBuildsPublicURL(t *testing.T) {
	cfg := config.UploadConfig{}
	cfg.Local.UploadPath = t.TempDir()
	cfg.URLPrefix = "https://deploy.example.com/"

	location, err := uploadToLocal(cfg, "releases/deploy.zip", strings.NewReader("release-content"))
	if err != nil {
		t.Fatalf("uploadToLocal() error = %v", err)
	}
	if location != "https://deploy.example.com/v1/file/local/releases/deploy.zip" {
		t.Fatalf("location = %q", location)
	}
	content, err := os.ReadFile(filepath.Join(cfg.Local.UploadPath, "releases", "deploy.zip"))
	if err != nil {
		t.Fatalf("read uploaded file: %v", err)
	}
	if string(content) != "release-content" {
		t.Fatalf("content = %q", content)
	}
}

func TestLocalReleaseUploadDropsLoopbackPrefix(t *testing.T) {
	cfg := config.UploadConfig{}
	cfg.Local.UploadPath = t.TempDir()
	cfg.URLPrefix = ""
	location, err := uploadToLocal(cfg, "releases/random-id.tgz", strings.NewReader("installer"))
	if err != nil || location != "/v1/file/local/releases/random-id.tgz" {
		t.Fatalf("location=%q err=%v", location, err)
	}
}

func TestObjectURLSupportsPathAndVirtualHostStyles(t *testing.T) {
	cfg := config.UploadConfig{}
	cfg.Remote.Endpoint = "https://s3.example.com"
	cfg.Remote.Bucket = "deploy"
	cfg.Remote.PathStyle = true
	if got := objectURL(cfg, "releases/a.zip"); got != "https://s3.example.com/deploy/releases/a.zip" {
		t.Fatalf("path style URL = %q", got)
	}

	cfg.Remote.PathStyle = false
	if got := objectURL(cfg, "releases/a.zip"); got != "https://deploy.s3.example.com/releases/a.zip" {
		t.Fatalf("virtual host URL = %q", got)
	}
}

func TestObjectKeyFromLocationRejectsTraversal(t *testing.T) {
	cfg := config.UploadConfig{}
	cfg.Type = "local"
	if _, err := objectKeyFromLocation(cfg, "/v1/file/local/../../secret"); err == nil {
		t.Fatal("expected traversal path to be rejected")
	}
	got, err := objectKeyFromLocation(cfg, "https://deploy.example.com/v1/file/local/images/a.png")
	if err != nil {
		t.Fatalf("valid object key error = %v", err)
	}
	if got != "images/a.png" {
		t.Fatalf("object key = %q", got)
	}
}

func TestWebsiteStorageValidationAndNormalization(t *testing.T) {
	storage := filecontract.WebsiteStorage{
		Provider: " MINIO ", Endpoint: "https://minio.example.com/", Bucket: " releases ",
		AccessKey: " access ", SecretKey: "secret", MaxSize: 1024,
		AllowedExts: []string{"ZIP", ".zip", " tgz "},
	}
	normalizeWebsiteStorage(&storage)
	if err := validateWebsiteStorage(storage); err != nil {
		t.Fatalf("validateWebsiteStorage() error = %v", err)
	}
	if storage.Provider != "minio" || storage.Endpoint != "https://minio.example.com" || storage.Bucket != "releases" {
		t.Fatalf("normalized storage = %#v", storage)
	}
	if !reflect.DeepEqual(storage.AllowedExts, []string{".zip", ".tgz"}) {
		t.Fatalf("allowed extensions = %#v", storage.AllowedExts)
	}

	storage.SecretKey = ""
	if err := validateWebsiteStorage(storage); err == nil {
		t.Fatal("expected missing remote credentials to be rejected")
	}
}
