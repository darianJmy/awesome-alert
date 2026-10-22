package models

import "time"

type AIModel struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"size:128;not null;uniqueIndex" json:"name"`
	Provider     string    `gorm:"size:32;not null" json:"provider"`
	BaseURL      string    `gorm:"size:1000;not null" json:"base_url"`
	Model        string    `gorm:"size:200;not null" json:"model"`
	SecretCipher string    `gorm:"type:text;not null" json:"-"`
	Status       int       `gorm:"not null" json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
