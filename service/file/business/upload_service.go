package services

import (
	"fmt"
	"io"
	"mime"

	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/google/uuid"

	"sitehub/common/config"
	errorx "sitehub/common/errorx"
	"sitehub/common/pkg/logger"
	filecontract "sitehub/service/file/contract"
)

const (
	maxImageSize      int64 = 20 * 1024 * 1024
	maxAttachmentSize int64 = 50 * 1024 * 1024
	maxReleaseSize    int64 = 2 * 1024 * 1024 * 1024
)

type UploadService interface {
	ServeLocalPath(string) (string, error)

	UploadImage(name string, size int64, src io.Reader) (string, error)
	DeleteImage(filename string) error
	UploadAttachment(name string, size int64, src io.Reader) (*AttachmentInfo, error)
	UploadRelease(name string, size int64, src io.Reader) (*AttachmentInfo, error)
}

type AttachmentInfo struct {
	URL      string `json:"url"`
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	MIMEType string `json:"mimeType"`
}

type uploadRuntime struct {
	config config.UploadConfig
	client *s3.S3
}

type uploadService struct {
	assetKey []byte

	baseConfig config.UploadConfig
	website    filecontract.SettingsReader
}

func NewUploadService(cfg *config.UploadConfig, website filecontract.SettingsReader) UploadService {
	base := config.UploadConfig{}
	if cfg != nil {
		base = *cfg
		base.Local.AllowedExts = append([]string(nil), cfg.Local.AllowedExts...)
	}
	return &uploadService{baseConfig: base, website: website}
}

func (s *uploadService) UploadImage(name string, size int64, src io.Reader) (string, error) {
	runtime, err := s.runtime()
	if err != nil {
		return "", err
	}
	ext := strings.ToLower(filepath.Ext(name))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true, ".ico": true}
	if !allowed[ext] || !isAllowedByStorage(runtime.config, ext) {
		return "", fmt.Errorf("不支持的图片类型: %s", ext)
	}
	if err := validateFileSize(size, effectiveLimit(runtime.config.Local.MaxSize, maxImageSize), "图片"); err != nil {
		return "", err
	}

	location, err := s.upload(runtime, "images", uuid.NewString()+ext, src, contentType(ext))
	if err != nil {
		return "", err
	}
	logger.Info("图片上传成功", logger.String("filename", name), logger.Int64("size", size), logger.String("url", location))
	return location, nil
}

func (s *uploadService) UploadAttachment(name string, size int64, src io.Reader) (*AttachmentInfo, error) {
	runtime, err := s.runtime()
	if err != nil {
		return nil, err
	}
	ext := strings.ToLower(filepath.Ext(name))
	allowed := map[string]bool{
		".pdf": true, ".xls": true, ".xlsx": true,
		".ppt": true, ".pptx": true, ".zip": true, ".rar": true, ".7z": true,
		".tar": true, ".gz": true, ".txt": true, ".md": true, ".csv": true,
		".mp3": true, ".mp4": true, ".avi": true, ".mov": true, ".json": true,
		".xml": true, ".yaml": true, ".yml": true,
	}
	if !allowed[ext] || !isAllowedByStorage(runtime.config, ext) {
		return nil, fmt.Errorf("不支持的附件类型: %s", ext)
	}
	if err := validateFileSize(size, effectiveLimit(runtime.config.Local.MaxSize, maxAttachmentSize), "附件"); err != nil {
		return nil, err
	}

	location, err := s.upload(runtime, "attachments", uuid.NewString()+ext, src, contentType(ext))
	if err != nil {
		return nil, err
	}
	return &AttachmentInfo{URL: location, Name: name, Size: size, MIMEType: contentType(ext)}, nil
}

// UploadRelease streams release artifacts directly to disk or through the S3
// multipart uploader. Large installers are never buffered entirely in memory.
func (s *uploadService) UploadRelease(name string, size int64, src io.Reader) (*AttachmentInfo, error) {
	runtime, err := s.runtime()
	if err != nil {
		return nil, err
	}
	ext := strings.ToLower(filepath.Ext(name))
	allowed := map[string]bool{
		".zip": true, ".tar": true, ".gz": true, ".tgz": true, ".deb": true,
		".rpm": true, ".exe": true, ".dmg": true, ".pkg": true, ".msi": true,
	}
	if !allowed[ext] || !isAllowedByStorage(runtime.config, ext) {
		return nil, errorx.Invalid("不支持的发布包类型")
	}
	if err := validateFileSize(size, effectiveLimit(runtime.config.Local.MaxSize, maxReleaseSize), "发布包"); err != nil {
		return nil, err
	}

	location, err := s.upload(runtime, "releases", uuid.NewString()+ext, src, contentType(ext))
	if err != nil {
		return nil, err
	}
	return &AttachmentInfo{URL: location, Name: name, Size: size, MIMEType: contentType(ext)}, nil
}

func (s *uploadService) DeleteImage(location string) error {
	runtime, err := s.runtime()
	if err != nil {
		return err
	}
	objectKey, err := objectKeyFromLocation(runtime.config, location)
	if err != nil {
		return err
	}
	if runtime.config.Type == "local" {
		root, err := filepath.Abs(runtime.config.Local.UploadPath)
		if err != nil {
			return fmt.Errorf("解析上传目录失败: %w", err)
		}
		target, err := filepath.Abs(filepath.Join(root, filepath.FromSlash(objectKey)))
		if err != nil || (target != root && !strings.HasPrefix(target, root+string(os.PathSeparator))) {
			return fmt.Errorf("无效的文件路径")
		}
		if err := os.Remove(target); err != nil {
			return fmt.Errorf("删除文件失败: %w", err)
		}
		return nil
	}
	_, err = runtime.client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(runtime.config.Remote.Bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return fmt.Errorf("从对象存储删除文件失败: %w", err)
	}
	return nil
}

func (s *uploadService) runtime() (*uploadRuntime, error) {
	cfg := s.baseConfig
	cfg.Local.AllowedExts = append([]string(nil), s.baseConfig.Local.AllowedExts...)
	if s.website != nil {
		storage, err := s.website.GetStorage()
		if err != nil {
			return nil, fmt.Errorf("读取存储配置失败: %w", err)
		}
		applyWebsiteStorage(&cfg, storage)
	}

	if cfg.Local.UploadPath == "" {
		cfg.Local.UploadPath = "./data/files"
	}
	if cfg.Type == "local" {
		return &uploadRuntime{config: cfg}, nil
	}
	client, err := newS3Client(cfg)
	if err != nil {
		return nil, err
	}
	return &uploadRuntime{config: cfg, client: client}, nil
}

func applyWebsiteStorage(cfg *config.UploadConfig, storage filecontract.WebsiteStorage) {
	provider := strings.ToLower(strings.TrimSpace(storage.Provider))
	cfg.Type = "local"
	if provider != "local" {
		cfg.Type = "remote"
	}
	cfg.URLPrefix = strings.TrimSpace(storage.URLPrefix)
	cfg.Local.UploadPath = strings.TrimSpace(storage.LocalPath)
	cfg.Local.AllowedExts = append([]string(nil), storage.AllowedExts...)
	cfg.Local.MaxSize = storage.MaxSize
	cfg.Remote.Provider = provider
	cfg.Remote.Endpoint = strings.TrimSpace(storage.Endpoint)
	cfg.Remote.AccessKey = strings.TrimSpace(storage.AccessKey)
	cfg.Remote.SecretKey = storage.SecretKey
	cfg.Remote.Bucket = strings.TrimSpace(storage.Bucket)
	cfg.Remote.Region = strings.TrimSpace(storage.Region)
	cfg.Remote.CDN = strings.TrimSpace(storage.CDN)
	cfg.Remote.PathStyle = storage.PathStyle
}

func newS3Client(cfg config.UploadConfig) (*s3.S3, error) {
	if cfg.Remote.Provider != "oss" && cfg.Remote.Provider != "minio" {
		return nil, fmt.Errorf("存储提供商仅支持 oss 或 minio")
	}
	if cfg.Remote.Endpoint == "" || cfg.Remote.Bucket == "" || cfg.Remote.AccessKey == "" || cfg.Remote.SecretKey == "" {
		return nil, fmt.Errorf("对象存储必须配置 endpoint、bucket、accessKey 和 secretKey")
	}
	region := cfg.Remote.Region
	if region == "" {
		region = "us-east-1"
	}
	endpoint, err := url.Parse(cfg.Remote.Endpoint)
	if err != nil || endpoint.Scheme == "" || endpoint.Host == "" {
		return nil, fmt.Errorf("对象存储 endpoint 必须是完整的 HTTP(S) 地址")
	}
	awsConfig := &aws.Config{
		Region:           aws.String(region),
		Endpoint:         aws.String(strings.TrimRight(cfg.Remote.Endpoint, "/")),
		DisableSSL:       aws.Bool(endpoint.Scheme == "http"),
		Credentials:      credentials.NewStaticCredentials(cfg.Remote.AccessKey, cfg.Remote.SecretKey, ""),
		S3ForcePathStyle: aws.Bool(cfg.Remote.PathStyle),
	}
	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, fmt.Errorf("创建对象存储客户端失败: %w", err)
	}
	return s3.New(sess), nil
}

func (s *uploadService) upload(runtime *uploadRuntime, prefix, filename string, body io.Reader, mediaType string) (string, error) {
	objectKey := strings.Trim(prefix, "/") + "/" + filepath.Base(filename)
	if runtime.config.Type == "local" {
		return uploadToLocal(runtime.config, objectKey, body)
	}
	uploader := s3manager.NewUploaderWithClient(runtime.client)
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(runtime.config.Remote.Bucket),
		Key:         aws.String(objectKey),
		Body:        body,
		ContentType: aws.String(mediaType),
	})
	if err != nil {
		return "", fmt.Errorf("上传到对象存储失败: %w", err)
	}
	return objectURL(runtime.config, objectKey), nil
}

func uploadToLocal(cfg config.UploadConfig, objectKey string, body io.Reader) (location string, err error) {
	root := cfg.Local.UploadPath
	if root == "" {
		root = "./data/files"
	}
	target := filepath.Join(root, filepath.FromSlash(objectKey))
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return "", fmt.Errorf("创建上传目录失败: %w", err)
	}
	destination, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return "", fmt.Errorf("创建目标文件失败: %w", err)
	}
	completed := false
	defer func() {
		_ = destination.Close()
		if !completed {
			_ = os.Remove(target)
		}
	}()
	if _, err := io.Copy(destination, body); err != nil {
		return "", fmt.Errorf("保存文件失败: %w", err)
	}
	if err := destination.Close(); err != nil {
		return "", fmt.Errorf("保存文件失败: %w", err)
	}
	completed = true
	relative := "/v1/file/local/" + strings.TrimLeft(filepath.ToSlash(objectKey), "/")
	if cfg.URLPrefix == "" {
		return relative, nil
	}
	return normalizeReleaseURL(strings.TrimRight(cfg.URLPrefix, "/") + relative), nil
}

func objectURL(cfg config.UploadConfig, objectKey string) string {
	if cfg.Remote.CDN != "" {
		return strings.TrimRight(cfg.Remote.CDN, "/") + "/" + objectKey
	}
	endpoint, err := url.Parse(cfg.Remote.Endpoint)
	if err != nil {
		return strings.TrimRight(cfg.Remote.Endpoint, "/") + "/" + cfg.Remote.Bucket + "/" + objectKey
	}
	if cfg.Remote.PathStyle {
		endpoint.Path = strings.TrimRight(endpoint.Path, "/") + "/" + cfg.Remote.Bucket + "/" + objectKey
	} else {
		if !strings.HasPrefix(endpoint.Hostname(), cfg.Remote.Bucket+".") {
			host := cfg.Remote.Bucket + "." + endpoint.Hostname()
			if endpoint.Port() != "" {
				host += ":" + endpoint.Port()
			}
			endpoint.Host = host
		}
		endpoint.Path = strings.TrimRight(endpoint.Path, "/") + "/" + objectKey
	}
	return endpoint.String()
}

func objectKeyFromLocation(cfg config.UploadConfig, location string) (string, error) {
	value := strings.TrimSpace(location)
	if value == "" {
		return "", fmt.Errorf("文件地址不能为空")
	}
	if parsed, err := url.Parse(value); err == nil && parsed.Path != "" {
		value = parsed.Path
	}
	value = strings.TrimPrefix(value, "/")
	value = strings.TrimPrefix(value, "v1/file/local/")
	if cfg.Type != "local" && cfg.Remote.PathStyle {
		value = strings.TrimPrefix(value, strings.Trim(cfg.Remote.Bucket, "/")+"/")
	}
	value = filepath.ToSlash(filepath.Clean(value))
	if value == "." || value == "" || strings.HasPrefix(value, "../") || filepath.IsAbs(value) {
		return "", fmt.Errorf("无效的文件地址")
	}
	return value, nil
}

func isAllowedByStorage(cfg config.UploadConfig, ext string) bool {
	if len(cfg.Local.AllowedExts) == 0 {
		return true
	}
	for _, allowed := range cfg.Local.AllowedExts {
		if strings.EqualFold(strings.TrimSpace(allowed), ext) {
			return true
		}
	}
	return false
}

func validateFileSize(size, limit int64, label string) error {
	if size <= 0 {
		return errorx.Invalid(fmt.Sprintf("%s不能为空", label))
	}
	if limit > 0 && size > limit {
		return errorx.Invalid(fmt.Sprintf("%s大小超过限制: 最大 %.0f MB", label, float64(limit)/(1024*1024)))
	}
	return nil
}

func effectiveLimit(configured, hardLimit int64) int64 {
	if configured <= 0 || configured > hardLimit {
		return hardLimit
	}
	return configured
}

func contentType(ext string) string {
	if detected := mime.TypeByExtension(strings.ToLower(ext)); detected != "" {
		return detected
	}
	return "application/octet-stream"
}
