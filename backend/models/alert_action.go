package models

import "time"

// AlertActionRule connects an alert lifecycle transition to an automated action.
type AlertActionRule struct {
	ID                    uint                `gorm:"primaryKey" json:"id"`
	FaultCenterID         uint                `gorm:"not null;index" json:"fault_center_id"`
	Name                  string              `gorm:"size:128;not null" json:"name"`
	MatchAll              bool                `gorm:"not null;default:false" json:"match_all"`
	MatchersJSON          string              `gorm:"column:matchers;type:text" json:"-"`
	Matchers              []AlertLabelMatcher `gorm:"-" json:"matchers"`
	TriggerStatus         string              `gorm:"size:20;not null;index" json:"trigger_status"` // firing, resolved
	ActionType            string              `gorm:"size:20;not null;index" json:"action_type"`    // script, workflow, webhook
	TargetID              uint                `gorm:"index" json:"target_id"`
	NodeType              string              `gorm:"size:20" json:"node_type"` // direct, ssh, agent
	NodeID                uint                `gorm:"index" json:"node_id"`
	WebhookURLCipher      string              `gorm:"type:text" json:"-"`
	WebhookSecretCipher   string              `gorm:"type:text" json:"-"`
	SkipSuppressed        bool                `gorm:"not null;default:true" json:"skip_suppressed"`
	CooldownSeconds       int                 `gorm:"not null;default:0" json:"cooldown_seconds"`
	MaxRetries            int                 `gorm:"not null;default:0" json:"max_retries"`
	TimeoutSeconds        int                 `gorm:"not null;default:30" json:"timeout_seconds"`
	NotifyOnSuccess       bool                `gorm:"not null;default:false" json:"notify_on_success"`
	NotifyOnFailure       bool                `gorm:"not null;default:true" json:"notify_on_failure"`
	NotificationChannelID uint                `gorm:"index" json:"notification_channel_id"`
	Status                int                 `gorm:"not null;default:1;index" json:"status"`
	CreatedAt             time.Time           `json:"created_at"`
	UpdatedAt             time.Time           `json:"updated_at"`
}

// AlertActionExecution is the immutable incident/action binding plus its latest execution state.
type AlertActionExecution struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	FaultCenterID uint       `gorm:"not null;index" json:"fault_center_id"`
	RuleID        uint       `gorm:"not null;uniqueIndex:idx_alert_action_once;index" json:"rule_id"`
	RuleName      string     `gorm:"size:128;not null" json:"rule_name"`
	AlertID       uint       `gorm:"not null;uniqueIndex:idx_alert_action_once;index" json:"alert_id"`
	AlertTitle    string     `gorm:"size:255;not null" json:"alert_title"`
	TriggerStatus string     `gorm:"size:20;not null;uniqueIndex:idx_alert_action_once;index" json:"trigger_status"`
	ActionType    string     `gorm:"size:20;not null" json:"action_type"`
	ScopeKey      string     `gorm:"size:64;index" json:"scope_key"`
	Status        string     `gorm:"size:32;not null;index" json:"status"`
	Attempt       int        `gorm:"not null;default:0" json:"attempt"`
	MaxAttempts   int        `gorm:"not null;default:1" json:"max_attempts"`
	NextRunAt     *time.Time `gorm:"index" json:"next_run_at"`
	LeaseUntil    *time.Time `gorm:"index" json:"-"`
	StartedAt     *time.Time `json:"started_at"`
	FinishedAt    *time.Time `json:"finished_at"`
	Output        string     `gorm:"type:text" json:"output"`
	Error         string     `gorm:"type:text" json:"error"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
