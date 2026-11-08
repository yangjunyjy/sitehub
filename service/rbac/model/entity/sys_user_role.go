package entity

// SysUserRole 用户角色关联表
type SysUserRole struct {
	ID     uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID uint64 `gorm:"column:user_id;type:bigint;not null;index:sys_user_role_idx_user_id,priority:1;uniqueIndex:uk_sys_user_role,priority:1;comment:用户ID" json:"user_id"`
	RoleID uint64 `gorm:"column:role_id;type:bigint;not null;index:sys_user_role_idx_role_id,priority:2;uniqueIndex:uk_sys_user_role,priority:2;comment:角色ID" json:"role_id"`
}

// TableName 指定表名
func (SysUserRole) TableName() string {
	return "sys_user_role"
}
