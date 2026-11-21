package services

import (
	"context"
	"gorm.io/gorm"
	"sitehub/common/errorx"
	"sitehub/service/product/contract"
	domain "sitehub/service/product/domain"
	"strings"
)

func releaseSettings(tx *gorm.DB) (domain.ReleaseSettings, error) {
	var value domain.ReleaseSettings
	if err := tx.First(&value, 1).Error; err != nil {
		return value, err
	}
	var installs []domain.ReleaseInstallOption
	if err := tx.Order("position").Find(&installs).Error; err != nil {
		return value, err
	}
	value.InstallOptions = make([]contract.WebsiteInstallOption, 0, len(installs))
	for _, row := range installs {
		value.InstallOptions = append(value.InstallOptions, row.WebsiteInstallOption)
	}
	var requirements []domain.ReleaseRequirement
	if err := tx.Order("position").Find(&requirements).Error; err != nil {
		return value, err
	}
	value.Requirements = make([]contract.WebsiteRequirement, 0, len(requirements))
	for _, row := range requirements {
		value.Requirements = append(value.Requirements, row.WebsiteRequirement)
	}
	return value, nil
}
func (s *ReleaseService) Settings(ctx context.Context) (value domain.ReleaseSettings, err error) {
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error { var e error; value, e = releaseSettings(tx); return e })
	return
}
func (s *ReleaseService) SaveSettings(ctx context.Context, value domain.ReleaseSettings) (domain.ReleaseSettings, error) {
	if value.InstallOptions == nil {
		value.InstallOptions = []contract.WebsiteInstallOption{}
	}
	if value.Requirements == nil {
		value.Requirements = []contract.WebsiteRequirement{}
	}
	value.ID = 1
	value.Title = strings.TrimSpace(value.Title)
	if value.Revision == 0 || value.Title == "" || !validText(value.Title, 80) || !validText(value.Description, 5000) || !validText(value.SupportDescription, 5000) || !validText(value.SupportTitle, 160) || !validText(value.SupportArchitectures, 255) || !validText(value.Version, 120) || !validText(value.ReleaseDate, 80) || len(value.InstallOptions) > 100 || len(value.Requirements) > 100 {
		return value, errorx.Invalid("下载设置字段或修订号无效")
	}
	seen := map[string]bool{}
	for _, v := range value.InstallOptions {
		if !releaseID.MatchString(v.ID) || seen[v.ID] || strings.TrimSpace(v.Title) == "" || !validText(v.Title, 120) || !validText(v.Badge, 120) || !validText(v.Command, 16000) || !validText(v.Description, 5000) {
			return value, errorx.Invalid("安装方式标识、名称或内容无效")
		}
		seen[v.ID] = true
	}
	for _, v := range value.Requirements {
		if strings.TrimSpace(v.Label) == "" || strings.TrimSpace(v.Value) == "" || !validText(v.Label, 120) || !validText(v.Value, 2000) {
			return value, errorx.Invalid("请填写有效的环境要求")
		}
	}
	rev := value.Revision
	value.Revision++
	return value, s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := affectedRelease(tx.Model(&domain.ReleaseSettings{}).Where("id = 1 AND revision = ?", rev).Select("*").Omit("id").Updates(value)); err != nil {
			return err
		}
		if err := tx.Where("position >= 0").Delete(&domain.ReleaseInstallOption{}).Error; err != nil {
			return err
		}
		if err := tx.Where("position >= 0").Delete(&domain.ReleaseRequirement{}).Error; err != nil {
			return err
		}
		for i, item := range value.InstallOptions {
			if err := tx.Create(&domain.ReleaseInstallOption{Position: i + 1, WebsiteInstallOption: item}).Error; err != nil {
				return err
			}
		}
		for i, item := range value.Requirements {
			if err := tx.Create(&domain.ReleaseRequirement{Position: i + 1, WebsiteRequirement: item}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
