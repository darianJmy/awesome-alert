package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type LLMFunc func(ctx context.Context, systemPrompt string, messages []Message) (string, error)

type Orchestrator struct {
	Registry       *Registry
	LLM            LLMFunc
	MaxRounds      int
	MaxResultBytes int
	Timeout        time.Duration
	Authorize      func(context.Context, ToolSpec, ToolCall) error
}

type RunRequest struct {
	Question          string
	DatasourceCatalog string
	History           string
	Memory            string
	PlatformKnowledge string
}

type RunResponse struct {
	Answered bool
	Answer   string
	Metadata map[string]interface{}
}

func (o *Orchestrator) systemPrompt() string {
	return `你是观澜运维平台的智能运维调度核心。
你必须使用中文。
你只能输出 JSON，不能输出 Markdown。
输出格式二选一：
1. 调用工具：
{"type":"tool_call","tool":"工具名","params":{...},"reason":"为什么调用"}
2. 最终回答：
{"type":"final","answer":"给用户的最终中文回答"}

规则：
- 如果问题是普通知识、概念解释、闲聊，直接输出 final。
- 如果问题需要查询平台真实数据，必须调用工具。
- 当前阶段仅支持只读查询；涉及创建、修改、删除或批量变更时说明暂不支持，不得声称已执行。
- 工具返回失败或没有证据时如实说明，不能编造查询结果。
- 你看不到也不能要求用户提供密钥、token、AK/SK、URL、header。
- 你只能使用工具清单中的工具名和参数。
- 单次查询窗口最大 24h；日志默认 limit 50，Pod 日志默认 tail 200。
- 工具失败时，根据错误调整参数或给出可执行解释。`
}

func (o *Orchestrator) initialPrompt(req RunRequest) string {
	var sb strings.Builder
	sb.WriteString("用户问题:\n")
	sb.WriteString(req.Question)
	sb.WriteString("\n\n可用工具 JSON:\n")
	sb.WriteString(MarshalJSONCompact(o.Registry.Specs()))
	sb.WriteString("\n\n可用数据源目录（只含 id/name/type/capability，不含密钥和真实连接细节）:\n")
	sb.WriteString(req.DatasourceCatalog)
	if strings.TrimSpace(req.PlatformKnowledge) != "" {
		sb.WriteString("\n\n平台知识:\n")
		sb.WriteString(req.PlatformKnowledge)
	}
	if strings.TrimSpace(req.History) != "" {
		sb.WriteString("\n\n最近对话:\n")
		sb.WriteString(req.History)
	}
	if strings.TrimSpace(req.Memory) != "" {
		sb.WriteString("\n\n长期记忆:\n")
		sb.WriteString(req.Memory)
	}
	return sb.String()
}

func parseModelResponse(raw string) (ModelResponse, error) {
	text := extractJSON(raw)
	var response ModelResponse
	if err := json.Unmarshal([]byte(text), &response); err != nil {
		return ModelResponse{}, fmt.Errorf("解析 LLM JSON 失败: %w", err)
	}
	response.Type = strings.TrimSpace(response.Type)
	return response, nil
}

func extractJSON(value string) string {
	text := strings.TrimSpace(value)
	if strings.HasPrefix(text, "```") {
		text = strings.TrimPrefix(text, "```json")
		text = strings.TrimPrefix(text, "```")
		text = strings.TrimSuffix(text, "```")
		return strings.TrimSpace(text)
	}
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start >= 0 && end > start {
		return text[start : end+1]
	}
	return text
}
