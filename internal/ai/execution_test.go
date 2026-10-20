package ai

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

func testOrchestrator(t *testing.T, handler ToolHandler) *Orchestrator {
	t.Helper()
	registry := NewRegistry()
	if err := registry.Register(ToolSpec{Name: "query", Risk: RiskReadonlyLow, TimeoutSeconds: 1, ParamsSchema: map[string]interface{}{"datasource_id": "number"}}, handler); err != nil {
		t.Fatal(err)
	}
	return &Orchestrator{Registry: registry, Authorize: func(context.Context, ToolSpec, ToolCall) error { return nil }}
}

func TestExecutionTraceSurvivesModelFailure(t *testing.T) {
	calls := 0
	o := testOrchestrator(t, func(context.Context, ToolCall) ToolResult {
		calls++
		return ToolResult{OK: true, Data: "private-result"}
	})
	round := 0
	o.LLM = func(context.Context, string, []Message) (string, error) {
		round++
		if round == 1 {
			return `{"type":"tool_call","tool":"query","params":{"datasource_id":1}}`, nil
		}
		return "", errors.New("secret-provider-token")
	}
	result := o.Run(context.Background(), RunRequest{Question: "check"})
	trace := result.Metadata["execution"].(RunTrace)
	if calls != 1 || trace.Status != "failed" || len(trace.Steps) != 3 || trace.Steps[1].Tool != "query" {
		t.Fatalf("lost steps: %#v", trace)
	}
	metadata := MarshalJSONCompact(result.Metadata)
	for _, secret := range []string{"private-result", "secret-provider-token", "datasource_id"} {
		if strings.Contains(metadata, secret) {
			t.Fatalf("trace exposed %s", secret)
		}
	}
	if !result.Answered || len(trace.ID) != 32 {
		t.Fatalf("invalid response: %#v", result)
	}
}

func TestExecutionRejectsUnsafeCalls(t *testing.T) {
	for _, tc := range []struct {
		name, tool, risk, code string
		params                 map[string]interface{}
		denied                 bool
	}{
		{name: "unknown", tool: "missing", risk: string(RiskReadonlyLow), code: "UNKNOWN_TOOL"},
		{name: "write", tool: "query", risk: string(RiskWriteConfirm), code: "READ_ONLY_REQUIRED"},
		{name: "missing", tool: "query", risk: string(RiskReadonlyLow), code: "INVALID_PARAMS"},
		{name: "scope injection", tool: "query", risk: string(RiskReadonlyLow), code: "INVALID_PARAMS", params: map[string]interface{}{"datasource_id": float64(1), "_workspace_id": float64(2)}},
		{name: "fractional id", tool: "query", risk: string(RiskReadonlyLow), code: "INVALID_PARAMS", params: map[string]interface{}{"datasource_id": 1.5}},
		{name: "unauthorized", tool: "query", risk: string(RiskReadonlyLow), code: "FORBIDDEN", params: map[string]interface{}{"datasource_id": float64(1)}, denied: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			o := testOrchestrator(t, func(context.Context, ToolCall) ToolResult { t.Fatal("denied tool executed"); return ToolResult{} })
			tool, _ := o.Registry.Get("query")
			tool.Spec.Risk = RiskLevel(tc.risk)
			o.Registry.tools["query"] = tool
			if tc.denied {
				o.Authorize = nil
			}
			result := o.execute(context.Background(), ToolCall{Tool: tc.tool, Params: tc.params})
			if result.OK || result.ErrorCode != tc.code {
				t.Fatalf("unexpected result: %#v", result)
			}
		})
	}
}

func TestExecutionCancellationStopsFurtherCalls(t *testing.T) {
	o := testOrchestrator(t, func(ctx context.Context, _ ToolCall) ToolResult { <-ctx.Done(); return ToolResult{OK: true} })
	o.Timeout = 10 * time.Millisecond
	calls := 0
	o.LLM = func(context.Context, string, []Message) (string, error) {
		calls++
		return `{"type":"tool_call","tool":"query","params":{"datasource_id":1}}`, nil
	}
	result := o.Run(context.Background(), RunRequest{})
	trace := result.Metadata["execution"].(RunTrace)
	if calls != 1 || trace.Status != "timeout" || len(trace.Steps) != 2 || trace.Steps[1].ErrorCode != "TOOL_TIMEOUT" {
		t.Fatalf("unexpected trace: %#v", trace)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result = o.Run(ctx, RunRequest{})
	if calls != 1 || result.Metadata["execution"].(RunTrace).Status != "cancelled" {
		t.Fatal("cancelled run called model")
	}
}

func TestExecutionRoundLimitAndSuccess(t *testing.T) {
	o := testOrchestrator(t, func(context.Context, ToolCall) ToolResult { return ToolResult{OK: true} })
	o.MaxRounds = 2
	calls := 0
	o.LLM = func(context.Context, string, []Message) (string, error) {
		calls++
		return `{"type":"tool_call","tool":"query","params":{"datasource_id":1}}`, nil
	}
	result := o.Run(context.Background(), RunRequest{})
	if calls != 2 || result.Metadata["execution"].(RunTrace).Status != "limited" {
		t.Fatal("round limit ignored")
	}
	o.LLM = func(context.Context, string, []Message) (string, error) {
		return `{"type":"final","answer":"完成"}`, nil
	}
	result = o.Run(context.Background(), RunRequest{})
	if result.Answer != "完成" || result.Metadata["execution"].(RunTrace).Status != "succeeded" {
		t.Fatalf("unexpected answer %#v", result)
	}
}

func TestBoundedResultIsValidJSON(t *testing.T) {
	text := boundedResult(ToolResult{OK: true, Data: strings.Repeat("中文数据", 10000)}, 1024)
	if len(text) > 1024 || !json.Valid([]byte(text)) || !utf8.ValidString(text) || !strings.Contains(text, "truncated_for_llm") {
		t.Fatalf("invalid truncated result %q", text)
	}
}

func TestToolReceivesDeclaredDeadline(t *testing.T) {
	o := testOrchestrator(t, func(ctx context.Context, _ ToolCall) ToolResult {
		deadline, ok := ctx.Deadline()
		if !ok || time.Until(deadline) > time.Second {
			t.Fatal("declared timeout was not applied")
		}
		return ToolResult{OK: true}
	})
	result := o.execute(context.Background(), ToolCall{Tool: "query", Params: map[string]interface{}{"datasource_id": float64(1)}})
	if !result.OK {
		t.Fatalf("unexpected result: %#v", result)
	}
}
