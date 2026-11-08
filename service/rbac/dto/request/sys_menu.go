package request

type DeleteMenuRequest struct {
	ID uint64 `json:"id" binding:"required"`
}

type CreateSysMenuRequest struct {
	Name        string `json:"name" binding:"required,max=50"`      // 必须
	Code        string `json:"code" binding:"max=100"`              // 按钮必填；目录和菜单可自动生成
	Type        uint8  `json:"type" binding:"oneof=1 2 3"`          // 必须
	Status      *uint8 `json:"status" binding:"required,oneof=0 1"` // 必须；指针区分缺失与禁用值 0
	ParentID    uint64 `json:"parent_id"`                           // 可选
	Sort        int    `json:"sort"`                                // 可选
	Component   string `json:"component" binding:"max=200"`         // 可选
	Path        string `json:"path" binding:"max=200"`
	Description string `json:"description" binding:"max=500"` // 可选
	IsVisible   bool   `json:"isVisible"`                     // 可选
	Icon        string `json:"icon" binding:"max=100"`        // 可选
}

type UpdateSysMenuRequest struct {
	ID          uint64 `json:"id" binding:"required"`
	Name        string `json:"name" binding:"required,max=50"`      // 必须
	Code        string `json:"code" binding:"max=100"`              // 目录和菜单留空时保留原编码
	Type        uint8  `json:"type" binding:"oneof=1 2 3"`          // 必须
	Status      *uint8 `json:"status" binding:"required,oneof=0 1"` // 必须；指针区分缺失与禁用值 0
	ParentID    uint64 `json:"parent_id"`                           // 可选
	Sort        int    `json:"sort"`                                // 可选
	Component   string `json:"component" binding:"max=200"`         // 可选
	Path        string `json:"path" binding:"max=200"`
	Description string `json:"description" binding:"max=500"` // 可选
	IsVisible   bool   `json:"isVisible"`                     // 可选
	Icon        string `json:"icon" binding:"max=100"`        // 可选
}
