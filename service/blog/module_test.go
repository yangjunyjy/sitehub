package blog

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	coreconfig "sitehub/common/config"
	settings "sitehub/common/settings/contract"
	settingsmodel "sitehub/common/settings/model"
)

type moduleSettingsStore struct {
	value *settingsmodel.SystemConfig
}

func (s *moduleSettingsStore) GetConfig(key string) (*settingsmodel.SystemConfig, error) {
	if s.value == nil || s.value.Key != key {
		return nil, nil
	}
	value := *s.value
	return &value, nil
}

func (s *moduleSettingsStore) CreateConfig(value *settingsmodel.SystemConfig) error {
	if s.value != nil {
		return errors.New("configuration already exists")
	}
	copy := *value
	s.value = &copy
	return nil
}

func (s *moduleSettingsStore) UpdateConfig(value *settingsmodel.SystemConfig) error {
	copy := *value
	s.value = &copy
	return nil
}

func (s *moduleSettingsStore) BatchUpdateConfigs(values []settings.BatchUpdateConfig) error {
	for _, value := range values {
		if s.value == nil || s.value.Key != value.Key {
			return errors.New("configuration does not exist")
		}
		s.value.Value = value.Value
	}
	return nil
}

func (s *moduleSettingsStore) UpdateJSONConfigIfUnchanged(key, expected, value string) error {
	if s.value == nil || s.value.Key != key || s.value.Value != expected {
		return errors.New("configuration changed")
	}
	s.value.Value = value
	return nil
}

func TestModuleSharesStorageSettingsAndRetainsPrivateSnapshots(t *testing.T) {
	root := t.TempDir()
	uploadConfig := coreconfig.UploadConfig{Type: "local"}
	uploadConfig.Local.UploadPath = filepath.Join(root, "initial")
	store := &moduleSettingsStore{}
	secret := "shared-blog-storage-secret-at-least-32-characters"
	module := New(nil, uploadConfig, store, secret)
	storage, err := module.StorageSettings.GetStorage()
	if err != nil {
		t.Fatal(err)
	}
	if storage.LocalPath != uploadConfig.Local.UploadPath {
		t.Fatalf("storage default path = %q", storage.LocalPath)
	}
	storage.LocalPath = filepath.Join(root, "configured")
	if err = module.StorageSettings.SaveStorage(storage); err != nil {
		t.Fatal(err)
	}

	public, err := module.Uploads.UploadRelease("shared.tgz", 7, strings.NewReader("release"))
	if err != nil {
		t.Fatal(err)
	}
	publicPath := filepath.Join(storage.LocalPath, filepath.FromSlash(strings.TrimPrefix(public.URL, "/v1/file/local/")))
	if content, err := os.ReadFile(publicPath); err != nil || string(content) != "release" {
		t.Fatalf("public upload did not use shared settings: content=%q err=%v", content, err)
	}
	snapshot, err := module.PrivateStorage.PreparePrivate("article-assets/shared.txt")
	if err != nil {
		t.Fatal(err)
	}
	if err = module.PrivateStorage.WritePrivate(context.Background(), snapshot, strings.NewReader("private"), "text/plain"); err != nil {
		t.Fatal(err)
	}
	privatePath := filepath.Join(storage.LocalPath+"-private", "article-assets", "shared.txt")
	if content, err := os.ReadFile(privatePath); err != nil || string(content) != "private" {
		t.Fatalf("private upload did not use shared settings: content=%q err=%v", content, err)
	}

	// Historical article resources retain their encrypted storage location after
	// settings changes and reconstruction of the unified module with the same key.
	storage.LocalPath = filepath.Join(root, "next")
	if err = module.StorageSettings.SaveStorage(storage); err != nil {
		t.Fatal(err)
	}
	restarted := New(nil, uploadConfig, store, secret)
	body, err := restarted.PrivateStorage.ReadPrivate(context.Background(), snapshot)
	if err != nil {
		t.Fatal(err)
	}
	defer body.Close()
	content, err := io.ReadAll(body)
	if err != nil || string(content) != "private" {
		t.Fatalf("historical snapshot unreadable: content=%q err=%v", content, err)
	}
	current, err := restarted.StorageSettings.GetStorage()
	if err != nil || current.LocalPath != storage.LocalPath {
		t.Fatalf("restarted module ignored persisted settings: path=%q err=%v", current.LocalPath, err)
	}
}
