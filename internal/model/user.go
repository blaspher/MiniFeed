package model

import "time"

type User struct {
	ID        uint      `gorm:"primarykey;AUTO_INCREMENT" json:"id"`
	Username  string    `gorm:"size:32;uniqueIndex;not null" json:"username"`
	Password  string    `gorm:"size:128;not null" json:"-"`
	CreatedAt time.Time `json:"created_at"`
}
