package business

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"sitehub/common/config"
	filecontract "sitehub/service/blog/contract"
)

type assetStorage struct {
	Config    config.UploadConfig
	ObjectKey string
}

func NewPrivateStorage(cfg *config.UploadConfig, settings filecontract.SettingsReader, key string) filecontract.PrivateStorage {
	s := NewUploadService(cfg, settings).(*uploadService)
	hash := sha256.Sum256([]byte("article-asset-storage-v1:" + key))
	s.assetKey = hash[:]
	return s
}
func (s *uploadService) PreparePrivate(key string) (string, error) {
	runtime, err := s.runtime()
	if err != nil {
		return "", err
	}
	cfg := runtime.config
	if cfg.Type == "local" {
		cfg.Local.UploadPath, err = filepath.Abs(cfg.Local.UploadPath + "-private")
		if err != nil {
			return "", err
		}
	}
	return s.sealStorage(assetStorage{Config: cfg, ObjectKey: key})
}
func (s *uploadService) WritePrivate(ctx context.Context, snapshot string, body io.Reader, mime string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stored, err := s.openStorage(snapshot)
	if err != nil {
		return err
	}
	if stored.Config.Type == "local" {
		_, err = uploadToLocal(stored.Config, stored.ObjectKey, body)
		return err
	}
	client, err := newS3Client(stored.Config)
	if err != nil {
		return err
	}
	_, err = s3manager.NewUploaderWithClient(client).UploadWithContext(ctx, &s3manager.UploadInput{Bucket: aws.String(stored.Config.Remote.Bucket), Key: aws.String(stored.ObjectKey), Body: body, ContentType: aws.String(mime), ACL: aws.String("private")})
	return err
}
func (s *uploadService) ReadPrivate(ctx context.Context, snapshot string) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	stored, err := s.openStorage(snapshot)
	if err != nil {
		return nil, err
	}
	if stored.Config.Type == "local" {
		target, err := assetLocalPath(stored)
		if err != nil {
			return nil, err
		}
		return os.Open(target)
	}
	client, err := newS3Client(stored.Config)
	if err != nil {
		return nil, err
	}
	result, err := client.GetObjectWithContext(ctx, &s3.GetObjectInput{Bucket: aws.String(stored.Config.Remote.Bucket), Key: aws.String(stored.ObjectKey)})
	if err != nil {
		return nil, err
	}
	return result.Body, nil
}
func (s *uploadService) sealStorage(v assetStorage) (string, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(s.assetKey)
	if err != nil {
		return "", err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(aead.Seal(nonce, nonce, raw, []byte("article-asset-v1"))), nil
}
func (s *uploadService) openStorage(value string) (assetStorage, error) {
	var v assetStorage
	raw, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return v, err
	}
	block, err := aes.NewCipher(s.assetKey)
	if err != nil {
		return v, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return v, err
	}
	if len(raw) < aead.NonceSize() {
		return v, errors.New("invalid asset storage")
	}
	plain, err := aead.Open(nil, raw[:aead.NonceSize()], raw[aead.NonceSize():], []byte("article-asset-v1"))
	if err != nil {
		return v, err
	}
	err = json.Unmarshal(plain, &v)
	return v, err
}

func assetLocalPath(stored assetStorage) (string, error) {
	root, err := filepath.Abs(stored.Config.Local.UploadPath)
	if err != nil {
		return "", err
	}
	target := filepath.Join(root, filepath.FromSlash(stored.ObjectKey))
	relative, err := filepath.Rel(root, target)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)) || filepath.IsAbs(relative) {
		return "", filecontract.ErrInvalidPath
	}
	return target, nil
}

func (s *uploadService) DeletePrivate(ctx context.Context, sealed string) error {
	stored, err := s.openStorage(sealed)
	if err != nil {
		return err
	}
	if stored.Config.Type == "local" {
		target, err := assetLocalPath(stored)
		if err != nil {
			return err
		}
		err = os.Remove(target)
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	client, err := newS3Client(stored.Config)
	if err != nil {
		return err
	}
	_, err = client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{Bucket: aws.String(stored.Config.Remote.Bucket), Key: aws.String(stored.ObjectKey)})
	return err
}
func (s *uploadService) ServeLocalPath(name string) (string, error) {
	runtime, err := s.runtime()
	if err != nil {
		return "", err
	}
	stored := assetStorage{Config: runtime.config, ObjectKey: strings.TrimPrefix(name, "/")}
	if stored.ObjectKey == "" || strings.Contains(stored.ObjectKey, "\\") {
		return "", filecontract.ErrInvalidPath
	}
	target, err := assetLocalPath(stored)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(target)
	if err != nil || !info.Mode().IsRegular() {
		return "", filecontract.ErrNotFound
	}
	return target, nil
}
