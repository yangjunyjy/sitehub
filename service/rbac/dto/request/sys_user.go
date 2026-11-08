package request

type DeleteSysUserRequest struct {
	ID uint64 `json:"id" binding:"required"`
}

type CreateSysUserRequest struct {
	Username string `json:"username" binding:"required"`
	Nickname string `json:"nickname"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Status   uint8  `json:"status" binding:"required,min=0,max=1"`
	Gender   uint8  `json:"gender"`
}

type UpdateSysUserRequest struct {
	Id       uint64 `json:"id" binding:"required"`
	UserName string `json:"username" binding:"required"`
	Nickname string `json:"nickname"`
	Password string `json:"password"` // 留空保留原密码，列表编辑不会回传密码摘要。
	Email    string `json:"email" binding:"required"`
	Status   uint8  `json:"status" binding:"oneof=0 1"`
	Gender   uint8  `json:"gender"`
}

type AssignRolesRequest struct {
	UserID  uint64   `json:"userId" binding:"required,gt=0"`
	RoleIDs []uint64 `json:"roleIds" binding:"required,max=200,dive,gt=0"`
}

type SearchSysUserQueryRequest struct {
	Username string `form:"username"`
	Email    string `form:"email"`
	Status   *uint8 `form:"status" binding:"omitempty,oneof=0 1"` // nil means all; zero means disabled.
	Page     int    `form:"page" binding:"gte=1"`
	PageSize int    `form:"pageSize" binding:"gte=1,lte=100"`
}

type UpdatePasswordRequest struct {
	Email       string `json:"email" binding:"required"`
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}
