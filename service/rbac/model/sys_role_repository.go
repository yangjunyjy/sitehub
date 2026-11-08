package repository

import (
	rbacrequest "sitehub/service/rbac/dto/request"
	rbacentity "sitehub/service/rbac/model/entity"
)

// SysRoleRepository 系统角色仓储接口
type SysRoleRepository interface {
	// FindByID 根据ID查找角色
	FindByID(id uint64) (*rbacentity.SysRole, error)
	// FindByName 根据角色名称查找角色
	FindByName(name string) (*rbacentity.SysRole, error)
	// List 查询角色列表
	List(query *rbacrequest.RoleSearchQueryRequest) ([]*rbacentity.SysRole, int64, error)
	// ListAll 查询所有角色
	ListAll() ([]*rbacentity.SysRole, error)
	// Create 创建角色
	Create(role *rbacentity.SysRole) error
	// Update 更新角色
	Update(role *rbacentity.SysRole) error
	// Delete 删除角色
	Delete(id uint64) error
	// AssignMenus 分配菜单给角色
	AssignMenus(roleID uint64, menuIDs []uint64) error
	// GetRoleMenus 获取角色的菜单列表
	GetRoleMenus(roleID uint64) ([]*rbacentity.SysMenu, error)
	// GetRoleUsers 获取角色的用户列表
	GetRoleUsers(roleID uint64) ([]*rbacentity.SysUser, error)
	// UpdateStatus 更新角色状态
	UpdateStatus(id uint64, status rbacentity.Status) error
}
