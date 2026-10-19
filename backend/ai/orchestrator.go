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

func (o *Orchestrator) Run(ctx context.Context, req RunRequest) RunResponse {
	if o.Registry == nil || o.LLM == nil {
		return RunResponse{Answered: false, Metadata: map[string]interface{}{"orchestrator_error": "missing registry or llm"}}
	}
	maxRounds := o.MaxRounds
	if maxRounds <= 0 {
		maxRounds = 8
	}

	systemPrompt := o.systemPrompt()
	conversation := []Message{{Role: "user", Content: o.initialPrompt(req)}}
	toolCalls := make([]map[string]interface{}, 0)

	for round := 1; round <= maxRounds; round++ {
		raw, err := o.LLM(ctx, systemPrompt, conversation)
		if err != nil {
			return RunResponse{Answered: false, Metadata: map[string]interface{}{"orchestrator_error": err.Error(), "round": round}}
		}

		parsed, err := parseModelResponse(raw)
		if err != nil {
			return RunResponse{Answered: false, Metadata: map[string]interface{}{"orchestrator_error": err.Error(), "raw": raw, "round": round}}
		}
		if parsed.Type == "final" {
			return RunResponse{Answered: true, Answer: strings.TrimSpace(parsed.Answer), Metadata: map[string]interface{}{"orchestrator": true, "rounds": round, "tool_calls": toolCalls}}
		}
		if parsed.Type != "tool_call" {
			return RunResponse{Answered: false, Metadata: map[string]interface{}{"orchestrator_error": "unsupported response type", "raw": raw, "round": round}}
		}

		tool, ok := o.Registry.Get(parsed.Tool)
		if !ok {
			result := ToolResult{Tool: parsed.Tool, OK: false, ErrorCode: "UNKNOWN_TOOL", Message: "工具未注册", Retryable: true}
			conversation = append(conversation, Message{Role: "assistant", Content: raw}, Message{Role: "user", Content: "工具返回:\n" + MarshalJSONCompact(result)})
			continue
		}
		start := time.Now()
		result := tool.Handler(ctx, ToolCall{Tool: parsed.Tool, Params: parsed.Params})
		if result.Meta == nil {
			result.Meta = map[string]interface{}{}
		}
		result.Meta["duration_ms"] = time.Since(start).Milliseconds()
		resultText := MarshalJSONCompact(result)
		if o.MaxResultBytes > 0 && len(resultText) > o.MaxResultBytes {
			resultText = resultText[:o.MaxResultBytes]
			result.Meta["truncated_for_llm"] = true
		}
		toolCalls = append(toolCalls, map[string]interface{}{
			"round":  round,
			"tool":   parsed.Tool,
			"params": parsed.Params,
			"ok":     result.OK,
		})
		conversation = append(conversation, Message{Role: "assistant", Content: raw}, Message{Role: "user", Content: "工具返回:\n" + resultText})
	}

	return RunResponse{Answered: true, Answer: "已达到本次智能分析最大工具调用轮次，基于已有信息无法继续自动探查。请缩小问题范围或指定数据源/namespace 后重试。", Metadata: map[string]interface{}{"orchestrator": true, "max_rounds_reached": true, "tool_calls": toolCalls}}
}

func (o *Orchestrator) systemPrompt() string {
	return `你是运维中枢平台的智能运维调度核心。
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
