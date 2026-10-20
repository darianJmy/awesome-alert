package models

import "time"

type Dashboard struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Provider    string    `gorm:"size:20;not null;default:grafana;index" json:"provider"` // grafana, aliyun_sls
	URL         string    `gorm:"size:1000;not null" json:"url"`                          // iframe URL
	Description string    `gorm:"size:500" json:"description"`
	Sort        int       `gorm:"default:0" json:"sort"`
	Status      int       `gorm:"default:1" json:"status"` // 1=启用 0=禁用
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
