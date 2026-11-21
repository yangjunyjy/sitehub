package main

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"gorm.io/gorm"

	"sitehub/common/runtimeconf"
	"sitehub/service/blog/domain"
	"sitehub/service/blog/model/entity"
	blogmysql "sitehub/service/blog/model/mysql"
)

func TestV5ReaderMigrationAndBookmarkOwnership(t *testing.T) {
	path := filepath.Join(t.TempDir(), "blog.db")
	t.Setenv("BLOG_DB_DSN", path)
	if err := run("blog", "sqlite", false); err != nil {
		t.Fatal(err)
	}
	db, err := (runtimeconf.Database{Driver: "sqlite", DSN: path, MaxOpen: 1}).Open()
	if err != nil {
		t.Fatal(err)
	}
	pool, _ := db.DB()
	t.Cleanup(func() { pool.Close() })
	for _, statement := range []string{
		"DROP TABLE article_favorites", "DROP TABLE favorite_folders", "DROP TABLE article_likes",
		"ALTER TABLE friend_links DROP COLUMN member_id", "ALTER TABLE guest_messages DROP COLUMN member_id",
		"INSERT INTO guest_messages (nickname,content,status) VALUES ('visitor','legacy','approved')",
		"UPDATE microservice_schema SET version = 5",
	} {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 2; i++ {
		if err := run("blog", "sqlite", false); err != nil {
			t.Fatal(err)
		}
	}
	var message entity.GuestMessage
	if err := db.Where("content = ?", "legacy").Take(&message).Error; err != nil || message.MemberID != 0 || message.Status != "approved" {
		t.Fatalf("legacy message: %+v %v", message, err)
	}
	article := entity.Article{Title: "edited draft", Slug: "reader-test", Content: "body", Status: 1, Likes: 15, Favorites: 2}
	if err := db.Create(&article).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&entity.ArticlePublication{ArticleID: article.ID, Title: "Published", PublishedAt: time.Now(), Payload: "{}"}).Error; err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	repo := blogmysql.NewReaderRepository(db)
	const reader, other = uint64(17), uint64(18)
	var requests sync.WaitGroup
	for i := 0; i < 8; i++ {
		requests.Add(1)
		go func() {
			defer requests.Done()
			state, err := repo.SetLike(ctx, reader, article.ID, true)
			if err != nil || state.Likes != 16 || !state.Liked {
				t.Errorf("replayed like: %+v %v", state, err)
			}
		}()
	}
	requests.Wait()
	state, err := repo.SetLike(ctx, reader, article.ID, false)
	if err != nil || state.Likes != 15 || state.Liked {
		t.Fatalf("undo like: %+v %v", state, err)
	}
	folder, err := repo.CreateFolder(ctx, reader, "  Go  ")
	if err != nil || folder.Name != "Go" || folder.ID == 0 {
		t.Fatalf("create folder: %+v %v", folder, err)
	}
	if _, err := repo.CreateFolder(ctx, other, "Go"); err != nil {
		t.Fatal(err)
	}
	for _, id := range []uint64{0, folder.ID, folder.ID} {
		state, err := repo.SetFavorite(ctx, reader, article.ID, id, true)
		if err != nil || state.Favorites != 3 || !state.Favorited || state.FolderID != id {
			t.Fatalf("save: %+v %v", state, err)
		}
	}
	if _, err := repo.SetFavorite(ctx, other, article.ID, folder.ID, true); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("foreign folder: %v", err)
	}
	if _, err := repo.Articles(ctx, other, folder.ID, 1, 10); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("foreign articles: %v", err)
	}
	list, err := repo.Articles(ctx, reader, folder.ID, 1, 10)
	if err != nil || list.Total != 1 || len(list.Items) != 1 || list.Items[0].Title != "Published" {
		t.Fatalf("public snapshot: %+v %v", list, err)
	}
	if err := db.Model(&article).Update("status", 0).Error; err != nil {
		t.Fatal(err)
	}
	list, err = repo.Articles(ctx, reader, folder.ID, 1, 10)
	if err != nil || len(list.Items) != 0 || list.Total != 0 {
		t.Fatalf("withdrawn article: %+v %v", list, err)
	}
	folders, err := repo.Folders(ctx, reader)
	if err != nil || len(folders) != 2 || folders[1].Count != 0 {
		t.Fatalf("visible counts: %+v %v", folders, err)
	}
	if _, err := repo.SetLike(ctx, reader, article.ID, true); !errors.Is(err, domain.ErrArticleNotFound) {
		t.Fatalf("withdrawn article like: %v", err)
	}
}
