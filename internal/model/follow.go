package model

import "time"

type Follow struct {
	UserID    uint      `gorm:"primaryKey;autoIncrement:false" json:"user_id"`
	FollowID  uint      `gorm:"primaryKey;autoIncrement:false" json:"follow_id"`
	CreatedAt time.Time `json:"created_at"`
}
