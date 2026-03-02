package models

import "time"

type User struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	Email     string     `gorm:"uniqueIndex;not null" json:"email"`
	Profile   UserData   `json:"userdata"` // One-to-One relationship
	CreatedAt time.Time  `json:"created_at"`
}

type UserData struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	UserID uint   `gorm:"uniqueIndex" json:"user_id"`
	Bio    string `json:"bio"`
}