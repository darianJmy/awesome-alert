package models

import (
	"awesome-alert/internal/types"
	"time"
)

type AlertRule struct {
	Group                  string                 `json:"group"`
	Name                   string                 `json:"name"`
	AdditionalLabel        map[string]interface{} `json:"additional_label"`
	Type                   string                 `json:"type"`
	DataSourceId           []int                  `json:"data_source_id"`
	Expr                   string                 `json:"expr"`
	Condition              map[string]interface{} `json:"condition"`
	Duration               int                    `json:"duration"`
	AlertContinuousCount   int                    `json:"alert_continuous_count"`
	RecoverContinuousCount int                    `json:"recover_continuous_count"`
	SendRecoverEvent       bool                   `json:"send_recover_event"`
	FaultCenterId          []int                  `json:"fault_center_id"`
	Enabled                bool                   `json:"enabled"`
	CreatedAt              time.Time              `json:"created_at"`
	UpdatedAt              time.Time              `json:"updated_at"`
}

type AlertRuleList struct {
	List       []*AlertRule      `json:"list"`
	Pagination *types.Pagination `json:"pagination"`
}

type RequestAlertRuleList struct {
	Group    string `query:"group"`
	Name     string `query:"name"`
	Page     int    `query:"page" `
	PageSize int    `query:"page_size"`
}
