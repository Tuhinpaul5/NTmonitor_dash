package models

import "time"

type OtpType string

const (
	OtpTypeRegister 		OtpType = "register"
	OtpTypeLogin    		OtpType = "login"
	OtpTypePasswordReset    OtpType = "password"
)

type OTP struct {
	ID        string    `gorm:"primaryKey;type:char(32);default:substr(gen_random_uuid()::text, 1, 32)" json:"id"`
	Email     string    `gorm:"not null" json:"email"`
	Otp       uint      `gorm:"not null;uniqueIndex" json:"otp"`
	Type      OtpType   `gorm:"type:otp_type_enum;not null;default:'register'" json:"type"`
	Used      bool      `gorm:"not null" json:"used"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
