package services

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"

	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "golang.org/x/image/webp"

	blogdomain "sitehub/service/blog/domain"
	repository "sitehub/service/blog/model"
	blogentity "sitehub/service/blog/model/entity"
	filecontract "sitehub/service/file/contract"
)

type ArticleAssetService interface {
	UploadArticleAsset(context.Context, blogdomain.ArticleActor, string, int64, io.ReadSeeker) (*ArticleAssetView, error)
	ReadArticleAsset(context.Context, blogdomain.ArticleActor, string) (*ArticleAssetReader, error)
	CleanupArticleAssets(context.Context) error
}
type articleAssetService struct {
	articles repository.ArticleRepository
	storage  filecontract.PrivateStorage
}

func NewArticleAssetService(repo repository.ArticleRepository, storage filecontract.PrivateStorage) ArticleAssetService {
	return &articleAssetService{articles: repo, storage: storage}
}

type ArticleAssetView struct {
	ID   string `json:"id"`
	URL  string `json:"url"`
	Name string `json:"name"`
	MIME string `json:"mime"`
	Size int64  `json:"size"`
}
type ArticleAssetReader struct {
	Body       io.ReadCloser
	Name, MIME string
	Size       int64
}

func (s *articleAssetService) UploadArticleAsset(ctx context.Context, actor blogdomain.ArticleActor, name string, size int64, src io.ReadSeeker) (*ArticleAssetView, error) {
	if actor.ID == 0 || !(actor.CanCreate || actor.CanUpdate) {
		return nil, blogdomain.ErrArticleForbidden
	}
	if s.articles == nil {
		return nil, errors.New("asset repository unavailable")
	}
	ext := strings.ToLower(filepath.Ext(name))
	images := map[string]string{".png": "image/png", ".jpg": "image/jpeg", ".jpeg": "image/jpeg", ".gif": "image/gif", ".webp": "image/webp"}
	attachments := map[string]bool{".pdf": true, ".txt": true, ".md": true, ".zip": true, ".csv": true, ".json": true, ".yaml": true, ".yml": true}
	if images[ext] == "" && !attachments[ext] {
		return nil, blogdomain.ErrInvalidArticle
	}
	limit := int64(50 << 20)
	if images[ext] != "" {
		limit = 10 << 20
	}
	if size <= 0 || size > limit {
		return nil, blogdomain.ErrInvalidArticle
	}
	header := make([]byte, 512)
	n, err := src.Read(header)
	if err != nil && err != io.EOF {
		return nil, err
	}
	mime := http.DetectContentType(header[:n])
	if _, err = src.Seek(0, 0); err != nil {
		return nil, err
	}
	if expected := images[ext]; expected != "" {
		if mime != expected {
			return nil, blogdomain.ErrInvalidArticle
		}
		dimensions, _, e := image.DecodeConfig(src)
		if e != nil || dimensions.Width <= 0 || dimensions.Height <= 0 || int64(dimensions.Width)*int64(dimensions.Height) > 40000000 {
			return nil, blogdomain.ErrInvalidArticle
		}
		mime = expected
	} else {
		if ext == ".pdf" && !strings.HasPrefix(string(header[:n]), "%PDF-") {
			return nil, blogdomain.ErrInvalidArticle
		}
		if ext == ".zip" && !strings.HasPrefix(string(header[:n]), "PK") {
			return nil, blogdomain.ErrInvalidArticle
		}
		mime = "application/octet-stream"
	}
	if _, err = src.Seek(0, 0); err != nil {
		return nil, err
	}
	hasher := sha256.New()
	if _, err = io.Copy(hasher, src); err != nil {
		return nil, err
	}
	if _, err = src.Seek(0, 0); err != nil {
		return nil, err
	}
	id := uuid.NewString()
	key := "article-assets/" + id + ext
	sealed, err := s.storage.PreparePrivate(key)
	if err != nil {
		return nil, err
	}
	a := &blogentity.ArticleAsset{ID: id, OwnerID: actor.ID, Name: filepath.Base(name), MIME: mime, Size: size, SHA256: fmt.Sprintf("%x", hasher.Sum(nil)), Storage: sealed, CreatedAt: time.Now()}
	// The pending record is durable before I/O, so crashes and failed uploads remain discoverable by cleanup.
	if err = s.articles.AssetCreate(a); err != nil {
		return nil, err
	}
	err = s.storage.WritePrivate(ctx, sealed, src, mime)
	if err != nil {
		return nil, err
	}
	a.Location = key
	if err = s.articles.AssetUpdate(a); err != nil {
		return nil, err
	}
	return &ArticleAssetView{ID: id, URL: "/v1/blog/assets?id=" + id, Name: a.Name, MIME: a.MIME, Size: a.Size}, nil
}
func (s *articleAssetService) ReadArticleAsset(ctx context.Context, actor blogdomain.ArticleActor, id string) (*ArticleAssetReader, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, blogdomain.ErrInvalidArticle
	}
	if s.articles == nil {
		return nil, blogdomain.ErrArticleNotFound
	}
	asset, err := s.articles.AssetGet(id)
	if err != nil {
		return nil, articleError(err)
	}
	if asset.DeletedAt != nil || asset.Location == "" {
		return nil, blogdomain.ErrArticleNotFound
	}
	public, err := s.articles.AssetPublished(id)
	if err != nil {
		return nil, err
	}
	if !public && (actor.ID == 0 || !(asset.OwnerID == actor.ID || actor.CanRead || actor.CanUpdate)) {
		return nil, blogdomain.ErrArticleNotFound
	}
	body, err := s.storage.ReadPrivate(ctx, asset.Storage)
	if err != nil {
		return nil, err
	}
	return &ArticleAssetReader{Body: body, Name: asset.Name, MIME: asset.MIME, Size: asset.Size}, nil
}
func (s *articleAssetService) CleanupArticleAssets(ctx context.Context) error {
	if s.articles == nil {
		return nil
	}
	// A cancelled scheduler run must also cancel SQL reads and lock acquisition.
	s = &articleAssetService{articles: s.articles.WithContext(ctx), storage: s.storage}
	assets, err := s.articles.AssetCleanupCandidates(time.Now().Add(-7 * 24 * time.Hour))
	if err != nil {
		return err
	}
	var failures []error
	for _, asset := range assets {
		if ctx.Err() != nil {
			return errors.Join(append(failures, ctx.Err())...)
		}
		if err = s.articles.WorkflowTransaction(func(repo repository.ArticleRepository) error { return repo.AssetRetire(asset.ID) }); errors.Is(err, blogdomain.ErrArticleState) {
			continue
		} else if err != nil {
			failures = append(failures, fmt.Errorf("retire asset %s: %w", asset.ID, err))
			continue
		}
		// Retirement is committed first; the persisted row is the retry queue until physical removal succeeds.
		removalCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
		err = s.storage.DeletePrivate(removalCtx, asset.Storage)
		cancel()
		if err == nil {
			err = s.articles.AssetForget(asset.ID)
		}
		if err != nil {
			failures = append(failures, fmt.Errorf("remove asset %s: %w", asset.ID, err))
		}
	}
	return errors.Join(failures...)
}
