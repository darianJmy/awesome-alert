package ai

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
	"unicode/utf8"
)

// StepTrace deliberately excludes arguments, results and provider error messages.
type StepTrace struct {
	Round      int    `json:"round"`
	Kind       string `json:"kind"`
	Tool       string `json:"tool,omitempty"`
	Status     string `json:"status"`
	DurationMS int64  `json:"duration_ms"`
	ErrorCode  string `json:"error_code,omitempty"`
}

type RunTrace struct {
	ID         string      `json:"run_id"`
	Status     string      `json:"status"`
	DurationMS int64       `json:"duration_ms"`
	Steps      []StepTrace `json:"steps"`
}

func (o *Orchestrator) Run(parent context.Context, req RunRequest) (response RunResponse) {
	started := time.Now()
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		return RunResponse{Answered: true, Answer: "无法创建分析任务，请重试。"}
	}
	trace := RunTrace{ID: hex.EncodeToString(id), Status: "failed", Steps: []StepTrace{}}
	response.Metadata = map[string]interface{}{"orchestrator": true}
	defer func() {
		trace.DurationMS = time.Since(started).Milliseconds()
		response.Metadata["execution"] = trace
	}()
	fail := func(code, answer string) RunResponse {
		response.Answered, response.Answer = true, answer
		response.Metadata["error_code"] = code
		return response
	}
	budget := o.Timeout
	if budget <= 0 {
		budget = 120 * time.Second
	}
	ctx, cancel := context.WithTimeout(parent, budget)
	defer cancel()
	stopped := func() RunResponse {
		trace.Status = "cancelled"
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			trace.Status = "timeout"
			return fail("TIMEOUT", "本次分析已超时，请缩小查询范围后重试。")
		}
		return fail("CANCELLED", "本次分析已停止。")
	}
	if ctx.Err() != nil {
		return stopped()
	}
	if o.Registry == nil || o.LLM == nil {
		return fail("NOT_CONFIGURED", "AI 分析服务未配置完整。")
	}
	maxRounds := o.MaxRounds
	if maxRounds <= 0 {
		maxRounds = 8
	}
	conversation := []Message{{Role: "user", Content: o.initialPrompt(req)}}
	for round := 1; round <= maxRounds; round++ {
		if ctx.Err() != nil {
			return stopped()
		}
		start := time.Now()
		raw, err := o.LLM(ctx, o.systemPrompt(), conversation)
		step := StepTrace{Round: round, Kind: "model", Status: "succeeded", DurationMS: time.Since(start).Milliseconds()}
		if err != nil || ctx.Err() != nil {
			step.Status, step.ErrorCode = "failed", "MODEL_ERROR"
			trace.Steps = append(trace.Steps, step)
			if ctx.Err() != nil {
				return stopped()
			}
			return fail("MODEL_ERROR", "AI 服务暂时无法完成分析，请检查模型配置或稍后重试。")
		}
		parsed, err := parseModelResponse(raw)
		if err != nil || (parsed.Type != "final" && parsed.Type != "tool_call") || (parsed.Type == "final" && strings.TrimSpace(parsed.Answer) == "") {
			step.Status, step.ErrorCode = "failed", "INVALID_MODEL_RESPONSE"
			trace.Steps = append(trace.Steps, step)
			return fail("INVALID_MODEL_RESPONSE", "AI 返回格式不正确，本次分析已停止，请重试。")
		}
		trace.Steps = append(trace.Steps, step)
		if parsed.Type == "final" {
			trace.Status = "succeeded"
			response.Answered, response.Answer = true, strings.TrimSpace(parsed.Answer)
			return response
		}
		start = time.Now()
		result := o.execute(ctx, ToolCall{Tool: parsed.Tool, Params: parsed.Params})
		status := "succeeded"
		if !result.OK {
			status = "failed"
		}
		// Only registered names are exposed; an arbitrary model string may contain secrets.
		toolName := parsed.Tool
		if _, ok := o.Registry.Get(toolName); !ok {
			toolName = "unknown"
		}
		trace.Steps = append(trace.Steps, StepTrace{Round: round, Kind: "tool", Tool: toolName, Status: status, DurationMS: time.Since(start).Milliseconds(), ErrorCode: result.ErrorCode})
		if ctx.Err() != nil {
			return stopped()
		}
		conversation = append(conversation, Message{Role: "assistant", Content: raw}, Message{Role: "user", Content: "工具返回:\n" + boundedResult(result, o.MaxResultBytes)})
	}
	trace.Status = "limited"
	return fail("ROUND_LIMIT", "已达到本次分析的调用轮次上限，请缩小问题范围后重试。")
}

func (o *Orchestrator) execute(ctx context.Context, call ToolCall) ToolResult {
	denied := func(code, message string) ToolResult {
		return ToolResult{Tool: call.Tool, ErrorCode: code, Message: message}
	}
	tool, ok := o.Registry.Get(call.Tool)
	if !ok {
		return denied("UNKNOWN_TOOL", "工具未注册")
	}
	if tool.Spec.Risk != RiskReadonlyLow && tool.Spec.Risk != RiskReadonlyExpensive {
		return denied("READ_ONLY_REQUIRED", "当前 AI 执行链路仅允许只读操作")
	}
	if err := validateParams(tool.Spec, call.Params); err != nil {
		return denied("INVALID_PARAMS", err.Error())
	}
	if o.Authorize == nil {
		return denied("FORBIDDEN", "未配置工具授权策略")
	}
	if err := o.Authorize(ctx, tool.Spec, call); err != nil {
		return denied("FORBIDDEN", "没有执行该工具的权限")
	}
	seconds := tool.Spec.TimeoutSeconds
	if seconds <= 0 {
		seconds = 15
	}
	toolCtx, cancel := context.WithTimeout(ctx, time.Duration(seconds)*time.Second)
	defer cancel()
	if toolCtx.Err() != nil {
		return denied("CANCELLED", "调用已停止")
	}
	result := tool.Handler(toolCtx, call)
	if toolCtx.Err() != nil {
		return denied("TOOL_TIMEOUT", "工具调用已超时或取消")
	}
	return result
}

func validateParams(spec ToolSpec, params map[string]interface{}) error {
	for name := range params {
		if _, ok := spec.ParamsSchema[name]; !ok {
			return fmt.Errorf("不支持的参数")
		}
	}
	for name, raw := range spec.ParamsSchema {
		schema, ok := raw.(string)
		if !ok {
			return fmt.Errorf("工具参数定义无效")
		}
		value, exists := params[name]
		if !exists || value == nil {
			if strings.Contains(schema, "optional") {
				continue
			}
			return fmt.Errorf("缺少参数 %s", name)
		}
		if strings.HasPrefix(schema, "number") {
			n, ok := value.(float64)
			if !ok || math.IsNaN(n) || math.IsInf(n, 0) || math.Trunc(n) != n || n < 0 || n > float64(1<<53-1) || (name == "datasource_id" && n == 0) {
				return fmt.Errorf("参数 %s 必须是有效整数", name)
			}
		} else {
			text, ok := value.(string)
			if !ok || len(text) > 8192 {
				return fmt.Errorf("参数 %s 必须是有效字符串", name)
			}
			kind := strings.Fields(schema)[0]
			if strings.Contains(kind, "|") {
				found := false
				for _, option := range strings.Split(kind, "|") {
					if text == option {
						found = true
					}
				}
				if !found {
					return fmt.Errorf("参数 %s 不在允许范围", name)
				}
			}
		}
	}
	return nil
}

func boundedResult(result ToolResult, limit int) string {
	text := MarshalJSONCompact(result)
	if limit <= 0 {
		limit = 64 * 1024
	}
	if len(text) <= limit {
		return text
	}
	// Keep a valid JSON envelope. Do not send a partial JSON document to the model.
	result.Data = nil
	result.Summary = "查询结果过大，请缩小范围后重新查询。"
	result.Message = ""
	result.Meta = map[string]interface{}{"truncated_for_llm": true, "original_bytes": len(text)}
	text = MarshalJSONCompact(result)
	if len(text) > limit {
		text = `{"ok":false,"error_code":"RESULT_TOO_LARGE"}`
	}
	if !utf8.ValidString(text) {
		return `{"ok":false,"error_code":"INVALID_ENCODING"}`
	}
	return text
}
