package entity

import "time"

type FriendLink struct {
	ID          uint64 `gorm:"primaryKey"`
	Name        string `gorm:"size:80;not null"`
	URL         string `gorm:"size:500;not null"`
	Description string `gorm:"size:240;not null"`
	Avatar      string `gorm:"size:500"`
	Status      string `gorm:"size:16;not null;default:pending;index"`
	CreatedAt   time.Time
}

type GuestMessage struct {
	ArticleID uint64 `gorm:"not null;default:0;index:idx_message_thread,priority:1"`
	ID        uint64 `gorm:"primaryKey"`
	Nickname  string `gorm:"size:40"`
	Content   string `gorm:"type:text;not null"`
	Status    string `gorm:"size:16;not null;default:pending;index"`
	CreatedAt time.Time
}

type SocialLink struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

type AuthorProfile struct {
	HomeAnimation *bool  `gorm:"column:home_animation"`
	HomeVerse     string `gorm:"size:120"`
	ID            uint64 `gorm:"primaryKey"`
	Name          string
	Title         string
	Bio           string `gorm:"type:text"`
	Content       string `gorm:"type:text"`
	Avatar        string `gorm:"size:500"`
	Location      string
	Email         string
	Skills        []string     `gorm:"serializer:json"`
	Socials       []SocialLink `gorm:"serializer:json"`
}
