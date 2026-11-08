package repository

import (
	rbacentity "sitehub/service/rbac/model/entity"
)

// IPBlacklistRepository IP黑名单仓储接口
type IPBlacklistRepository interface {
	Create(ip, reason string) error
	FindByID(id uint) (*rbacentity.IPBlacklist, error)
	FindByIP(ip string) (*rbacentity.IPBlacklist, error)
	List(offset, limit int) ([]rbacentity.IPBlacklist, error)
	FindByIsBlocked(isBlocked bool) ([]rbacentity.IPBlacklist, error)
	Delete(id uint) error
}
