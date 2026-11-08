package contract

import "sitehub/common/errorx"

// Storage errors belong to file. Callers map them at their business boundary.
var (
	ErrInvalidPath = &errorx.PublicError{Kind: errorx.InvalidInput, Message: "文件路径无效"}
	ErrNotFound    = &errorx.PublicError{Kind: errorx.NotFound, Message: "文件不存在"}
)
