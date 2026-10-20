package models

import "time"

// GitCredential stores one reusable Git host authentication account.
type GitCredential struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"size:128;not null;uniqueIndex" json:"name"`
	Provider     string    `gorm:"size:32;not null;default:other" json:"provider"` // github, gitlab, gitee, other
	BaseURL      string    `gorm:"size:500;not null" json:"base_url"`
	AuthType     string    `gorm:"size:20;not null;default:none" json:"auth_type"` // none, token, ssh_key
	Username     string    `gorm:"size:128" json:"username"`
	SecretCipher string    `gorm:"type:longtext" json:"-"`
	Description  string    `gorm:"size:500" json:"description"`
	Status       int       `gorm:"not null;default:1;index" json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
