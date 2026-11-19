package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	coreconfig "sitehub/common/config"
	"sitehub/common/runtimeconfig"
	blog "sitehub/service/blog"
	"sitehub/service/blog/api/internal/svc"
	blogservice "sitehub/service/blog/business"
	filecontract "sitehub/service/blog/contract"
	"sitehub/service/blog/domain"
	"sitehub/service/blog/jobs"
	blogrepository "sitehub/service/blog/model"
	"sitehub/service/blog/model/entity"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func mockDatabase(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	connection, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = connection.Close() })
	database, err := gorm.Open(mysql.New(mysql.Config{Conn: connection, SkipInitializeWithVersion: true}), &gorm.Config{DisableAutomaticPing: true, Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	return database, mock
}

type storageSettings struct{ value filecontract.WebsiteStorage }

func (s storageSettings) GetStorage() (filecontract.WebsiteStorage, error) { return s.value, nil }

type assetRepository struct {
	blogrepository.ArticleRepository
	assets map[string]*entity.ArticleAsset
}

func (r *assetRepository) AssetCreate(asset *entity.ArticleAsset) error {
	r.assets[asset.ID] = asset
	return nil
}
func (r *assetRepository) AssetUpdate(asset *entity.ArticleAsset) error {
	r.assets[asset.ID] = asset
	return nil
}
func (r *assetRepository) AssetGet(id string) (*entity.ArticleAsset, error) { return r.assets[id], nil }
func (r *assetRepository) AssetPublished(string) (bool, error)              { return false, nil }

func TestBlogAPISharesArticleAndFileStorage(t *testing.T) {
	blogDB, blogMock := mockDatabase(t)
	fileDB, fileMock := mockDatabase(t)
	root := filepath.Join(t.TempDir(), "files")
	storageConfig := coreconfig.UploadConfig{Type: "local"}
	storageConfig.Local.UploadPath = root
	private := blogservice.NewPrivateStorage(&storageConfig, storageSettings{filecontract.WebsiteStorage{Provider: "local", LocalPath: root}}, "stable-storage-secret-at-least-32-characters")
	repository := &assetRepository{assets: map[string]*entity.ArticleAsset{}}
	serviceContext := &svc.ServiceContext{DB: blogDB, StorageDB: fileDB, Core: &blog.Module{Assets: blogservice.NewArticleAssetService(repository, private), PrivateStorage: private}}
	requestCtx, cancel := contextpkgWithTimeout()
	defer cancel()
	blogMock.ExpectPing()
	fileMock.ExpectPing()
	if err := serviceContext.Health(requestCtx); err != nil {
		t.Fatal(err)
	}
	content := []byte("private asset written through the local blog capability")
	asset, err := serviceContext.Core.Assets.UploadArticleAsset(requestCtx, domain.ArticleActor{ID: 7, CanCreate: true}, "note.txt", int64(len(content)), bytes.NewReader(content))
	if err != nil {
		t.Fatal(err)
	}
	reader, err := serviceContext.Core.Assets.ReadArticleAsset(requestCtx, domain.ArticleActor{ID: 7}, asset.ID)
	if err != nil {
		t.Fatal(err)
	}
	read, err := io.ReadAll(reader.Body)
	_ = reader.Body.Close()
	if err != nil || !bytes.Equal(read, content) {
		t.Fatalf("local asset roundtrip failed: %q %v", read, err)
	}
	source, err := serviceContext.Core.PrivateStorage.ReadPrivate(requestCtx, repository.assets[asset.ID].Storage)
	if err != nil {
		t.Fatal(err)
	}
	streamed, err := io.ReadAll(source)
	source.Close()
	if err != nil || !bytes.Equal(streamed, content) {
		t.Fatalf("private storage disagrees with asset reader: %v", err)
	}
	if err := blogMock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
	if err := fileMock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func contextpkgWithTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

func TestUnifiedHealthDetectsFileDatabaseFailureAndCloseReleasesBothPools(t *testing.T) {
	blogDB, blogMock := mockDatabase(t)
	fileDB, fileMock := mockDatabase(t)
	service := &svc.ServiceContext{DB: blogDB, StorageDB: fileDB}
	blogMock.ExpectPing()
	fileMock.ExpectPing().WillReturnError(errors.New("file database offline"))
	if err := service.Health(context.Background()); err == nil || !strings.Contains(err.Error(), "file database") {
		t.Fatalf("file DB health was ignored: %v", err)
	}
	blogMock.ExpectClose()
	fileMock.ExpectClose()
	service.Close()
	service.Close()
	if err := blogMock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
	if err := fileMock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

type localJobPorts struct {
	started  chan struct{}
	finished chan struct{}
	once     sync.Once
}

func (p *localJobPorts) PublishScheduledArticles(ctx context.Context) (int64, error) {
	p.once.Do(func() { close(p.started) })
	<-ctx.Done()
	close(p.finished)
	return 0, ctx.Err()
}
func (p *localJobPorts) CleanupArticleAssets(context.Context) error { return nil }

func TestLocalSchedulerHonorsMaintenanceAndDrainsBeforeResourcesClose(t *testing.T) {
	traffic, err := runtimeconfig.New(runtimeconfig.Config{Maintenance: true, MaxConcurrentRequests: 1}, "test", "test")
	if err != nil {
		t.Fatal(err)
	}
	ports := &localJobPorts{started: make(chan struct{}), finished: make(chan struct{})}
	guarded := &localJobs{publisher: ports, cleaner: ports, traffic: traffic}
	if _, err := guarded.PublishScheduledArticles(context.Background()); err == nil {
		t.Fatal("maintenance did not block scheduled publishing")
	}
	if err := guarded.CleanupArticleAssets(context.Background()); err == nil {
		t.Fatal("maintenance did not block scheduled cleanup")
	}
	if err := traffic.Commit(runtimeconfig.Config{MaxConcurrentRequests: 1}, "running"); err != nil {
		t.Fatal(err)
	}
	scheduler, err := jobs.NewWithConfig(guarded, guarded, jobs.Config{Enabled: true, PublishSchedule: "@every 1s", CleanupSchedule: "@every 1h", JobTimeout: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	scheduler.Start()
	t.Cleanup(scheduler.Stop)
	select {
	case <-ports.started:
	case <-time.After(3 * time.Second):
		t.Fatal("local scheduled work never started")
	}
	// run defers scheduler.Stop before ServiceContext.Close; a returned Stop
	// must mean local database/file users are no longer running.
	scheduler.Stop()
	select {
	case <-ports.finished:
	default:
		t.Fatal("scheduler returned before its local work drained")
	}
	release, err := traffic.Enter()
	if err != nil {
		t.Fatal("scheduled work retained its concurrency permit")
	}
	release()
}
