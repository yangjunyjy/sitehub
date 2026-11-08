package repository

import (
	rbacentity "sitehub/service/rbac/model/entity"
)

// SysMenuRepository 系统菜单仓储接口
type SysMenuRepository interface {
	InTransaction(func(SysMenuRepository) error) error
	FindByIDForUpdate(id uint64) (*rbacentity.SysMenu, error)
	// FindByID 根据ID查找菜单
	FindByID(id uint64) (*rbacentity.SysMenu, error)
	// FindByCode 根据菜单编码查找菜单
	FindByCode(code string) (*rbacentity.SysMenu, error)
	// FindByName 根据菜单名称查找菜单
	FindByName(name string) (*rbacentity.SysMenu, error)
	// FindByPath 根据菜单路径查找菜单
	FindByPath(path string) (*rbacentity.SysMenu, error)
	// List 查询菜单列表
	List(page, pageSize int) ([]*rbacentity.SysMenu, int64, error)
	// GetMenuTree 获取菜单树
	GetMenuTree() ([]*rbacentity.SysMenu, error)
	// GetUserMenus 获取用户的菜单列表
	GetUserMenus(userID uint64) ([]*rbacentity.SysMenu, error)
	// Create 创建菜单
	Create(menu *rbacentity.SysMenu) error
	// Update 更新菜单
	Update(menu *rbacentity.SysMenu) error
	// Delete 删除菜单
	Delete(id uint64) error
	// GetChildren 获取子菜单
	GetChildren(parentID uint64) ([]*rbacentity.SysMenu, error)
	// UpdateStatus 更新菜单状态
	UpdateStatus(id uint64, status rbacentity.Status) error
	// GetAllMenus 获取所有菜单
	GetAllMenus() ([]*rbacentity.SysMenu, error)
}
