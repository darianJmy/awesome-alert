package models

import "time"

type ChatMemory struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"not null;index:idx_user_scope_key,unique" json:"user_id"`
	WorkspaceID uint      `gorm:"default:0;index" json:"workspace_id"`
	Scope       string    `gorm:"size:50;not null;default:user;index:idx_user_scope_key,unique" json:"scope"`
	Key         string    `gorm:"size:100;not null;index:idx_user_scope_key,unique" json:"key"`
	Value       string    `gorm:"type:text;not null" json:"value"`
	Source      string    `gorm:"size:50;default:chat" json:"source"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
