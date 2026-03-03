package models

import "time"

type OTP struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	User      User      `gorm:"foreignKey:UserID; constraint:OnUpdate:CASCADE, OnDelete:CASCADE" json:"-"`
	Otp       string    `gorm:"not null;uniqueIndex" json:"otp"`
	Type      string    `gorm:"not null" json:"type"` // e.g., "email_verification", "password_reset"
	Used      bool      `gorm:"not null" json:"used"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
