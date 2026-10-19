package handlers

import (
	"strings"
	"testing"
	"time"
)

func TestNotificationVarsHandlesNilLifecycleTimes(t *testing.T) {
	now := time.Date(2026, 8, 31, 11, 3, 0, 0, time.Local)
	var recoveredAt *time.Time

	vars := notificationVars(map[string]interface{}{
		"fired_at":     &now,
		"recovered_at": recoveredAt,
	})

	if vars["fired_at"] != "2026-08-31 11:03:00" {
		t.Fatalf("unexpected firing time: %q", vars["fired_at"])
	}
	if vars["recovered_at"] != "" || vars["recover_time_format"] != "" {
		t.Fatalf("nil recovery time must render as empty text: %#v", vars)
	}
}

func TestSystemNotificationEnvelopeFixesTitleColorAndFooter(t *testing.T) {
	firing := map[string]interface{}{"title": "CPU 使用率过高", "title_color": "green", "content": "正文", "status": "firing", "severity": "P0"}
	applySystemNotificationEnvelope(firing)
	if firing["title"] != "【告警中】- 运维中枢 🚨" || firing["title_color"] != "red" {
		t.Fatalf("unexpected firing envelope: %#v", firing)
	}
	if !strings.HasSuffix(firing["content"].(string), "🧑‍💻 运维中枢 · 运维团队") {
		t.Fatalf("fixed footer is missing: %#v", firing)
	}

	recovered := map[string]interface{}{"title": "CPU 使用率过高", "content": "恢复正文", "status": "resolved", "severity": "P0"}
	applySystemNotificationEnvelope(recovered)
	if recovered["title"] != "【已恢复】- 运维中枢 ✅" || recovered["title_color"] != "green" {
		t.Fatalf("unexpected recovery envelope: %#v", recovered)
	}

	applySystemNotificationEnvelope(recovered)
	if strings.Count(recovered["content"].(string), "🧑‍💻 运维中枢 · 运维团队") != 1 {
		t.Fatalf("fixed footer must not be duplicated: %#v", recovered)
	}
}

func TestNotificationTemplateRejectsDefinedSections(t *testing.T) {
	content := `{{- define "Event" -}}legacy{{- end -}}`
	if err := validateNotificationTemplateContent(content); err == nil {
		t.Fatal("templates with legacy defined sections must be rejected")
	}
}
