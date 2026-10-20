package models

import "time"

type NotificationChannel struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	Name             string    `gorm:"size:128;not null;index" json:"name"`
	Type             string    `gorm:"size:30;not null;default:webhook;index" json:"type"`
	TemplateID       uint      `gorm:"index" json:"template_id"`
	WebhookURLCipher string    `gorm:"type:longtext" json:"-"`
	SignSecretCipher string    `gorm:"type:longtext" json:"-"`
	Status           int       `gorm:"not null;default:1;index" json:"status"`
	Description      string    `gorm:"size:500" json:"description"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type NotificationTemplate struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"size:128;not null;index" json:"name"`
	Type         string    `gorm:"size:40;not null;default:feishu;index" json:"type"`
	Description  string    `gorm:"size:500" json:"description"`
	Content      string    `gorm:"type:longtext" json:"content"`
	AdvancedCard bool      `gorm:"not null;default:false" json:"advanced_card"`
	Status       int       `gorm:"not null;default:1;index" json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type NotificationRecord struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ChannelID   uint      `gorm:"index" json:"channel_id"`
	ChannelName string    `gorm:"size:128" json:"channel_name"`
	EventType   string    `gorm:"size:80;not null;index" json:"event_type"`
	EventKey    string    `gorm:"size:255;index" json:"event_key"`
	Title       string    `gorm:"size:255" json:"title"`
	Status      string    `gorm:"size:20;not null;index" json:"status"`
	RequestBody string    `gorm:"type:text" json:"request_body"`
	Response    string    `gorm:"type:text" json:"response"`
	Error       string    `gorm:"type:text" json:"error"`
	CreatedAt   time.Time `json:"created_at"`
}
