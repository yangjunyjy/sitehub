package file

import (
	"sitehub/common/config"
	settings "sitehub/common/settings/contract"
	fileservice "sitehub/service/file/business"
	"sitehub/service/file/contract"
)

type Module struct {
	Settings contract.Settings
	Uploads  fileservice.UploadService
	Private  contract.PrivateStorage
}

func New(cfg *config.Config, store settings.Store) *Module {
	defaults := fileservice.DefaultWebsiteStorage()
	if cfg.Upload.Local.UploadPath != "" {
		defaults.LocalPath = cfg.Upload.Local.UploadPath
	}
	defaults.URLPrefix = cfg.Upload.URLPrefix
	s := fileservice.NewStorageSettingsWithDefaults(store, defaults)
	return &Module{
		Settings: s,
		Uploads:  fileservice.NewUploadService(&cfg.Upload, s),
		Private:  fileservice.NewPrivateStorage(&cfg.Upload, s, cfg.Auth.JWTSecret),
	}
}
