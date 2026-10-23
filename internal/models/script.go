package models

import "time"

// Script 脚本
type Script struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Name          string    `gorm:"size:100;not null;uniqueIndex" json:"name"`
	Type          string    `gorm:"size:20;not null;default:shell" json:"type"` // shell, python
	Content       string    `gorm:"type:text" json:"content"`
	SourceType    string    `gorm:"size:20;not null;default:inline;index" json:"source_type"` // inline, git
	RepositoryURL string    `gorm:"size:500" json:"repository_url"`
	GitRef        string    `gorm:"size:255" json:"git_ref"`
	Entrypoint    string    `gorm:"size:500" json:"entrypoint"`
	Description   string    `gorm:"size:500" json:"description"`
	Status        int       `gorm:"default:1" json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ScriptExecution 脚本执行记录
type ScriptExecution struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	ScriptID   uint      `gorm:"index;not null" json:"script_id"`
	Status     string    `gorm:"size:20;not null" json:"status"` // success, fail, error
	Output     string    `gorm:"type:text" json:"output"`
	DurationMs int64     `json:"duration_ms"`
	CreatedAt  time.Time `json:"created_at"`
}

// Workflow 工作流
type Workflow struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Description string    `gorm:"size:500" json:"description"`
	Steps       string    `gorm:"type:text" json:"steps"`
	Status      int       `gorm:"default:1" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// WorkflowExecution 工作流执行记录
type WorkflowExecution struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	WorkflowID uint      `gorm:"index;not null" json:"workflow_id"`
	Status     string    `gorm:"size:20;not null" json:"status"` // success, fail, error
	Result     string    `gorm:"type:text" json:"result"`        // JSON array of step results
	DurationMs int64     `json:"duration_ms"`
	CreatedAt  time.Time `json:"created_at"`
}
