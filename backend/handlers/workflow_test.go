package handlers

import (
	"testing"
)

func TestValidateWorkflowRequestRejectsLegacyGraph(t *testing.T) {
	_, err := validateWorkflowRequest(CreateWorkflowReq{
		Name: "旧工作流", Status: 1, Steps: `[{"type":"script","id":1}]`,
	})
	if err == nil {
		t.Fatal("workflow without the current graph version must be rejected")
	}
}

func TestValidateWorkflowRequestRejectsEmptyWorkflow(t *testing.T) {
	if _, err := validateWorkflowRequest(CreateWorkflowReq{Name: "empty", Steps: `[]`, Status: 1}); err == nil {
		t.Fatal("empty workflow should be rejected")
	}
}

func TestValidateWorkflowAcceptsScriptAutomationWithEnd(t *testing.T) {
	steps, err := validateWorkflowRequest(CreateWorkflowReq{
		Name:   "脚本自动化",
		Status: 1,
		Steps:  `[{"key":"prepare","type":"script","id":1,"graph_version":2,"start":true,"next":["done"]},{"key":"done","type":"end","graph_version":2,"result_status":"success","input":"{{ steps.prepare.output }}"}]`,
	})
	if err != nil {
		t.Fatal(err)
	}
	if steps[1].ResultStatus != "success" || len(steps[1].Next) != 0 {
		t.Fatalf("unexpected end step: %+v", steps[1])
	}
}

func TestValidateWorkflowRejectsOpenBranchWhenEndExists(t *testing.T) {
	_, err := validateWorkflowRequest(CreateWorkflowReq{
		Name:   "未闭合流程",
		Status: 1,
		Steps:  `[{"key":"prepare","type":"script","id":1,"graph_version":2,"start":true},{"key":"done","type":"end","graph_version":2,"start":true,"result_status":"success"}]`,
	})
	if err == nil {
		t.Fatal("non-end terminal branch should be rejected")
	}
}

func TestExplicitEndControlsWorkflowFinalStatus(t *testing.T) {
	current := workflowFinalStatus(true, "success", WorkflowStep{Type: "script"}, stepResult{Status: "fail"})
	if current != "success" {
		t.Fatalf("handled script failure should not decide final status: %s", current)
	}
	current = workflowFinalStatus(true, current, WorkflowStep{Type: "end"}, stepResult{Status: "fail"})
	if current != "fail" {
		t.Fatalf("failed end should decide final status: %s", current)
	}
}

func TestRenderWorkflowTemplateUsesInputAndStepOutput(t *testing.T) {
	runtime := map[string]interface{}{
		"input": map[string]interface{}{"alert": map[string]interface{}{"target": "api-1"}},
		"steps": map[string]interface{}{"lookup": map[string]interface{}{"output": map[string]interface{}{"port": float64(8080)}}},
	}
	got := renderWorkflowTemplate(`http://{{ input.alert.target }}:{{ steps.lookup.output.port }}/health`, runtime)
	if got != "http://api-1:8080/health" {
		t.Fatalf("unexpected rendered template: %s", got)
	}
}

func TestEvaluateWorkflowCondition(t *testing.T) {
	runtime := map[string]interface{}{"input": map[string]interface{}{"severity": "critical", "count": float64(3)}}
	if matched, _ := evaluateWorkflowCondition(runtime, "input.severity", "equals", "critical"); !matched {
		t.Fatal("string condition should match")
	}
	if matched, _ := evaluateWorkflowCondition(runtime, "input.count", "greater_than", "2"); !matched {
		t.Fatal("numeric condition should match")
	}
}

func TestValidateWorkflowConditionAndOutputKeys(t *testing.T) {
	steps, err := validateWorkflowRequest(CreateWorkflowReq{Name: "条件处置", Status: 1, Steps: `[{"key":"check_severity","type":"condition","graph_version":2,"start":true,"condition_path":"input.severity","condition_operator":"equals","condition_value":"critical","true_next":"handle_critical","false_next":"handle_normal"},{"key":"handle_critical","type":"end","graph_version":2,"result_status":"success"},{"key":"handle_normal","type":"end","graph_version":2,"result_status":"fail"}]`})
	if err != nil {
		t.Fatal(err)
	}
	if steps[0].TrueNext != "handle_critical" || steps[0].FalseNext != "handle_normal" {
		t.Fatalf("unexpected branches: %+v", steps[0])
	}
}

func TestValidateWorkflowRejectsLegacyNodeTypes(t *testing.T) {
	for _, nodeType := range []string{"input", "http", "inspection", "delay", "notification", "output"} {
		step := WorkflowStep{Key: "legacy", Type: nodeType, GraphVersion: 2}
		if err := normalizeAndValidateWorkflowStep(&step); err == nil {
			t.Fatalf("legacy node type %q must be rejected", nodeType)
		}
	}
}

func TestValidateWorkflowAcceptsParallelGraph(t *testing.T) {
	steps, err := validateWorkflowRequest(CreateWorkflowReq{Name: "并行处置", Status: 1, Steps: `[{"key":"start_service","type":"script","id":1,"graph_version":2,"start":true,"next":["verify"]},{"key":"request_data","type":"script","id":2,"graph_version":2,"start":true,"next":["verify"]},{"key":"verify","type":"script","id":3,"graph_version":2}]`})
	if err != nil {
		t.Fatal(err)
	}
	if !steps[0].Start || !steps[1].Start || len(steps[0].Next) != 1 {
		t.Fatalf("unexpected parallel graph: %+v", steps)
	}
}

func TestValidateWorkflowRejectsGraphCycle(t *testing.T) {
	_, err := validateWorkflowRequest(CreateWorkflowReq{Name: "循环图", Status: 1, Steps: `[{"key":"one","type":"script","id":1,"graph_version":2,"start":true,"next":["two"]},{"key":"two","type":"script","id":2,"graph_version":2,"next":["one"]}]`})
	if err == nil {
		t.Fatal("cyclic workflow graph should be rejected")
	}
}
