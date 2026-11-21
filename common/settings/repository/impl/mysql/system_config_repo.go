// Package mysql provides MySQL implementation for system config repository
package mysql

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	settingsmodel "sitehub/common/settings/model"
	settingsrepository "sitehub/common/settings/repository"
)

type systemConfigRepository struct {
	db *gorm.DB
}

// NewSystemConfigRepository 创建系统配置仓储
func NewSystemConfigRepository(db *gorm.DB) settingsrepository.SystemConfigRepository {
	return &systemConfigRepository{db: db}
}

func (r *systemConfigRepository) CompareAndSwapValue(key, expected, value string) (bool, error) {
	matched := false
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var current settingsmodel.SystemConfig
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("config_key = ?", key).First(&current).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}
		// Compare exact bytes, not the database's case-insensitive text collation.
		if current.Value != expected {
			return nil
		}
		if err := tx.Model(&current).Update("value", value).Error; err != nil {
			return err
		}
		matched = true
		return nil
	})
	return matched && err == nil, err
}

func (r *systemConfigRepository) MutateValue(initial settingsmodel.SystemConfig, change func(string) (string, error)) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Creating the row and preserving a previous secret share the same transaction.
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&initial).Error; err != nil {
			return err
		}
		var current settingsmodel.SystemConfig
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("config_key = ?", initial.Key).First(&current).Error; err != nil {
			return err
		}
		value, err := change(current.Value)
		if err != nil {
			return err
		}
		return tx.Model(&current).Updates(map[string]any{"value": value, "is_public": initial.IsPublic}).Error
	})
}

func (r *systemConfigRepository) GetByKey(key string) (*settingsmodel.SystemConfig, error) {
	var config settingsmodel.SystemConfig
	err := r.db.Where("config_key = ?", key).First(&config).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &config, nil
}

func (r *systemConfigRepository) GetByKeys(keys []string) ([]settingsmodel.SystemConfig, error) {
	var configs []settingsmodel.SystemConfig
	err := r.db.Where("config_key IN ?", keys).Find(&configs).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get configs by keys: %w", err)
	}
	return configs, nil
}

func (r *systemConfigRepository) GetByCategory(ctx context.Context, category settingsmodel.ConfigCategory) ([]settingsmodel.SystemConfig, error) {
	var configs []settingsmodel.SystemConfig
	err := r.db.WithContext(ctx).Where("category = ?", category).
		Order("sort_order ASC, id ASC").
		Find(&configs).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []settingsmodel.SystemConfig{}, nil
		}
		return nil, fmt.Errorf("根据分类获取配置列表失败: %w", err)
	}
	return configs, nil
}

func (r *systemConfigRepository) GetPublicConfigs() ([]settingsmodel.SystemConfig, error) {
	var configs []settingsmodel.SystemConfig
	err := r.db.Where("is_public = ?", true).
		Order("sort_order ASC, id ASC").
		Find(&configs).Error
	return configs, err
}

func (r *systemConfigRepository) GetAll() ([]settingsmodel.SystemConfig, error) {
	var configs []settingsmodel.SystemConfig
	err := r.db.Order("sort_order ASC, id ASC").Find(&configs).Error
	return configs, err
}

func (r *systemConfigRepository) GetList(page, pageSize int, category string) ([]settingsmodel.SystemConfig, int64, error) {
	var configs []settingsmodel.SystemConfig
	var total int64

	query := r.db.Model(&settingsmodel.SystemConfig{})

	if category != "" {
		query = query.Where("category = ?", category)
	}

	offset := (page - 1) * pageSize
	err := query.Count(&total).
		Offset(offset).
		Limit(pageSize).
		Order("sort_order ASC, id ASC").
		Find(&configs).Error

	return configs, total, err
}

func (r *systemConfigRepository) Create(config *settingsmodel.SystemConfig) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return tx.Create(config).Error
	})
}

func (r *systemConfigRepository) Update(config *settingsmodel.SystemConfig) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return tx.Save(config).Error
	})
}

// BatchUpdate 批量更新配置值（事务 + CASE WHEN 单条 SQL）
func (r *systemConfigRepository) BatchUpdate(kvs map[string]string) error {
	if len(kvs) == 0 {
		return nil
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		// 构建 CASE WHEN SQL
		whens := make([]string, 0, len(kvs))
		keys := make([]string, 0, len(kvs))
		args := make([]interface{}, 0, len(kvs)*2)
		for key, value := range kvs {
			whens = append(whens, "WHEN ? THEN ?")
			args = append(args, key, value)
			keys = append(keys, key)
		}

		sql := fmt.Sprintf(
			"UPDATE system_configs SET value = CASE config_key %s END, updated_at = CURRENT_TIMESTAMP WHERE config_key IN (?)",
			strings.Join(whens, " "),
		)

		return tx.Exec(sql, append(args, keys)...).Error
	})
}

func (r *systemConfigRepository) UpdateValueByKey(key, value string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return tx.Model(&settingsmodel.SystemConfig{}).
			Where("config_key = ?", key).
			Update("value", value).Error
	})
}

func (r *systemConfigRepository) Delete(id uint64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return tx.Delete(&settingsmodel.SystemConfig{}, id).Error
	})
}

func (r *systemConfigRepository) ExistsByKey(key string) (bool, error) {
	var count int64
	err := r.db.Model(&settingsmodel.SystemConfig{}).
		Where("config_key = ?", key).
		Count(&count).Error
	return count > 0, err
}
