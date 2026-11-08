// Package services provides system configuration business logic
package services

import (
	"context"
	"encoding/json"
	"fmt"

	errorx "sitehub/common/errorx"
	"sitehub/common/pkg/logger"
	settingscontract "sitehub/common/settings/contract"
	settingsmodel "sitehub/common/settings/model"
	settingsrepository "sitehub/common/settings/repository"
)

// SystemConfigService 系统配置服务
type SystemConfigService interface {
	// UpdateJSONConfigIfUnchanged commits only the snapshot validated by the caller.
	UpdateJSONConfigIfUnchanged(key, expected, value string) error
	ValidateConfigs(configs []settingscontract.BatchUpdateConfig) error

	// GetByKey 根据键获取配置值（返回类型化的值）
	GetByKey(key string) (interface{}, error)

	// GetByKeyString 获取字符串配置值
	GetByKeyString(key string) (string, error)

	// GetByKeyInt 获取整数配置值
	GetByKeyInt(key string) (int, error)

	// GetByKeyBool 获取布尔配置值
	GetByKeyBool(key string) (bool, error)

	// GetConfigList 获取配置列表
	GetConfigList(page, pageSize int, category string) ([]settingsmodel.SystemConfig, int64, error)

	// GetConfigByCategory 根据分类获取配置
	GetConfigByCategory(ctx context.Context, category settingsmodel.ConfigCategory) ([]settingsmodel.SystemConfig, error)

	// GetPublicConfigs 获取公开配置
	GetPublicConfigs() ([]settingsmodel.SystemConfig, error)

	// CreateConfig 创建配置
	CreateConfig(config *settingsmodel.SystemConfig) error

	// UpdateConfig 更新配置
	UpdateConfig(config *settingsmodel.SystemConfig) error

	// BatchUpdateConfigs 批量更新配置
	BatchUpdateConfigs(configs []settingscontract.BatchUpdateConfig) error

	// DeleteConfig 删除配置
	DeleteConfig(id uint64) error

	// GetValueByKey 获取配置的原始值
	GetConfig(key string) (*settingsmodel.SystemConfig, error)

	// ExportConfigs 导出配置为 JSON
	ExportConfigs() ([]byte, error)

	// ImportConfigs 从 JSON 导入配置
	ImportConfigs(configs []settingsmodel.SystemConfig) error
}

type systemConfigService struct {
	repo settingsrepository.SystemConfigRepository
}

// NewSystemConfigService 创建系统配置服务
func NewSystemConfigService(repo settingsrepository.SystemConfigRepository) SystemConfigService {
	return &systemConfigService{repo: repo}
}

func (s *systemConfigService) UpdateJSONConfigIfUnchanged(key, expected, value string) error {
	if err := s.validateConfigValue(value, settingsmodel.ConfigTypeJSON); err != nil {
		return err
	}
	matched, err := s.repo.CompareAndSwapValue(key, expected, value)
	if err != nil {
		return err
	}
	if !matched {
		return errorx.Conflicting("配置已被其他操作更新，请重新加载后重试")
	}
	return nil
}

func (s *systemConfigService) GetByKey(key string) (interface{}, error) {
	config, err := s.repo.GetByKey(key)
	if err != nil || config == nil {
		return nil, fmt.Errorf("配置不存在: %s", key)
	}

	return s.parseValue(config.Value, config.ConfigType)
}

func (s *systemConfigService) GetByKeyString(key string) (string, error) {
	config, err := s.repo.GetByKey(key)
	if err != nil || config == nil {
		return "", fmt.Errorf("配置不存在: %s", key)
	}

	value, err := s.parseValue(config.Value, config.ConfigType)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%v", value), nil
}

func (s *systemConfigService) GetByKeyInt(key string) (int, error) {
	value, err := s.GetByKey(key)
	if err != nil {
		return 0, err
	}

	switch v := value.(type) {
	case float64:
		return int(v), nil
	case int:
		return v, nil
	case int64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("配置值不是整数类型")
	}
}

func (s *systemConfigService) GetByKeyBool(key string) (bool, error) {
	value, err := s.GetByKey(key)
	if err != nil {
		return false, err
	}

	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		return v == "true" || v == "1", nil
	default:
		return false, fmt.Errorf("配置值不是布尔类型")
	}
}

func (s *systemConfigService) GetConfigList(page, pageSize int, category string) ([]settingsmodel.SystemConfig, int64, error) {
	return s.repo.GetList(page, pageSize, category)
}

func (s *systemConfigService) GetConfigByCategory(ctx context.Context, category settingsmodel.ConfigCategory) ([]settingsmodel.SystemConfig, error) {
	if category == "" {
		logger.Info("配置分类为空")
		return []settingsmodel.SystemConfig{}, nil
	}
	configs, err := s.repo.GetByCategory(ctx, category)
	if err != nil {
		return nil, err
	}
	// password 类型脱敏，避免密钥泄露
	for i := range configs {
		if configs[i].ConfigType == settingsmodel.ConfigTypePassword && configs[i].Value != "" {
			configs[i].Value = "********"
		}
	}
	return configs, nil
}

func (s *systemConfigService) GetPublicConfigs() ([]settingsmodel.SystemConfig, error) {
	configs, err := s.repo.GetPublicConfigs()
	if err != nil {
		return nil, err
	}
	visible := make([]settingsmodel.SystemConfig, 0, len(configs))
	for _, item := range configs {
		// Legacy public flags must not reveal raw artifact URLs or storage credentials.
		if item.Key != "website.downloads" && item.Key != "website.storage" {
			visible = append(visible, item)
		}
	}
	return visible, nil
}

func (s *systemConfigService) CreateConfig(config *settingsmodel.SystemConfig) error {
	// 检查键是否已存在
	exists, err := s.repo.ExistsByKey(config.Key)
	if err != nil {
		return fmt.Errorf("检查配置键失败: %w", err)
	}
	if exists {
		return fmt.Errorf("配置键 %s 已存在", config.Key)
	}

	// 验证配置值
	if err := s.validateConfigValue(config.Value, config.ConfigType); err != nil {
		return fmt.Errorf("配置值验证失败: %w", err)
	}

	return s.repo.Create(config)
}

func (s *systemConfigService) UpdateConfig(config *settingsmodel.SystemConfig) error {
	// 验证配置值
	if err := s.validateConfigValue(config.Value, config.ConfigType); err != nil {
		return fmt.Errorf("配置值验证失败: %w", err)
	}

	return s.repo.Update(config)
}

func (s *systemConfigService) BatchUpdateConfigs(configs []settingscontract.BatchUpdateConfig) error {
	if err := s.ValidateConfigs(configs); err != nil {
		return err
	}

	// 获取这些配置的类型，用于判断 password 跳过逻辑
	keys := make([]string, 0, len(configs))
	for _, c := range configs {
		keys = append(keys, c.Key)
	}
	existingConfigs, err := s.repo.GetByKeys(keys)
	if err != nil {
		return fmt.Errorf("获取配置失败: %w", err)
	}
	configTypeMap := make(map[string]settingsmodel.ConfigType, len(existingConfigs))
	for _, cfg := range existingConfigs {
		configTypeMap[cfg.Key] = cfg.ConfigType
	}

	// 转换为 map[key]value，password 类型如果值为空或脱敏占位符则跳过
	kvs := make(map[string]string, len(configs))
	for _, c := range configs {
		ct := configTypeMap[c.Key]
		if ct == settingsmodel.ConfigTypePassword && (c.Value == "" || c.Value == "********") {
			continue // 不更新空的或脱敏占位的密码
		}
		kvs[c.Key] = c.Value
	}
	return s.repo.BatchUpdate(kvs)
}

func (s *systemConfigService) ValidateConfigs(configs []settingscontract.BatchUpdateConfig) error {
	keys := make([]string, 0, len(configs))
	for _, config := range configs {
		keys = append(keys, config.Key)
	}
	systemconfig, err := s.repo.GetByKeys(keys)
	if err != nil {
		return fmt.Errorf("获取配置失败: %w", err)
	}

	var configMap = make(map[string]settingsmodel.ConfigType, len(systemconfig))
	for _, cfg := range systemconfig {
		configMap[cfg.Key] = cfg.ConfigType
	}

	for _, config := range configs {
		if err := s.validateConfigValue(config.Value, configMap[config.Key]); err != nil {
			return fmt.Errorf("配置 %s 值验证失败: %w", config.Key, err)
		}
	}
	return nil
}

func (s *systemConfigService) DeleteConfig(id uint64) error {
	return s.repo.Delete(id)
}

func (s *systemConfigService) GetConfig(key string) (*settingsmodel.SystemConfig, error) {
	return s.repo.GetByKey(key)
}

func (s *systemConfigService) ExportConfigs() ([]byte, error) {
	configs, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(configs, "", "  ")
}

func (s *systemConfigService) ImportConfigs(configs []settingsmodel.SystemConfig) error {
	kvs := make(map[string]string, len(configs))
	for _, c := range configs {
		kvs[c.Key] = c.Value
	}
	return s.repo.BatchUpdate(kvs)
}

// parseValue 解析配置值
func (s *systemConfigService) parseValue(value string, configType settingsmodel.ConfigType) (interface{}, error) {
	switch configType {
	case settingsmodel.ConfigTypeString, settingsmodel.ConfigTypePassword:
		return value, nil
	case settingsmodel.ConfigTypeInt:
		var result int
		_, err := fmt.Sscanf(value, "%d", &result)
		return result, err
	case settingsmodel.ConfigTypeFloat:
		var result float64
		_, err := fmt.Sscanf(value, "%f", &result)
		return result, err
	case settingsmodel.ConfigTypeBool:
		return value == "true" || value == "1", nil
	case settingsmodel.ConfigTypeJSON:
		var result map[string]interface{}
		err := json.Unmarshal([]byte(value), &result)
		return result, err
	default:
		return value, nil
	}
}

// validateConfigValue 验证配置值
func (s *systemConfigService) validateConfigValue(value string, configType settingsmodel.ConfigType) error {
	if value == "" {
		return nil // 空值不验证，由业务层判断
	}

	_, err := s.parseValue(value, configType)
	return err
}
