package schema

import (
	settingsmodel "sitehub/common/settings/model"
	domain "sitehub/service/rbac/domain"
	entity "sitehub/service/rbac/model/entity"

	"gorm.io/gorm"
)

// Apply creates only this service's schema. It is invoked by the explicit bootstrap tool.
func Apply(db *gorm.DB) error {
	for _, join := range []struct {
		model any
		field string
		join  any
	}{{&entity.SysUser{}, "Roles", &entity.SysUserRole{}}, {&entity.SysRole{}, "Users", &entity.SysUserRole{}}, {&entity.SysRole{}, "Menus", &entity.SysRoleMenu{}}, {&entity.SysMenu{}, "Roles", &entity.SysRoleMenu{}}, {&entity.SysUser{}, "Posts", &entity.SysUserPost{}}} {
		if err := db.SetupJoinTable(join.model, join.field, join.join); err != nil {
			return err
		}
	}
	return db.AutoMigrate(&entity.SysSystem{}, &entity.SysUser{}, &entity.SysRole{}, &entity.SysMenu{}, &entity.SysDept{}, &entity.SysPost{}, &entity.SysUserRole{}, &entity.SysRoleMenu{}, &entity.SysUserPost{}, &entity.SysLoginLog{}, &entity.SysOperationLog{}, &entity.IPBlacklist{}, &domain.Member{}, &domain.MemberVerification{}, &domain.MemberSession{}, &settingsmodel.SystemConfig{})
}
