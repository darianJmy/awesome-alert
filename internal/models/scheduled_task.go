package models

import "time"

type ScheduledTask struct {
	ID                  uint       `gorm:"primaryKey" json:"id"`
	Name                string     `gorm:"size:100;not null" json:"name"`
	Type                string     `gorm:"size:20;not null;default:script" json:"type"` // script, workflow
	TargetID            uint       `gorm:"index" json:"target_id"`
	CronExpr            string     `gorm:"size:50;not null" json:"cron_expr"` // */5 * * * *
	Description         string     `gorm:"size:500" json:"description"`
	ExternalLabels      string     `gorm:"type:text" json:"external_labels"`
	Severity            string     `gorm:"size:20;not null;default:P2" json:"severity"`
	FailureThreshold    int        `gorm:"not null;default:1" json:"failure_threshold"`
	ConsecutiveFailures int        `gorm:"not null;default:0" json:"consecutive_failures"`
	FaultCenterID       uint       `gorm:"not null;default:0;index" json:"fault_center_id"`
	AlertActive         bool       `gorm:"not null;default:false" json:"alert_active"`
	Status              int        `gorm:"default:1" json:"status"` // 1=启用 0=禁用
	LastRunAt           *time.Time `json:"last_run_at"`
	LastStatus          string     `gorm:"size:20" json:"last_status"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type ScheduledTaskExecution struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	TaskID     uint      `gorm:"index;not null" json:"task_id"`
	Status     string    `gorm:"size:20;not null" json:"status"` // success, fail, error
	Output     string    `gorm:"type:text" json:"output"`
	DurationMs int64     `json:"duration_ms"`
	CreatedAt  time.Time `json:"created_at"`
}
