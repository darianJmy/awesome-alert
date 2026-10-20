package models

import "time"

// MCPClient represents one externally managed AI client installation. Its
// credential remains an APIAccessToken; this record only carries identity,
// compatibility and connection metadata.
type MCPClient struct {
	ID              string         `gorm:"primaryKey;size:32" json:"id"`
	Name            string         `gorm:"size:100;not null;uniqueIndex" json:"name"`
	ClientType      string         `gorm:"size:32;not null;index" json:"client_type"`
	APITokenID      string         `gorm:"size:32;not null;uniqueIndex" json:"api_token_id"`
	APIToken        APIAccessToken `gorm:"foreignKey:APITokenID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"api_token"`
	ReportedName    string         `gorm:"size:100" json:"reported_name"`
	ClientVersion   string         `gorm:"size:64" json:"client_version"`
	ProtocolVersion string         `gorm:"size:64" json:"protocol_version"`
	Description     string         `gorm:"size:500" json:"description"`
	Status          int            `gorm:"not null;default:1;index" json:"status"`
	LastSeenAt      *time.Time     `gorm:"index" json:"last_seen_at"`
	LastIP          string         `gorm:"size:64" json:"last_ip"`
	LastUserAgent   string         `gorm:"size:500" json:"last_user_agent"`
	CreatedBy       uint           `gorm:"not null;index" json:"created_by"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}
