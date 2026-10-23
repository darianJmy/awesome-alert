package models

import "time"

type DatasourceGroup struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:128;not null;uniqueIndex" json:"name"`
	Description string    `gorm:"size:500" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type DatasourceGroupMember struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	GroupID      uint      `gorm:"not null;uniqueIndex:idx_datasource_group_member;index" json:"group_id"`
	DatasourceID uint      `gorm:"not null;uniqueIndex:idx_datasource_group_member;index" json:"datasource_id"`
	CreatedAt    time.Time `json:"created_at"`
}
