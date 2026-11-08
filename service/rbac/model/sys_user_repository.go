package repository

import (
	rbacrequest "sitehub/service/rbac/dto/request"
	rbacresponse "sitehub/service/rbac/dto/response"
	rbacentity "sitehub/service/rbac/model/entity"
	// SysUserRepository 系统用户仓储接口
)

type SysUserRepository interface {
	// InTransaction 使服务编排的关系替换共享一个数据库事务。
	InTransaction(func(SysUserRepository) error) error
	// FindByUsername 根据用户名查找用户
	FindByUsername(username string) (*rbacentity.SysUser, error)
	// FindByID 根据ID查找用户
	FindByID(id uint64) (*rbacentity.SysUser, error)
	// FindByEmail 根据邮箱查找用户
	FindByEmail(email string) (*rbacentity.SysUser, error)
	// List 查询用户列表
	List(query *rbacrequest.SearchSysUserQueryRequest) ([]*rbacresponse.UserListResponse, int64, error)
	// Create 创建用户
	Create(user *rbacentity.SysUser) error
	// Update 更新用户
	Update(user *rbacentity.SysUser) error
	// Delete 删除用户（软删除）
	Delete(id uint64) error
	// GetUserRoles 获取用户的角色列表
	GetUserRoles(userID uint64) ([]*rbacentity.SysRole, error)
	// GetUserPosts 获取用户的岗位列表
	GetUserPosts(userID uint64) ([]*rbacentity.SysPost, error)
	// AssignRoles 分配角色给用户
	AssignRoles(userID uint64, roleIDs []uint64) error
	// AssignPosts 分配岗位给用户
	AssignPosts(userID uint64, postIDs []uint64) error
	// UpdateStatus 更新用户状态
	UpdateStatus(id uint64, status rbacentity.Status) error
	// ResetPassword 重置密码
	ResetPassword(id uint64, newPassword string) error
	// 更新密码
}
