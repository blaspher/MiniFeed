package model

import "time"

type Post struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	ImageURL  string    `gorm:"type:varchar(255)" json:"image_url"`
	LikeCount int       `gorm:"not null;default:0" json:"like_count"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
}
