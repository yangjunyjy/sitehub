// Package models defines the data models for system configuration
package model

import (
	"time"
)

// ConfigCategory 配置分类
type ConfigCategory string

const (
	ConfigCategorySystem   ConfigCategory = "system"   // 系统配置
	ConfigCategoryEmail    ConfigCategory = "email"    // 邮件配置
	ConfigCategoryDatabase ConfigCategory = "database" // 数据库配置
	ConfigCategoryRedis    ConfigCategory = "redis"    // Redis 配置
	ConfigCategoryAuth     ConfigCategory = "auth"     // 认证配置
	ConfigCategoryFile     ConfigCategory = "file"     // 文件上传配置
	ConfigCategoryApp      ConfigCategory = "app"      // 应用配置
	ConfigCategoryContent  ConfigCategory = "content"  // 内容配置
	ConfigCategoryLog      ConfigCategory = "log"      // 日志配置
	ConfigCategoryUpload   ConfigCategory = "upload"   // 上传配置
)

// ConfigType 配置值类型
type ConfigType string

const (
	ConfigTypeString   ConfigType = "string"   // 字符串
	ConfigTypeInt      ConfigType = "int"      // 整数
	ConfigTypeFloat    ConfigType = "float"    // 浮点数
	ConfigTypeBool     ConfigType = "bool"     // 布尔值
	ConfigTypeJSON     ConfigType = "json"     // JSON 对象
	ConfigTypePassword ConfigType = "password" // 密码（加密存储）
)

// SystemConfig 系统配置模型
type SystemConfig struct {
	ID          uint64         `gorm:"column:id;primaryKey" json:"id"`
	Key         string         `gorm:"column:config_key;type:varchar(100);uniqueIndex;not null" json:"key"`             // 配置键
	Value       string         `gorm:"type:longtext" json:"value"`                                                      // 可存储大型站点配置，产品文档正文使用独立表
	Category    ConfigCategory `gorm:"column:category;type:varchar(50)" json:"category"`                                // 配置分类
	ConfigType  ConfigType     `gorm:"column:config_type;type:varchar(20);not null;default:'string'" json:"configType"` // 配置值类型
	Label       string         `gorm:"column:label;type:varchar(255)" json:"label"`                                     // 配置名称/标签
	Description string         `gorm:"column:description;type:text" json:"description"`                                 // 配置描述
	IsPublic    bool           `gorm:"column:is_public;default:false" json:"isPublic"`                                  // 是否公开（前端可访问）
	IsEditable  bool           `gorm:"column:is_editable;default:true" json:"isEditable"`                               // 是否可编辑
	IsRequired  bool           `gorm:"column:is_required;default:false" json:"isRequired"`                              // 是否必填
	Options     *string        `gorm:"column:options;type:json" json:"options"`                                         // 可选值（JSON 数组）
	Validation  string         `gorm:"column:validation;type:varchar(255)" json:"validation"`                           // 校验规则
	SortOrder   int            `gorm:"column:sort_order;default:0" json:"sortOrder"`                                    // 排序
	CreatedBy   string         `gorm:"column:created_by;type:varchar(100)" json:"createdBy"`
	UpdatedBy   string         `gorm:"column:updated_by;type:varchar(100)" json:"updatedBy"`
	CreatedAt   time.Time      `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

// TableName 指定表名
func (SystemConfig) TableName() string {
	return "system_configs"
}
