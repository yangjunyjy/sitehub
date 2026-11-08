package services

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	errorx "sitehub/common/errorx"
	filecontract "sitehub/service/file/contract"
	productcontract "sitehub/service/product/contract"
)

func (s *websiteContentService) GetPublishedDownload(id string) (productcontract.ProductDownload, error) {
	if strings.TrimSpace(id) == "" || id != strings.TrimSpace(id) || len(id) > 200 {
		return productcontract.ProductDownload{}, errorx.Invalid("安装包 ID 无效")
	}
	value, err := s.GetDownloads()
	if err != nil {
		return productcontract.ProductDownload{}, err
	}
	for _, item := range value.Packages {
		if item.ID == id && item.Status == "published" && validDownloadURL(item.URL) {
			return productcontract.ProductDownload{URL: item.URL, FileName: item.FileName}, nil
		}
	}
	return productcontract.ProductDownload{}, &errorx.PublicError{Kind: errorx.NotFound, Message: "安装包不存在或已下架"}
}

// 本地上传地址不应携带开发机的回环主机名；外部下载服务器的 URL 原样保留。
func ValidDownloadName(name string) bool {
	if name == "" {
		return true
	}
	if name != strings.TrimSpace(name) || name == "." || name == ".." || len(name) > 255 || strings.ContainsAny(name, `\/:*?"<>|`) {
		return false
	}
	for _, r := range name {
		if r < 32 || r == 127 {
			return false
		}
	}
	return true
}

func (s *websiteContentService) validateLocalReleases(previous, downloads productcontract.WebsiteDownloads) error {
	previousPackages := make(map[string]productcontract.WebsiteDownloadPackage, len(previous.Packages))
	for _, item := range previous.Packages {
		previousPackages[item.ID] = item
	}
	for _, item := range downloads.Packages {
		// 草稿允许准备地址；文件丢失也不能阻止管理员将已有版本下架。
		if item.Status != "published" {
			continue
		}
		// 迁移后可能仍保留未搬迁的旧文件。未变更的已发布地址不应阻止
		// 管理员删除/下架其他记录或发布新文件；重新发布仍必须检查文件。
		old, exists := previousPackages[item.ID]
		if exists && old.Status == "published" && old.URL == item.URL {
			continue
		}
		parsed, err := url.Parse(item.URL)
		if err != nil || parsed.Host != "" || !strings.HasPrefix(parsed.Path, "/v1/file/local/releases/") {
			continue
		}
		if err := s.files.ValidateRelease(item.URL); err != nil {
			var public *errorx.PublicError
			if errors.As(err, &public) && public.Kind == errorx.InvalidInput {
				return fmt.Errorf("%w: %s", ErrInvalidWebsiteDownloads, public.Message)
			}
			return err
		}

	}
	return nil
}

func normalizeReleaseURL(raw string) string { return filecontract.NormalizeReleaseURL(raw) }
