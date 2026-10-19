package ai

import "encoding/json"

type RiskLevel string

const (
	RiskReadonlyLow       RiskLevel = "readonly_low"
	RiskReadonlyExpensive RiskLevel = "readonly_expensive"
	RiskWriteConfirm      RiskLevel = "write_requires_confirmation"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ToolSpec struct {
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	DatasourceType string                 `json:"datasource_type,omitempty"`
	Risk           RiskLevel              `json:"risk"`
	ParamsSchema   map[string]interface{} `json:"params_schema"`
	MaxTimeRange   string                 `json:"max_time_range,omitempty"`
	TimeoutSeconds int                    `json:"timeout_seconds,omitempty"`
}

type ToolCall struct {
	Tool   string                 `json:"tool"`
	Params map[string]interface{} `json:"params"`
}

type ToolResult struct {
	Tool      string                 `json:"tool"`
	OK        bool                   `json:"ok"`
	Summary   string                 `json:"summary,omitempty"`
	Data      interface{}            `json:"data,omitempty"`
	ErrorCode string                 `json:"error_code,omitempty"`
	Message   string                 `json:"message,omitempty"`
	Retryable bool                   `json:"retryable,omitempty"`
	Meta      map[string]interface{} `json:"meta,omitempty"`
}

type ModelResponse struct {
	Type   string                 `json:"type"`
	Answer string                 `json:"answer,omitempty"`
	Tool   string                 `json:"tool,omitempty"`
	Params map[string]interface{} `json:"params,omitempty"`
	Reason string                 `json:"reason,omitempty"`
}

func MarshalJSONCompact(value interface{}) string {
	body, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(body)
}
