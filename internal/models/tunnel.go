package models

import "time"

type Tunnel struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Type        string    `gorm:"size:20;not null;default:ssh" json:"type"` // ssh, agent
	Host        string    `gorm:"size:200;not null" json:"host"`
	Port        int       `gorm:"default:22" json:"port"`
	Username    string    `gorm:"size:50" json:"username"`
	AuthType    string    `gorm:"size:20;default:password" json:"auth_type"` // password, key
	AuthConfig  string    `gorm:"type:text" json:"auth_config"`              // JSON: {"password":"xxx"} or {"private_key":"xxx"}
	Region      string    `gorm:"size:50" json:"region"`                     // 地域标签: 华东/华南/美东...
	Status      int       `gorm:"default:1" json:"status"`
	Description string    `gorm:"size:500" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
