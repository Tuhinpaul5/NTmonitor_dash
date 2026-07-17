package models

import "time"

type NodeStatus string

const (
	NodeStatusActive	NodeStatus = "active"
	NodeStatusInactive	NodeStatus = "inactive"
	NodeStatusIdle		NodeStatus = "idle"
)

type Node struct {
	ID        string    `gorm:"primaryKey;type:char(32);default:substr(gen_random_uuid()::text, 1, 32)" json:"id"`
	NodeName   string   `json:"node_name"`
	GateWayId  string    `json:"gateway_id"`
	Status    NodeStatus `gorm:"type:node_status_enum;not null;default:'active'" json:"status"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}