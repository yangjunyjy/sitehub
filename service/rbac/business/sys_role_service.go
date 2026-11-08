package services

import (
	errorx "sitehub/common/errorx"
	"sitehub/common/pkg/e"
	rbacrequest "sitehub/service/rbac/dto/request"
	rbacrepository "sitehub/service/rbac/model"
	rbacentity "sitehub/service/rbac/model/entity"
)

// SysRoleService 角色服务
type SysRoleService interface {
	// List 列表
	List(query *rbacrequest.RoleSearchQueryRequest) ([]*rbacentity.SysRole, int64, error)
	// ListAll 所有角色
	ListAll() ([]*rbacentity.SysRole, error)
	// GetByID 获取详情
	GetByID(id uint64) (*rbacentity.SysRole, error)
	// Create 创建
	Create(role *rbacentity.SysRole) error
	// Update 更新
	Update(role *rbacentity.SysRole) error
	// Delete 删除
	Delete(id uint64) error
	// AssignMenus 分配菜单
	AssignMenus(req *rbacrequest.RoleAssignMenusRequest) error
	// GetRoleMenus 获取角色菜单
	GetRoleMenus(roleID uint64) ([]*rbacentity.SysMenu, error)
}

type sysRoleService struct {
	roleRepo rbacrepository.SysRoleRepository
}

func NewSysRoleService(roleRepo rbacrepository.SysRoleRepository) SysRoleService {
	return &sysRoleService{roleRepo: roleRepo}
}

func (s *sysRoleService) List(query *rbacrequest.RoleSearchQueryRequest) ([]*rbacentity.SysRole, int64, error) {
	if query.Page <= 0 || query.PageSize <= 0 {
		return nil, 0, e.ERROR_PARAMS
	}
	return s.roleRepo.List(query)
}

func (s *sysRoleService) ListAll() ([]*rbacentity.SysRole, error) {
	return s.roleRepo.ListAll()
}

func (s *sysRoleService) GetByID(id uint64) (*rbacentity.SysRole, error) {
	return s.roleRepo.FindByID(id)
}

func (s *sysRoleService) Create(role *rbacentity.SysRole) error {
	return s.roleRepo.Create(role)
}

func (s *sysRoleService) Update(role *rbacentity.SysRole) error {
	return s.roleRepo.Update(role)
}

func (s *sysRoleService) Delete(id uint64) error {
	// 不能删除超级管理员角色
	if id == rbacentity.SuperAdminRoleID {
		return errorx.Conflicting("不能删除超级管理员角色")
	}
	return s.roleRepo.Delete(id)
}

func (s *sysRoleService) AssignMenus(req *rbacrequest.RoleAssignMenusRequest) error {
	return s.roleRepo.AssignMenus(req.RoleID, req.MenuIDs)
}

func (s *sysRoleService) GetRoleMenus(roleID uint64) ([]*rbacentity.SysMenu, error) {
	return s.roleRepo.GetRoleMenus(roleID)
}
