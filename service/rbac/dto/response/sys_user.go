package response

import (
	"time"

	rbacentity "sitehub/service/rbac/model/entity"
)

type UserListResponse struct {
	ID         uint64                `json:"id"`
	Username   string                `json:"username"`
	Nickname   string                `json:"nickname"`
	Avatar     string                `json:"avatar"`
	Email      string                `json:"email"`
	Gender     int8                  `json:"gender"`
	Status     int8                  `json:"status"`
	IsAdmin    int8                  `json:"isAdmin"`
	CreateTime time.Time             `json:"createTime"`
	Roles      []*rbacentity.SysRole `json:"roles"`
}
