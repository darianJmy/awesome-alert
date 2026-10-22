package handlers

import (
	"context"
	"encoding/json"
	"github.com/cloudwego/eino/schema"
	aicore "guanlan/internal/ai"
)

// Reuse the local MCP business handlers; no loopback HTTP or service token is needed.
func registerAssistantMCPTools(ctx context.Context, registry *aicore.Registry, userID uint) {
	for _, tool := range mcpTools {
		if tool.Annotations["readOnlyHint"] != true || authorizeAIPermission(ctx, userID, tool.Permission) != nil {
			continue
		}
		params := map[string]*schema.ParameterInfo{}
		properties, _ := tool.InputSchema["properties"].(map[string]interface{})
		for name, raw := range properties {
			prop := raw.(map[string]interface{})
			kind := schema.String
			if prop["type"] == "integer" {
				kind = schema.Integer
			}
			values, _ := prop["enum"].([]string)
			desc, _ := prop["description"].(string)
			params[name] = &schema.ParameterInfo{Type: kind, Enum: values, Desc: desc}

		}
		mustRegisterTool(registry, aicore.ToolSpec{Name: tool.Name, Description: tool.Description + "（平台级查询，不按数据源分组筛选）", Risk: aicore.RiskReadonlyLow, TimeoutSeconds: 10, ParamsSchema: params}, assistantMCPTool)
	}
}

func assistantMCPTool(ctx context.Context, call aicore.ToolCall) aicore.ToolResult {
	var data interface{}
	var err error
	switch call.Tool {
	case "query_alerts":
		data, err = mcpQueryAlertsContext(ctx, call.Params)
	case "query_alert_rules":
		data, err = mcpQueryAlertRulesContext(ctx, call.Params)
	case "query_probe_tasks":
		data, err = mcpQueryProbeTasksContext(ctx, call.Params)
	case "query_automation_tasks":
		data, err = mcpQueryAutomationTasksContext(ctx, call.Params)
	default:
		return aicore.ToolResult{Tool: call.Tool, ErrorCode: "TOOL_NOT_ALLOWED", Message: "工具不可用"}
	}
	if err != nil {
		return aicore.ToolResult{Tool: call.Tool, ErrorCode: "QUERY_FAILED", Message: "平台查询失败"}
	}
	// Explicit projection keeps commands, configuration and raw payloads out of model context.
	raw, err := json.Marshal(data)
	if err != nil {
		return aicore.ToolResult{Tool: call.Tool, ErrorCode: "RESULT_INVALID"}
	}
	var result map[string]interface{}
	if json.Unmarshal(raw, &result) != nil {
		return aicore.ToolResult{Tool: call.Tool, ErrorCode: "RESULT_INVALID"}
	}
	fields := map[string][]string{
		"query_alerts":           {"id", "title", "status", "severity", "source", "fault_center_id", "first_event_at", "last_event_at"},
		"query_alert_rules":      {"id", "name", "status", "severity", "group_id", "datasource_id"},
		"query_probe_tasks":      {"id", "name", "type", "status", "last_status", "last_run_at", "alert_active"},
		"query_automation_tasks": {"id", "name", "status", "source_type", "node_type", "node_id", "timeout_seconds"},
	}
	items := []map[string]interface{}{}
	if rows, ok := result["list"].([]interface{}); ok {
		for _, raw := range rows {
			row, ok := raw.(map[string]interface{})
			if !ok {
				continue
			}
			item := map[string]interface{}{}
			for _, key := range fields[call.Tool] {
				if value, ok := row[key]; ok {
					item[key] = value
				}
			}
			items = append(items, item)
		}
	}
	result["list"] = items
	return aicore.ToolResult{Tool: call.Tool, OK: true, Summary: "平台查询成功（平台级范围）", Data: result}
}
