package logic

import "sitehub/common/rpcx"

func classify(err error) error { return rpcx.Error(err) }
