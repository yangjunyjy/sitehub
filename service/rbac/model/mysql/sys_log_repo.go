package mysql

import (
	"time"

	"gorm.io/gorm"

	rbacrepository "sitehub/service/rbac/model"
	rbacentity "sitehub/service/rbac/model/entity"
)

type SysLogRepositoryImpl struct {
	db *gorm.DB
}

func NewSysLogRepositoryImpl(db *gorm.DB) rbacrepository.SysLogRepository {
	return &SysLogRepositoryImpl{db: db}
}

// 操作日志
func (r *SysLogRepositoryImpl) CreateOperationLog(log *rbacentity.SysOperationLog) error {
	return r.db.Create(log).Error
}

func (r *SysLogRepositoryImpl) ListOperationLogs(page, pageSize int) ([]*rbacentity.SysOperationLog, int64, error) {
	var logs []*rbacentity.SysOperationLog
	var total int64

	offset := (page - 1) * pageSize
	err := r.db.Model(&rbacentity.SysOperationLog{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Offset(offset).Limit(pageSize).Order("id DESC").Find(&logs).Error
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *SysLogRepositoryImpl) DeleteOperationLogs(ids []uint64) error {
	return r.db.Delete(&rbacentity.SysOperationLog{}, ids).Error
}

// 登录日志
func (r *SysLogRepositoryImpl) CreateLoginLog(log *rbacentity.SysLoginLog) error {
	return r.db.Create(log).Error
}

func (r *SysLogRepositoryImpl) ListLoginLogs(page, pageSize int) ([]*rbacentity.SysLoginLog, int64, error) {
	var logs []*rbacentity.SysLoginLog
	var total int64

	offset := (page - 1) * pageSize
	err := r.db.Model(&rbacentity.SysLoginLog{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Offset(offset).Limit(pageSize).Order("id DESC").Find(&logs).Error
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *SysLogRepositoryImpl) DeleteLoginLogs(ids []uint64) error {
	return r.db.Delete(&rbacentity.SysLoginLog{}, ids).Error
}

// 清理日志
func (r *SysLogRepositoryImpl) ClearOldOperationLogs(days int) error {
	beforeDate := time.Now().AddDate(0, 0, -days)
	return r.db.Where("created_at < ?", beforeDate).Delete(&rbacentity.SysOperationLog{}).Error
}

func (r *SysLogRepositoryImpl) ClearOldLoginLogs(days int) error {
	beforeDate := time.Now().AddDate(0, 0, -days)
	return r.db.Where("created_at < ?", beforeDate).Delete(&rbacentity.SysLoginLog{}).Error
}
