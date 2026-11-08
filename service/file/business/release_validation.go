package services

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	errorx "sitehub/common/errorx"
)

func (s *StorageSettings) ValidateRelease(location string) error {
	parsed, err := url.Parse(location)
	if err != nil {
		return errorx.Invalid("安装包路径无效")
	}
	if parsed.Host != "" || !strings.HasPrefix(parsed.Path, "/v1/file/local/releases/") {
		return nil
	}
	storage, err := s.GetStorage()
	if err != nil {
		return err
	}
	// 仅验证本站实际文件，不对任意外部 URL 发起网络请求。
	root := storage.LocalPath
	if root == "" {
		root = "./data/files"
	}
	root, err = filepath.Abs(root)
	if err != nil {
		return err
	}
	key := strings.TrimPrefix(parsed.Path, "/v1/file/local/")
	if strings.ContainsAny(key, `\:`) {
		return errorx.Invalid("安装包路径无效，请重新上传文件")
	}
	for _, part := range strings.Split(key, "/") {
		if part == ".." || part == "." {
			return errorx.Invalid("安装包路径无效，请重新上传文件")
		}
	}
	target := filepath.Join(root, filepath.FromSlash(key))
	relative, err := filepath.Rel(root, target)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)) || filepath.IsAbs(relative) {
		return errorx.Invalid("安装包路径无效，请重新上传文件")
	}
	info, err := os.Stat(target)
	if errors.Is(err, os.ErrNotExist) {
		return errorx.Invalid("发布地址对应文件不存在，请重新上传文件；暂不发布可保存为草稿")
	}
	if err != nil {
		return fmt.Errorf("检查本地安装包: %w", err)
	}
	if !info.Mode().IsRegular() {
		return errorx.Invalid("安装包地址必须指向文件")
	}
	return nil
}
