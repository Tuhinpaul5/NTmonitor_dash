package models

import "time"

type UserSession struct {
	ID           string    `gorm:"primaryKey;type:char(32);default:substr(gen_random_uuid()::text, 1, 32)" json:"id"`
	UserID       string    `gorm:"type:char(32);not null;index" json:"user_id"`
	SessionToken string    `gorm:"uniqueIndex;not null" json:"session_token"`
	ExpiresAt    time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}
