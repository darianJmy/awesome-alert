package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/cloudwego/eino/schema"
	aicore "guanlan/internal/ai"
)

func mcpParameterSchema(object map[string]interface{}) map[string]*schema.ParameterInfo {
	fields := map[string]*schema.ParameterInfo{}
	required, _ := object["required"].([]string)
	props, _ := object["properties"].(map[string]interface{})
	for name, raw := range props {
		prop, _ := raw.(map[string]interface{})
		field := mcpParameterInfo(prop)
		for _, key := range required {
			if key == name {
				field.Required = true
			}
		}
		fields[name] = field
	}
	return fields
}

func mcpParameterInfo(prop map[string]interface{}) *schema.ParameterInfo {
	kind, _ := prop["type"].(string)
	desc, _ := prop["description"].(string)
	values, _ := prop["enum"].([]string)
	field := &schema.ParameterInfo{Type: schema.DataType(kind), Desc: desc, Enum: values}
	if kind == "array" {
		item, _ := prop["items"].(map[string]interface{})
		field.ElemInfo = mcpParameterInfo(item)
	}
	if kind == "object" {
		field.SubParams = mcpParameterSchema(prop)
	}
	return field
}

// Normalize both direct Go calls and JSON-RPC calls to the same JSON value types.
func validatePlatformMCPArgs(tool mcpTool, args map[string]interface{}) (map[string]interface{}, error) {
	if args == nil {
		args = map[string]interface{}{}
	}
	raw, err := json.Marshal(args)
	if err != nil || len(raw) > 64*1024 {
		return nil, fmt.Errorf("工具参数格式不正确或过长")
	}
	var normalized map[string]interface{}
	if json.Unmarshal(raw, &normalized) != nil {
		return nil, fmt.Errorf("工具参数格式不正确")
	}
	if err := aicore.ValidateToolParams(aicore.ToolSpec{ParamsSchema: mcpParameterSchema(tool.InputSchema)}, normalized); err != nil {
		return nil, err
	}
	if err := validateMCPBounds(tool.InputSchema, normalized); err != nil {
		return nil, err
	}
	return normalized, nil
}

func validateMCPBounds(def map[string]interface{}, value interface{}) error {
	if n, ok := value.(float64); ok {
		for _, bound := range []string{"minimum", "maximum"} {
			var limit float64
			switch v := def[bound].(type) {
			case int:
				limit = float64(v)
			case float64:
				limit = v
			default:
				continue
			}
			if (bound == "minimum" && n < limit) || (bound == "maximum" && n > limit) {
				return fmt.Errorf("数值参数超出允许范围")
			}
		}
	}
	if text, ok := value.(string); ok {
		if max, ok := def["maxLength"].(int); ok && len([]rune(text)) > max {
			return fmt.Errorf("文本参数过长")
		}
	}
	if values, ok := value.([]interface{}); ok {
		if min, ok := def["minItems"].(int); ok && len(values) < min {
			return fmt.Errorf("列表至少需要 %d 项", min)
		}
		if max, ok := def["maxItems"].(int); ok && len(values) > max {
			return fmt.Errorf("列表最多允许 %d 项", max)
		}
		item, _ := def["items"].(map[string]interface{})
		for _, v := range values {
			if err := validateMCPBounds(item, v); err != nil {
				return err
			}
		}
	}
	if values, ok := value.(map[string]interface{}); ok {
		props, _ := def["properties"].(map[string]interface{})
		for key, v := range values {
			prop, _ := props[key].(map[string]interface{})
			if err := validateMCPBounds(prop, v); err != nil {
				return fmt.Errorf("%s: %w", key, err)
			}
		}
	}
	return nil
}
