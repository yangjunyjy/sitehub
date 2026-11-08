package mysql

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"sitehub/common/pkg/logger"
	_ "sitehub/common/pkg/logger"
	rbacrequest "sitehub/service/rbac/dto/request"
	rbacresponse "sitehub/service/rbac/dto/response"
	rbacrepository "sitehub/service/rbac/model"
	rbacentity "sitehub/service/rbac/model/entity"
)

type SysUserRepositoryImpl struct {
	db *gorm.DB
}

func NewSysUserRepositoryImpl(db *gorm.DB) rbacrepository.SysUserRepository {
	return &SysUserRepositoryImpl{db: db}
}

func (r *SysUserRepositoryImpl) InTransaction(fn func(rbacrepository.SysUserRepository) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error { return fn(&SysUserRepositoryImpl{db: tx}) })
}

func (r *SysUserRepositoryImpl) FindByUsername(username string) (*rbacentity.SysUser, error) {
	var user rbacentity.SysUser
	err := r.db.Preload("Roles").Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *SysUserRepositoryImpl) FindByID(id uint64) (*rbacentity.SysUser, error) {
	var user rbacentity.SysUser
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *SysUserRepositoryImpl) FindByEmail(email string) (*rbacentity.SysUser, error) {
	var user rbacentity.SysUser
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *SysUserRepositoryImpl) List(query *rbacrequest.SearchSysUserQueryRequest) ([]*rbacresponse.UserListResponse, int64, error) {
	var users []*rbacentity.SysUser
	var total int64

	// 计算偏移量
	offset := (query.Page - 1) * query.PageSize

	// 构建基础查询
	db := r.db.Model(&rbacentity.SysUser{})

	// 应用查询条件
	if query.Username != "" {
		db = db.Where("username LIKE ?", "%"+query.Username+"%")
	}
	if query.Email != "" {
		db = db.Where("email LIKE ?", "%"+query.Email+"%")
	}
	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}

	// 统计总数（必须在分页之前）
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询，通过中间表 sys_user_role 预加载角色信息
	err := db.Preload("Roles").
		Offset(offset).
		Limit(query.PageSize).
		Order("id DESC").
		Find(&users).Error

	if err != nil {
		return nil, 0, err
	}

	// 将 models.SysUser 转换为 response.UserListResponse
	userResponses := make([]*rbacresponse.UserListResponse, len(users))
	for i, user := range users {
		userResponses[i] = &rbacresponse.UserListResponse{
			ID:         user.ID,
			Username:   user.Username,
			Nickname:   user.Nickname,
			Avatar:     user.Avatar,
			Email:      user.Email,
			Gender:     user.Gender,
			Status:     user.Status,
			IsAdmin:    user.IsAdmin,
			CreateTime: *user.CreateTime,
			Roles:      user.Roles,
		}
	}

	return userResponses, total, nil
}

func (r *SysUserRepositoryImpl) Create(user *rbacentity.SysUser) error {
	return r.db.Omit(clause.Associations).Create(user).Error
}

func (r *SysUserRepositoryImpl) Update(user *rbacentity.SysUser) error {
	// 登录更新会携带已加载的 Roles；普通用户保存不得再次写入关联表。
	// 角色关系只由显式的分配入口维护。
	return r.db.Omit(clause.Associations).Save(user).Error
}

func (r *SysUserRepositoryImpl) Delete(id uint64) error {
	return r.db.Delete(&rbacentity.SysUser{}, id).Error
}

func (r *SysUserRepositoryImpl) GetUserRoles(userID uint64) ([]*rbacentity.SysRole, error) {
	var roles []*rbacentity.SysRole

	// 方法2：通过关联表直接查询用户的所有角色
	err := r.db.
		Distinct("sys_role.*").
		Joins("JOIN sys_user_role sur ON sur.role_id = sys_role.id").
		Where("sur.user_id = ?", userID).
		Preload("Menus"). // 预加载每个角色的菜单
		Find(&roles).Error

	if err != nil {
		return nil, err
	}
	// logger.Info("repo层查询到的角色信息", logger.Any("roles", roles))
	return roles, nil
}

func (r *SysUserRepositoryImpl) GetUserPosts(userID uint64) ([]*rbacentity.SysPost, error) {
	var posts []*rbacentity.SysPost
	err := r.db.Table("sys_post").
		Joins("INNER JOIN sys_user_post ON sys_post.id = sys_user_post.post_id").
		Where("sys_user_post.user_id = ?", userID).
		Find(&posts).Error
	if err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *SysUserRepositoryImpl) AssignRoles(userID uint64, roleIDs []uint64) error {
	// 在服务开启的事务中锁定父记录；相同用户的分配串行执行。
	var user rbacentity.SysUser
	if err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id").First(&user, userID).Error; err != nil {
		return err
	}
	if len(roleIDs) > 0 {
		var roles []rbacentity.SysRole
		if err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id").Where("id IN ?", roleIDs).Order("id").Find(&roles).Error; err != nil {
			return err
		}
		if len(roles) != len(roleIDs) {
			return gorm.ErrRecordNotFound
		}
	}
	logger.Info("分配角色", logger.Any("userID", userID), logger.Any("roleIDs", roleIDs))

	// 先删除现有角色
	if err := r.db.Where("user_id = ?", userID).Delete(&rbacentity.SysUserRole{}).Error; err != nil {
		logger.Error("删除用户角色失败", logger.Err("error", err))
		return err
	}

	// 批量插入新角色
	if len(roleIDs) == 0 {
		logger.Info("角色ID列表为空,跳过插入")
		return nil
	}

	var userRoles []rbacentity.SysUserRole
	for _, roleID := range roleIDs {
		userRoles = append(userRoles, rbacentity.SysUserRole{
			UserID: userID,
			RoleID: roleID,
		})
	}

	logger.Info("准备插入用户角色", logger.Any("count", len(userRoles)), logger.Any("data", userRoles))
	err := r.db.CreateInBatches(userRoles, 100).Error
	if err != nil {
		logger.Error("插入用户角色失败", logger.Err("error", err))
	}
	return err
}

func (r *SysUserRepositoryImpl) AssignPosts(userID uint64, postIDs []uint64) error {
	// 先删除现有岗位
	if err := r.db.Where("user_id = ?", userID).Delete(&rbacentity.SysUserPost{}).Error; err != nil {
		return err
	}

	// 批量插入新岗位
	if len(postIDs) == 0 {
		return nil
	}

	var userPosts []rbacentity.SysUserPost
	for _, postID := range postIDs {
		userPosts = append(userPosts, rbacentity.SysUserPost{
			UserID: userID,
			PostID: postID,
		})
	}

	return r.db.CreateInBatches(userPosts, 100).Error
}

func (r *SysUserRepositoryImpl) UpdateStatus(id uint64, status rbacentity.Status) error {
	return r.db.Model(&rbacentity.SysUser{}).Where("id = ?", id).Update("status", status).Error
}

func (r *SysUserRepositoryImpl) ResetPassword(id uint64, newPassword string) error {
	return r.db.Model(&rbacentity.SysUser{}).Where("id = ?", id).Update("password", newPassword).Error
}
