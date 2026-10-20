package models

import "time"

// AuditLog is append-only application evidence for authenticated operations.
// It intentionally stores summaries rather than raw secrets or response bodies.
type AuditLog struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	UserID          uint      `gorm:"index" json:"user_id"`
	Username        string    `gorm:"size:50;index" json:"username"`
	WorkspaceID     uint      `gorm:"default:0;index" json:"workspace_id"`
	Module          string    `gorm:"size:50;index" json:"module"`
	Action          string    `gorm:"size:32;index" json:"action"`
	ResourceType    string    `gorm:"size:64;index" json:"resource_type"`
	ResourceID      string    `gorm:"size:128;index" json:"resource_id"`
	ResourceName    string    `gorm:"size:255" json:"resource_name"`
	Method          string    `gorm:"size:10" json:"method"`
	Path            string    `gorm:"size:500" json:"path"`
	RequestSummary  string    `gorm:"type:text" json:"request_summary"`
	Result          string    `gorm:"size:16;index" json:"result"`
	StatusCode      int       `json:"status_code"`
	ResponseMessage string    `gorm:"size:500" json:"response_message"`
	DurationMS      int64     `json:"duration_ms"`
	IP              string    `gorm:"size:64;index" json:"ip"`
	UserAgent       string    `gorm:"size:500" json:"user_agent"`
	CreatedAt       time.Time `gorm:"index" json:"created_at"`
}
