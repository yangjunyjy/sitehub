package repository

import (
	rbacentity "sitehub/service/rbac/model/entity"
)

// SysLogRepository 系统日志仓储接口
type SysLogRepository interface {
	// 操作日志
	CreateOperationLog(log *rbacentity.SysOperationLog) error
	ListOperationLogs(page, pageSize int) ([]*rbacentity.SysOperationLog, int64, error)
	DeleteOperationLogs(ids []uint64) error

	// 登录日志
	CreateLoginLog(log *rbacentity.SysLoginLog) error
	ListLoginLogs(page, pageSize int) ([]*rbacentity.SysLoginLog, int64, error)
	DeleteLoginLogs(ids []uint64) error

	// 清理日志
	ClearOldOperationLogs(days int) error
	ClearOldLoginLogs(days int) error
}
