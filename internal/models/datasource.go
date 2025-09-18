package models

import (
	"awesome-alert/internal/types"
	"time"
)

type DataSource struct {
	Name             string           `json:"name"`
	Type             string           `json:"type"`
	Description      string           `json:"description"`
	PrometheusConfig PrometheusConfig `json:"prometheus_config"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

type PrometheusConfig struct {
	URL            string            `json:"url"`
	Headers        map[string]string `json:"headers"`
	Params         map[string]string `json:"params"`
	Authentication struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	} `json:"authentication"`
	SkipTLS bool `json:"skip_tls"`
}

type DataSourceList struct {
	List       []*DataSource     `json:"list"`
	Pagination *types.Pagination `json:"pagination"`
}

type RequestDataSourceList struct {
	Name     string `query:"name"`
	Page     int    `query:"page" `
	PageSize int    `query:"page_size"`
}
