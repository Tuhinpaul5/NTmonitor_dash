package models

import "time"

// UserStatus enum
type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusPending   UserStatus = "pending"
)

// UserType enum
type UserType string

const (
	UserTypeAdmin     UserType = "admin"
	UserTypeModerator UserType = "moderator"
	UserTypeUser      UserType = "user"
	UserTypeGuest     UserType = "guest"
)

type User struct {
	ID         string     `gorm:"primaryKey;type:char(32);default:substr(gen_random_uuid()::text, 1, 32)" json:"id"`
	Username   string     `gorm:"uniqueIndex;not null" json:"username"`
	Password   string     `gorm:"not null" json:"-"`
	Email      string     `gorm:"uniqueIndex;not null" json:"email"`
	Profile    UserData   `json:"userdata"`                                                       // One-to-One relationship
	Status     UserStatus `gorm:"type:user_status_enum;not null;default:'pending'" json:"status"` // Account status
	Type       UserType   `gorm:"type:user_type_enum;not null;default:'user'" json:"type"`        // Role: user/admin
	IsVerified bool       `gorm:"not null;default:false" json:"is_verified"`                      // Email verification status
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type UserData struct {
	ID        string    `gorm:"primaryKey;type:char(32);default:substr(gen_random_uuid()::text, 1, 32)" json:"id"`
	UserID    string    `gorm:"uniqueIndex;type:char(32)" json:"user_id"`
	Bio       string    `json:"bio"`
	Phone     string    `json:"phone"`
	Ip        string    `json:"ip"`
	Country   string    `json:"country"`
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
