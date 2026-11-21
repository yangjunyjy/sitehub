package entity

import "time"

type ArticleLike struct {
	MemberID  uint64 `gorm:"primaryKey;autoIncrement:false"`
	ArticleID uint64 `gorm:"primaryKey;autoIncrement:false;index"`
	CreatedAt time.Time
}

type ArticleFavorite struct {
	MemberID  uint64 `gorm:"primaryKey;autoIncrement:false;index:idx_reader_folder,priority:1"`
	ArticleID uint64 `gorm:"primaryKey;autoIncrement:false;index"`
	FolderID  uint64 `gorm:"not null;default:0;index:idx_reader_folder,priority:2"`
	CreatedAt time.Time
}

type FavoriteFolder struct {
	ID       uint64 `gorm:"primaryKey"`
	MemberID uint64 `gorm:"uniqueIndex:idx_member_folder_name;not null"`
	Name     string `gorm:"size:50;uniqueIndex:idx_member_folder_name;not null"`
}
