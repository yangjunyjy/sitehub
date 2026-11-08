package mysql

import (
	"gorm.io/gorm"

	rbacrepository "sitehub/service/rbac/model"
	rbacentity "sitehub/service/rbac/model/entity"
)

// NewIPBlacklistRepository 创建IP黑名单仓储
func NewIPBlacklistRepository(db *gorm.DB) rbacrepository.IPBlacklistRepository {
	return &ipBlacklistRepository{db: db}
}

type ipBlacklistRepository struct {
	db *gorm.DB
}

func (r *ipBlacklistRepository) Create(ip, reason string) error {
	blacklist := rbacentity.IPBlacklist{
		IPAddress: ip,
		Reason:    reason,
		IsBlocked: true,
	}
	return r.db.Create(&blacklist).Error
}

func (r *ipBlacklistRepository) FindByID(id uint) (*rbacentity.IPBlacklist, error) {
	var blacklist rbacentity.IPBlacklist
	err := r.db.First(&blacklist, id).Error
	if err != nil {
		return nil, err
	}
	return &blacklist, nil
}

func (r *ipBlacklistRepository) FindByIP(ip string) (*rbacentity.IPBlacklist, error) {
	var blacklist rbacentity.IPBlacklist
	err := r.db.Where("ip_address = ?", ip).First(&blacklist).Error
	if err != nil {
		return nil, err
	}
	return &blacklist, nil
}

func (r *ipBlacklistRepository) List(offset, limit int) ([]rbacentity.IPBlacklist, error) {
	var blacklists []rbacentity.IPBlacklist
	err := r.db.Order("created_at DESC").Offset(offset).Limit(limit).Find(&blacklists).Error
	return blacklists, err
}

func (r *ipBlacklistRepository) FindByIsBlocked(isBlocked bool) ([]rbacentity.IPBlacklist, error) {
	var blacklists []rbacentity.IPBlacklist
	err := r.db.Where("is_blocked = ?", isBlocked).Order("created_at DESC").Find(&blacklists).Error
	return blacklists, err
}

func (r *ipBlacklistRepository) Delete(id uint) error {
	return r.db.Delete(&rbacentity.IPBlacklist{}, id).Error
}
