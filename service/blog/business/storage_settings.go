package business

import (
	"encoding/json"
	"fmt"
	"strings"

	errorx "sitehub/common/errorx"
	settingsservice "sitehub/common/settings/contract"
	settingsmodel "sitehub/common/settings/model"
	filecontract "sitehub/service/blog/contract"
)

const websiteStorageKey = "website.storage"

type StorageSettings struct {
	configs  settingsservice.Store
	defaults filecontract.WebsiteStorage
}

func NewStorageSettings(configs settingsservice.Store) *StorageSettings {
	return NewStorageSettingsWithDefaults(configs, DefaultWebsiteStorage())
}
func NewStorageSettingsWithDefaults(configs settingsservice.Store, defaults filecontract.WebsiteStorage) *StorageSettings {
	return &StorageSettings{configs: configs, defaults: defaults}
}
func (s *StorageSettings) EnsureDefaults() error {
	row, err := s.configs.GetConfig(websiteStorageKey)
	if err != nil {
		return err
	}
	if row != nil {
		if row.IsPublic {
			row.IsPublic = false
			return s.configs.UpdateConfig(row)
		}
		return nil
	}
	raw, err := json.Marshal(s.defaults)
	if err != nil {
		return err
	}
	return s.configs.CreateConfig(&settingsmodel.SystemConfig{Key: websiteStorageKey, Value: string(raw), Category: settingsmodel.ConfigCategoryContent, ConfigType: settingsmodel.ConfigTypeJSON, Label: "文件存储", Description: "官网图片、附件和下载包的存储适配器", IsEditable: true, IsRequired: true, SortOrder: 4, IsPublic: false})
}
func (s *StorageSettings) GetStorage() (filecontract.WebsiteStorage, error) {
	value := s.defaults
	row, err := s.configs.GetConfig(websiteStorageKey)
	if err != nil {
		return value, err
	}
	if row != nil && strings.TrimSpace(row.Value) != "" {
		if err = json.Unmarshal([]byte(row.Value), &value); err != nil {
			return value, fmt.Errorf("配置 %s 不是有效 JSON: %w", websiteStorageKey, err)
		}
	}
	return value, nil
}
func (s *StorageSettings) SaveStorage(storage filecontract.WebsiteStorage) error {
	normalizeWebsiteStorage(&storage)
	if err := validateWebsiteStorage(storage); err != nil {
		return errorx.Invalid(err.Error())
	}
	raw, err := json.Marshal(storage)
	if err != nil {
		return err
	}
	if err = s.EnsureDefaults(); err != nil {
		return err
	}
	return s.configs.BatchUpdateConfigs([]settingsservice.BatchUpdateConfig{{Key: websiteStorageKey, Value: string(raw), ConfigType: string(settingsmodel.ConfigTypeJSON)}})
}

func validateWebsiteStorage(storage filecontract.WebsiteStorage) error {
	provider := strings.ToLower(strings.TrimSpace(storage.Provider))
	if provider != "local" && provider != "minio" && provider != "oss" {
		return fmt.Errorf("存储类型仅支持 local、minio 或 oss")
	}
	if storage.MaxSize <= 0 {
		return fmt.Errorf("文件大小限制必须大于 0")
	}
	if provider == "local" && strings.TrimSpace(storage.LocalPath) == "" {
		return fmt.Errorf("本地存储必须配置存储目录")
	}
	if provider != "local" && (strings.TrimSpace(storage.Endpoint) == "" || strings.TrimSpace(storage.Bucket) == "" || strings.TrimSpace(storage.AccessKey) == "" || strings.TrimSpace(storage.SecretKey) == "") {
		return fmt.Errorf("对象存储必须配置 endpoint、bucket、accessKey 和 secretKey")
	}
	return nil
}

func normalizeWebsiteStorage(storage *filecontract.WebsiteStorage) {
	storage.Provider = strings.ToLower(strings.TrimSpace(storage.Provider))
	storage.LocalPath = strings.TrimSpace(storage.LocalPath)
	storage.URLPrefix = strings.TrimRight(strings.TrimSpace(storage.URLPrefix), "/")
	storage.CDN = strings.TrimRight(strings.TrimSpace(storage.CDN), "/")
	storage.Endpoint = strings.TrimRight(strings.TrimSpace(storage.Endpoint), "/")
	storage.AccessKey = strings.TrimSpace(storage.AccessKey)
	storage.Bucket = strings.TrimSpace(storage.Bucket)
	storage.Region = strings.TrimSpace(storage.Region)
	exts := make([]string, 0, len(storage.AllowedExts))
	seen := make(map[string]struct{}, len(storage.AllowedExts))
	for _, value := range storage.AllowedExts {
		ext := strings.ToLower(strings.TrimSpace(value))
		if ext == "" {
			continue
		}
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		if _, ok := seen[ext]; ok {
			continue
		}
		seen[ext] = struct{}{}
		exts = append(exts, ext)
	}
	storage.AllowedExts = exts
}

func DefaultWebsiteStorage() filecontract.WebsiteStorage {
	return filecontract.WebsiteStorage{
		Provider: "local", LocalPath: "./data/files", URLPrefix: "",
		PathStyle: true, AllowedExts: []string{
			".png", ".jpg", ".jpeg", ".gif", ".webp", ".ico",
			".pdf", ".xls", ".xlsx", ".ppt", ".pptx", ".txt", ".md", ".csv",
			".json", ".xml", ".yaml", ".yml", ".mp3", ".mp4", ".avi", ".mov", ".rar", ".7z",
			".zip", ".tar", ".gz", ".tgz", ".deb", ".rpm", ".exe", ".dmg", ".pkg", ".msi",
		},
		MaxSize: 2 * 1024 * 1024 * 1024,
	}
}
