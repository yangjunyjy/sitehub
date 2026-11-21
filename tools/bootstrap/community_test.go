package main

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"sitehub/common/runtimeconf"
	"sitehub/service/blog/domain"
	"sitehub/service/blog/model/entity"
	blogmysql "sitehub/service/blog/model/mysql"
)

func TestV3GuestbookAndModeratedComments(t *testing.T) {
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
	defer pool.Close()
	for _, stmt := range []string{
		"DROP INDEX idx_message_thread", "ALTER TABLE guest_messages DROP COLUMN article_id", "UPDATE microservice_schema SET version = 3",
		"INSERT INTO guest_messages (content,status) VALUES ('old message','approved')",
	} {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := run("blog", "sqlite", false); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	repo := blogmysql.NewCommunityRepository(db)
	article := entity.Article{Title: "Article", Slug: "article", Content: "text", Status: 1}
	if err := db.Create(&article).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&entity.ArticlePublication{ArticleID: article.ID, Title: "Article", Slug: "article", Content: "text", Payload: "{}"}).Error; err != nil {
		t.Fatal(err)
	}
	comment := entity.GuestMessage{ArticleID: article.ID, Content: "article only", Status: "pending"}
	if err := repo.Submit(ctx, &comment); err != nil {
		t.Fatal(err)
	}
	queue, total, err := repo.ReviewQueue(ctx, "comments", "pending", 1)
	if err != nil || total != 1 || len(queue) != 1 || queue[0].ArticleID != article.ID {
		t.Fatalf("queue: %+v %v", queue, err)
	}
	if err := repo.Review(ctx, "messages", comment.ID, "approved"); !errors.Is(err, domain.ErrCommunityItemNotFound) {
		t.Fatalf("wrong scope: %v", err)
	}
	for _, state := range []string{"approved", "approved", "rejected"} {
		if err := repo.Review(ctx, "comments", comment.ID, state); err != nil {
			t.Fatal(err)
		}
		expected := int64(1)
		if state == "rejected" {
			expected = 0
		}
		items, total, err := repo.PublishedMessages(ctx, article.ID, 0, 10)
		if err != nil || total != expected || int64(len(items)) != expected {
			t.Fatalf("article comments: %+v %v", items, err)
		}
		var count int64
		if err := db.Model(&article).Select("comment_count").Scan(&count).Error; err != nil || count != expected {
			t.Fatalf("article count: %d %v", count, err)
		}
	}
	items, total, err := repo.PublishedMessages(ctx, 0, 0, 10)
	if err != nil || total != 1 || len(items) != 1 || items[0].Content != "old message" {
		t.Fatalf("guestbook: %+v %v", items, err)
	}
	friend := entity.FriendLink{Name: "My blog", URL: "https://example.org", Description: "Notes"}
	if err := repo.Apply(ctx, &friend); err != nil {
		t.Fatal(err)
	}
	if err := repo.Review(ctx, "friends", friend.ID, "approved"); err != nil {
		t.Fatal(err)
	}
	friends, err := repo.PublishedFriends(ctx)
	if err != nil || len(friends) != 1 || friends[0].ID != friend.ID {
		t.Fatalf("public friends: %+v %v", friends, err)
	}
}
