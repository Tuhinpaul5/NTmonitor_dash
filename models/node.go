package models

import "time"

type NodeStatus string

const (
	NodeStatusActive	NodeStatus = "active"
	NodeStatusInactive	NodeStatus = "inactive"
	NodeStatusIdle	NodeStatus = "idle"
)

type Node struct {
	ID        string    	`gorm:"primaryKey;type:char(32);default:substr(gen_random_uuid()::text, 1, 32)" json:"id"`
	Email     string    	`gorm:"not null" json:"email"`
	Otp       uint      	`gorm:"not null;uniqueIndex" json:"otp"`
	Status    NodeStatus	`gorm:"type:node_status_enum;not null;default:'active'" json:"type"`
	CreatedAt time.Time 	`json:"created_at"`
	UpdatedAt time.Time 	`json:"updated_at"`
}