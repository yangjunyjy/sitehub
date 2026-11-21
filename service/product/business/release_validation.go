package services

import (
	"encoding/hex"
	"net/url"
	"path"
	"regexp"
	"sitehub/common/errorx"
	domain "sitehub/service/product/domain"
	"strings"
	"time"
	"unicode/utf8"
)

var releaseID = regexp.MustCompile("^[a-zA-Z0-9][a-zA-Z0-9._-]{0,119}$")

func validTarget(raw string) bool {
	u, err := url.ParseRequestURI(raw)
	if err != nil || strings.ContainsAny(raw, "\\\r\n\t") || u.User != nil || u.Fragment != "" {
		return false
	}
	if u.Scheme != "" {
		return (u.Scheme == "http" || u.Scheme == "https") && u.Hostname() != ""
	}
	const prefix = "/v1/file/local/releases/"
	if !strings.HasPrefix(u.Path, prefix) || u.RawQuery != "" || u.RawPath != "" {
		return false
	}
	name := strings.TrimPrefix(u.Path, prefix)
	return name != "" && path.Base(name) == name && ValidDownloadName(name)
}
func validText(value string, limit int) bool {
	return utf8.ValidString(value) && utf8.RuneCountInString(value) <= limit && !strings.ContainsRune(value, 0)
}
func validateFile(f domain.ReleaseFile) error {
	if !releaseID.MatchString(f.ID) || (f.GroupID != "" && !releaseID.MatchString(f.GroupID)) {
		return errorx.Invalid("下载文件或分组标识无效")
	}
	if f.Product == "" || f.Platform == "" || f.Version == "" || !validText(f.Product, 120) || !validText(f.Platform, 80) || !validText(f.Version, 120) {
		return errorx.Invalid("请填写有效的产品名称、版本号和平台")
	}
	if !validText(f.Edition, 80) || !validText(f.Arch, 120) || !validText(f.Format, 80) || !validText(f.Icon, 255) || !validText(f.Description, 5000) || f.SortOrder < 0 || f.SortOrder > 99999 || f.Size < 0 {
		return errorx.Invalid("下载文件元数据无效")
	}
	if !ValidDownloadName(f.FileName) || (f.URL != "" && (!validText(f.URL, 4096) || !validTarget(f.URL))) {
		return errorx.Invalid("请填写合法的文件名、HTTP(S) 下载地址或本站上传地址")
	}
	if f.Checksum != "" {
		b, err := hex.DecodeString(f.Checksum)
		if err != nil || len(b) != 32 {
			return errorx.Invalid("SHA256 校验和需为 64 位十六进制")
		}
	}
	if f.ReleaseDate != "" {
		if _, err := time.Parse("2006-01-02", f.ReleaseDate); err != nil {
			return errorx.Invalid("发布日期格式应为 YYYY-MM-DD")
		}
	}
	return nil
}
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
