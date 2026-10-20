package models

import "time"

type AgentConnection struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	Name           string     `gorm:"size:128;not null;uniqueIndex" json:"name"`
	TokenHash      string     `gorm:"size:64;not null;uniqueIndex" json:"-"`
	Status         string     `gorm:"size:20;not null;default:offline;index" json:"status"`
	Hostname       string     `gorm:"size:255" json:"hostname"`
	Region         string     `gorm:"size:100;index" json:"region"`
	Version        string     `gorm:"size:64" json:"version"`
	RemoteAddr     string     `gorm:"size:128" json:"remote_addr"`
	Capabilities   string     `gorm:"type:text" json:"capabilities"`
	LastSeenAt     *time.Time `json:"last_seen_at"`
	ConnectedAt    *time.Time `json:"connected_at"`
	DisconnectedAt *time.Time `json:"disconnected_at"`
	Description    string     `gorm:"size:500" json:"description"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
