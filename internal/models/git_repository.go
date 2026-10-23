package models

import "time"

// GitRepository is a centrally fetched source repository. Execution details live on AutomationTask.
type GitRepository struct {
	ID                uint          `gorm:"primaryKey" json:"id"`
	Name              string        `gorm:"size:128;not null;uniqueIndex" json:"name"`
	RepositoryURL     string        `gorm:"size:500;not null" json:"repository_url"`
	DefaultRef        string        `gorm:"size:255" json:"default_ref"`
	CredentialID      uint          `gorm:"index" json:"credential_id"`
	Credential        GitCredential `gorm:"foreignKey:CredentialID" json:"-"`
	Description       string        `gorm:"size:500" json:"description"`
	Status            int           `gorm:"not null;default:1;index" json:"status"`
	SyncedRevision    string        `gorm:"size:64;index" json:"synced_revision"`
	SyncedSHA256      string        `gorm:"size:64" json:"synced_sha256"`
	SyncedSize        int64         `json:"synced_size"`
	SyncedAt          *time.Time    `json:"synced_at"`
	LastSyncAttemptAt *time.Time    `json:"last_sync_attempt_at"`
	LastSyncError     string        `gorm:"type:text" json:"last_sync_error"`
	CreatedAt         time.Time     `json:"created_at"`
	UpdatedAt         time.Time     `json:"updated_at"`
}
