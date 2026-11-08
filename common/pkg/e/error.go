package e

import "errors"

var (
	ERROR_PARAMS     error = errors.New("参数错误")
	ERROR_OVER_LIMIT error = errors.New("请求频繁")
)
