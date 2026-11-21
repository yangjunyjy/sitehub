package domain

import "time"

type ArticleInteraction struct {
	Likes     int64  `json:"likes"`
	Favorites int64  `json:"favorites"`
	Liked     bool   `json:"liked"`
	Favorited bool   `json:"favorited"`
	FolderID  uint64 `json:"folderId,string"`
}

type ReaderFolder struct {
	ID    uint64 `json:"id,string"`
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

type SavedArticle struct {
	ID          uint64    `json:"id,string"`
	Title       string    `json:"title"`
	PublishedAt time.Time `json:"publishedAt"`
	SavedAt     time.Time `json:"savedAt"`
}

type SavedArticlePage struct {
	Items []SavedArticle `json:"items"`
	Total int64          `json:"total"`
}
