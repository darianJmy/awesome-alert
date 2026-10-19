package models

import "time"

// APIAccessToken stores only a one-way hash of the credential. The plaintext
// token is returned once when it is created and can never be recovered.
type APIAccessToken struct {
	ID          string     `gorm:"primaryKey;size:32" json:"id"`
	Name        string     `gorm:"size:100;not null;uniqueIndex" json:"name"`
	TokenHash   string     `gorm:"size:64;not null;uniqueIndex" json:"-"`
	TokenPrefix string     `gorm:"size:32;not null" json:"token_prefix"`
	UserID      uint       `gorm:"not null;index" json:"user_id"`
	User        User       `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"user"`
	Description string     `gorm:"size:500" json:"description"`
	Status      int        `gorm:"not null;default:1;index" json:"status"`
	ExpiresAt   time.Time  `gorm:"not null;index" json:"expires_at"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	LastUsedIP  string     `gorm:"size:64" json:"last_used_ip"`
	CreatedBy   uint       `gorm:"not null;index" json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
