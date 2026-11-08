// Package repository defines the interface for system config data access
package repository

import (
	"context"

	settingsmodel "sitehub/common/settings/model"
)

// SystemConfigRepository 系统配置仓储接口
type SystemConfigRepository interface {
	// CompareAndSwapValue rechecks an existing configuration under a row lock.
	// False means the row disappeared or changed; it never recreates deleted rows.
	CompareAndSwapValue(key, expected, value string) (bool, error)
	// MutateValue locks one configuration and commits the callback's value atomically.
	// The callback must contain only validation and in-memory transformation.
	MutateValue(initial settingsmodel.SystemConfig, change func(string) (string, error)) error
	// GetByKey 根据键获取配置
	GetByKey(key string) (*settingsmodel.SystemConfig, error)

	// GetByCategory 根据分类获取配置列表
	GetByCategory(ctx context.Context, category settingsmodel.ConfigCategory) ([]settingsmodel.SystemConfig, error)

	GetByKeys(keys []string) ([]settingsmodel.SystemConfig, error)

	// GetPublicConfigs 获取公开配置
	GetPublicConfigs() ([]settingsmodel.SystemConfig, error)

	// GetAll 获取所有配置
	GetAll() ([]settingsmodel.SystemConfig, error)

	// GetList 分页获取配置列表
	GetList(page, pageSize int, category string) ([]settingsmodel.SystemConfig, int64, error)

	// Create 创建配置
	Create(config *settingsmodel.SystemConfig) error

	// Update 更新配置
	Update(config *settingsmodel.SystemConfig) error

	// Delete 删除配置
	Delete(id uint64) error

	// BatchUpdate 批量更新配置（按 key 更新 value）
	BatchUpdate(kvs map[string]string) error

	// UpdateValueByKey 更新配置值
	UpdateValueByKey(key, value string) error

	// ExistsByKey 检查键是否存在
	ExistsByKey(key string) (bool, error)
}
