package handlers

import (
	"aiops-backend/models"
	"testing"
	"time"
)

func TestAlertActionTransitionTriggersOnlyLifecycleBoundaries(t *testing.T) {
	tests := []struct {
		eventType string
		status    string
		want      bool
	}{
		{"created", alertStatusFiring, true},
		{"state_changed", alertStatusFiring, true},
		{"state_changed", alertStatusResolved, true},
		{"created", alertStatusPending, false},
		{"notification_sent", alertStatusFiring, false},
		{"state_changed", alertStatusTerminated, false},
	}
	for _, test := range tests {
		if got := alertActionTransitionTriggered(test.eventType, test.status); got != test.want {
			t.Fatalf("alertActionTransitionTriggered(%q, %q) = %v, want %v", test.eventType, test.status, got, test.want)
		}
	}
}

func TestAlertActionScopeKeyUsesAllNamespacesDeterministically(t *testing.T) {
	first := namespacedAlert(`{"instance":"node-1"}`, `{"env":"prod"}`, `{"service":"order"}`)
	second := namespacedAlert(`{"instance":"node-1"}`, `{"env":"prod"}`, `{"service":"order"}`)
	if alertActionScopeKey(first) != alertActionScopeKey(second) {
		t.Fatal("same label namespaces should produce the same scope key")
	}
	second = namespacedAlert(`{"instance":"node-2"}`, `{"env":"prod"}`, `{"service":"order"}`)
	if alertActionScopeKey(first) == alertActionScopeKey(second) {
		t.Fatal("changing a runtime label should change the scope key")
	}
}

func TestBuildAlertWorkflowInputProvidesStructuredAlert(t *testing.T) {
	alert := models.AlertInstance{ID: 12, Title: "API 不可用", Target: "api-1", Severity: "critical", LastResult: `{"labels":{"instance":"api-1"}}`}
	input := buildAlertWorkflowInput(alert, alertStatusFiring, time.Now())
	alertObject, ok := input["alert"].(map[string]interface{})
	if !ok || alertObject["target"] != "api-1" {
		t.Fatalf("workflow alert input is not structured: %#v", input["alert"])
	}
	labels, ok := input["labels"].(map[string]string)
	if !ok || labels["instance"] != "api-1" {
		t.Fatalf("workflow label input is missing: %#v", input["labels"])
	}
}

func TestAlertActionRetryDelayIsBoundedExponential(t *testing.T) {
	want := []time.Duration{5, 10, 20, 40, 80, 160, 160}
	for index, seconds := range want {
		got := alertActionRetryDelay(index + 1)
		if got != seconds*time.Second {
			t.Fatalf("attempt %d retry delay = %s, want %s", index+1, got, seconds*time.Second)
		}
	}
}

func TestAlertActionWebhookSignatureUsesHMACSHA256(t *testing.T) {
	got := alertActionWebhookSignature([]byte("secret"), []byte("payload"))
	want := "sha256=b82fcb791acec57859b989b430a826488ce2e479fdf92326bd0a2e8375a42ba4"
	if got != want {
		t.Fatalf("signature = %q, want %q", got, want)
	}
}

func TestAlertActionMatcherSummaryIsStable(t *testing.T) {
	matchers := []models.AlertLabelMatcher{
		{Namespace: "labels", Key: "instance", Value: "node-1"},
		{Namespace: "external_labels", Key: "service", Value: "order"},
	}
	want := "external_labels.service=order, labels.instance=node-1"
	if got := sortedAlertActionMatcherSummary(matchers); got != want {
		t.Fatalf("matcher summary = %q, want %q", got, want)
	}
}
