package mysql

import (
	"time"

	"gorm.io/gorm"

	rbacrequest "sitehub/service/rbac/dto/request"
	rbacrepository "sitehub/service/rbac/model"
	rbacentity "sitehub/service/rbac/model/entity"
)

type SysRoleRepositoryImpl struct {
	db *gorm.DB
}

func NewSysRoleRepositoryImpl(db *gorm.DB) rbacrepository.SysRoleRepository {
	return &SysRoleRepositoryImpl{db: db}
}

func (r *SysRoleRepositoryImpl) FindByID(id uint64) (*rbacentity.SysRole, error) {
	var role rbacentity.SysRole
	err := r.db.First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *SysRoleRepositoryImpl) FindByName(name string) (*rbacentity.SysRole, error) {
	var role rbacentity.SysRole
	err := r.db.Where("name = ?", name).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *SysRoleRepositoryImpl) List(query *rbacrequest.RoleSearchQueryRequest) ([]*rbacentity.SysRole, int64, error) {
	var roles []*rbacentity.SysRole
	var total int64

	// 构建查询条件
	db := r.db.Model(&rbacentity.SysRole{})

	// 应用查询条件
	if query.RoleName != "" {
		db = db.Where("role_name LIKE ?", "%"+query.RoleName+"%") // 通常支持模糊查询
	}
	if query.RoleCode != "" {
		db = db.Where("role_code like ?", "%"+query.RoleCode+"%")
	}
	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}

	// 先统计总数（必须在分页之前）
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (query.Page - 1) * query.PageSize
	if err := db.Order("id ASC").
		Offset(offset).
		Limit(query.PageSize).
		Find(&roles).Error; err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}

func (r *SysRoleRepositoryImpl) ListAll() ([]*rbacentity.SysRole, error) {
	var roles []*rbacentity.SysRole
	err := r.db.Where("status = ?", rbacentity.StatusEnabled).Find(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *SysRoleRepositoryImpl) Create(role *rbacentity.SysRole) error {
	return r.db.Create(role).Error
}

func (r *SysRoleRepositoryImpl) Update(role *rbacentity.SysRole) error {
	return r.db.Save(role).Error
}

func (r *SysRoleRepositoryImpl) Delete(id uint64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 删除角色
		if err := tx.Delete(&rbacentity.SysRole{}, id).Error; err != nil {
			return err
		}
		// 删除角色菜单关联
		if err := tx.Where("role_id = ?", id).Delete(&rbacentity.SysRoleMenu{}).Error; err != nil {
			return err
		}
		// 删除用户角色关联
		if err := tx.Where("role_id = ?", id).Delete(&rbacentity.SysUserRole{}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *SysRoleRepositoryImpl) AssignMenus(roleID uint64, menuIDs []uint64) error {
	// 先删除现有菜单
	if err := r.db.Where("role_id = ?", roleID).Delete(&rbacentity.SysRoleMenu{}).Error; err != nil {
		return err
	}

	// 批量插入新菜单
	if len(menuIDs) == 0 {
		return nil
	}

	now := time.Now()
	var roleMenus []rbacentity.SysRoleMenu
	for _, menuID := range menuIDs {
		roleMenus = append(roleMenus, rbacentity.SysRoleMenu{
			RoleID:    roleID,
			MenuID:    menuID,
			CreatedAt: now,
		})
	}

	return r.db.CreateInBatches(roleMenus, 100).Error
}

func (r *SysRoleRepositoryImpl) GetRoleMenus(roleID uint64) ([]*rbacentity.SysMenu, error) {
	var role rbacentity.SysRole
	err := r.db.Preload("Menus").First(&role, roleID).Error
	if err != nil {
		return nil, err
	}
	return role.Menus, nil
}

func (r *SysRoleRepositoryImpl) GetRoleUsers(roleID uint64) ([]*rbacentity.SysUser, error) {
	var users []*rbacentity.SysUser
	err := r.db.Table("sys_user").
		Joins("INNER JOIN sys_user_role ON sys_user.id = sys_user_role.user_id").
		Where("sys_user_role.role_id = ?", roleID).
		Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *SysRoleRepositoryImpl) UpdateStatus(id uint64, status rbacentity.Status) error {
	return r.db.Model(&rbacentity.SysRole{}).Where("id = ?", id).Update("status", status).Error
}
