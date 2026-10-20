package models

import "time"

type PromQueryHistory struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	UserID         uint      `gorm:"not null;index:idx_prom_history_user_created" json:"-"`
	DatasourceID   uint      `gorm:"not null;index" json:"datasource_id"`
	DatasourceName string    `gorm:"size:100;not null" json:"datasource_name"`
	Query          string    `gorm:"type:text;not null" json:"query"`
	Mode           string    `gorm:"size:20;not null" json:"mode"`
	CreatedAt      time.Time `gorm:"index:idx_prom_history_user_created" json:"created_at"`
}
