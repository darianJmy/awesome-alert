package models

import "time"

type ChatConversation struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"not null;index:idx_chat_user_updated,priority:1" json:"user_id"`
	Title       string    `gorm:"size:200" json:"title"`
	ModelID     uint      `gorm:"not null;default:0" json:"model_id"`
	WorkspaceID uint      `gorm:"default:0" json:"workspace_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `gorm:"index:idx_chat_user_updated,priority:2" json:"updated_at"`
}

type ChatMessage struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	ConversationID uint      `gorm:"index;not null" json:"conversation_id"`
	Role           string    `gorm:"size:20;not null" json:"role"`
	Content        string    `gorm:"type:text" json:"content"`
	Metadata       string    `gorm:"type:text" json:"metadata"`
	CreatedAt      time.Time `json:"created_at"`
}
