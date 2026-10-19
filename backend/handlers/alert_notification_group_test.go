package handlers

import (
	"aiops-backend/models"
	"reflect"
	"testing"
	"time"
)

func TestAlertGroupStateAllowsMissingLastNotificationTime(t *testing.T) {
	field, ok := reflect.TypeOf(models.AlertGroupState{}).FieldByName("LastNotifiedAt")
	if !ok || field.Type.Kind() != reflect.Ptr {
		t.Fatalf("a new group has not been notified yet, so LastNotifiedAt must be nullable; got %v", field.Type)
	}
}

func TestAlertNotificationGroupKeyUsesConfiguredLabels(t *testing.T) {
	namespaces := map[string]map[string]string{
		"system_labels":     {"alertname": "NodeDown"},
		"labels":            {"instance": "node-1", "cluster": "prod", "env": "metric"},
		"datasource_labels": {"env": "prod"},
		"external_labels":   {"env": "service"},
	}
	groupBy := `[{"namespace":"datasource_labels","key":"env"},{"namespace":"system_labels","key":"alertname"}]`
	got := alertNotificationGroupKey(groupBy, namespaces)
	if got != "datasource_labels.env=prod|system_labels.alertname=NodeDown" {
		t.Fatalf("unexpected group key %q", got)
	}
	if alertNotificationGroupKey("", namespaces) != "" {
		t.Fatal("empty group_by must disable grouping")
	}
}

func TestAlertNotificationGroupKeyRejectsLegacyPromQLKeys(t *testing.T) {
	namespaces := map[string]map[string]string{"labels": {"cluster": "prod"}}
	got := alertNotificationGroupKey("cluster, alertname", namespaces)
	if got != "" {
		t.Fatalf("legacy group keys must not be interpreted, got %q", got)
	}
}

func TestAlertGroupingRuleKeyIsolatesPoliciesAndSupportsOneLargeGroup(t *testing.T) {
	namespaces := map[string]map[string]string{
		"system_labels": {"alertname": "NodeDown"},
		"labels":        {"cluster": "prod"},
	}
	allAlerts := models.AlertGroupingRule{ID: 7, GroupByJSON: "[]"}
	if got := alertGroupingRuleKey(allAlerts, namespaces); got != "rule:7" {
		t.Fatalf("empty group_by must aggregate all matched alerts into one rule group, got %q", got)
	}
	byCluster := models.AlertGroupingRule{ID: 8, GroupByJSON: `[{"namespace":"labels","key":"cluster"}]`}
	if got := alertGroupingRuleKey(byCluster, namespaces); got != "rule:8|labels.cluster=prod" {
		t.Fatalf("group key must include the rule id and selected values, got %q", got)
	}
}

func TestSelectAlertGroupingRulePrefersSpecificRuleAtSamePriority(t *testing.T) {
	alert := models.AlertInstance{LastResult: `{"datasource_labels":{"env":"prod"}}`}
	rules := []models.AlertGroupingRule{
		{ID: 1, Priority: 100, MatchersJSON: "[]"},
		{ID: 2, Priority: 100, MatchersJSON: `[{"namespace":"datasource_labels","key":"env","value":"prod"}]`},
	}
	selected, ok := selectAlertGroupingRule(rules, alert)
	if !ok || selected.ID != 2 {
		t.Fatalf("specific rule must win over catch-all at the same priority, got %#v", selected)
	}
	rules = append([]models.AlertGroupingRule{{ID: 3, Priority: 50, MatchersJSON: "[]"}}, rules...)
	selected, ok = selectAlertGroupingRule(rules, alert)
	if !ok || selected.ID != 3 {
		t.Fatalf("a higher-priority catch-all must still win, got %#v", selected)
	}
}

func TestNextAlertGroupNotificationUsesFirstSeenAndGroupInterval(t *testing.T) {
	now := time.Date(2026, 8, 22, 10, 0, 0, 0, time.UTC)

	status, due := nextAlertGroupNotification(models.AlertGroupState{}, 30, 300, now)
	if status != notificationStatusWaitingGroup || !due.Equal(now.Add(30*time.Second)) {
		t.Fatalf("first alert should start group_wait: status=%q due=%s", status, due)
	}

	lastNotifiedAt := now.Add(-2 * time.Minute)
	state := models.AlertGroupState{FirstSeenAt: now.Add(-10 * time.Minute), LastNotifiedAt: &lastNotifiedAt}
	status, due = nextAlertGroupNotification(state, 30, 300, now)
	if status != notificationStatusWaitingUpdate || !due.Equal(state.LastNotifiedAt.Add(5*time.Minute)) {
		t.Fatalf("group update should use group_interval: status=%q due=%s", status, due)
	}
}

func TestP0UsesConfiguredGroupWaitAndInterval(t *testing.T) {
	now := time.Date(2026, 8, 22, 10, 0, 0, 0, time.UTC)
	status, due := nextAlertGroupNotification(models.AlertGroupState{}, 30, 300, now)
	if status != notificationStatusWaitingGroup || !due.Equal(now.Add(30*time.Second)) {
		t.Fatalf("P0 first notification must use group_wait: status=%q due=%s", status, due)
	}

	lastNotifiedAt := now.Add(-time.Minute)
	state := models.AlertGroupState{FirstSeenAt: now.Add(-time.Minute), LastNotifiedAt: &lastNotifiedAt}
	status, due = nextAlertGroupNotification(state, 30, 300, now)
	if status != notificationStatusWaitingUpdate || !due.Equal(lastNotifiedAt.Add(300*time.Second)) {
		t.Fatalf("P0 group update must use group_interval: status=%q due=%s", status, due)
	}
}

func TestResolvedBeforeFirstNotificationIsRecordedAsUnnotified(t *testing.T) {
	if got := resolvedAlertNotificationStatus(nil); got != notificationStatusResolvedUnnotified {
		t.Fatalf("never-notified recovery should be explicit, got %q", got)
	}
	notifiedAt := time.Now()
	if got := resolvedAlertNotificationStatus(&notifiedAt); got != notificationStatusWaitingRecovery {
		t.Fatalf("notified recovery should wait for grouped recovery, got %q", got)
	}
}

func TestGroupedAlertPayloadContainsEveryMember(t *testing.T) {
	now := time.Date(2026, 8, 22, 10, 0, 0, 0, time.UTC)
	firedAt := now.Add(-2 * time.Minute)
	first := &models.AlertInstance{ID: 1, Title: "节点不可用", Target: "node-1", Severity: "P0", Labels: `{"cluster":"prod"}`, Fingerprint: "fingerprint-1", FiredAt: &firedAt, LastResult: `{"alertname":"节点不可用","datasource":"生产 Prometheus","value":"1","condition":"== 1","annotations":{"description":"节点探测失败"}}`}
	second := &models.AlertInstance{ID: 2, Title: "节点负载高", Target: "node-2", Severity: "P1", Labels: `{"cluster":"prod"}`}
	title, payload := groupedAlertNotificationPayload(
		models.FaultCenter{ID: 7, Name: "生产故障中心"},
		models.AlertGroupState{GroupKey: "cluster=prod"},
		[]groupedAlertDelivery{{alert: first, status: "firing"}, {alert: second, status: "firing"}},
		1,
		now,
	)
	if title != "告警分组：生产故障中心（2 条）" {
		t.Fatalf("unexpected title %q", title)
	}
	if payload["firing_count"] != 2 || payload["suppressed_count"] != 1 {
		t.Fatalf("unexpected grouped counters: %#v", payload)
	}
	alerts, ok := payload["alerts"].([]map[string]interface{})
	if !ok || len(alerts) != 2 {
		t.Fatalf("every member must be present in the grouped payload: %#v", payload["alerts"])
	}
	if payload["title_color"] != "red" || payload["severity"] != "P0" {
		t.Fatalf("grouped firing alert must use the highest active severity: %#v", payload)
	}
	if alerts[0]["alertname"] != "节点不可用" || alerts[0]["datasource"] != "生产 Prometheus" || alerts[0]["value"] != "1" || alerts[0]["fingerprint"] != "fingerprint-1" {
		t.Fatalf("group member must expose its structured alert data: %#v", alerts[0])
	}
	if alerts[0]["first_trigger_time_format"] != "2026-08-22 09:58:00" || alerts[0]["recover_time_format"] != "" {
		t.Fatalf("group member must expose formatted lifecycle times: %#v", alerts[0])
	}
}

func TestAlertNotificationTitleColorFollowsSeverityAndRecovery(t *testing.T) {
	tests := []struct {
		severity  string
		recovered bool
		want      string
	}{
		{severity: "P0", want: "red"},
		{severity: "P1", want: "orange"},
		{severity: "P2", want: "blue"},
		{severity: "unknown", want: "blue"},
		{severity: "P0", recovered: true, want: "green"},
	}
	for _, test := range tests {
		if got := alertNotificationTitleColor(test.severity, test.recovered); got != test.want {
			t.Fatalf("severity=%q recovered=%t: got %q, want %q", test.severity, test.recovered, got, test.want)
		}
	}
}

func TestResolvedAlertGroupUsesGreenCard(t *testing.T) {
	now := time.Date(2026, 8, 22, 10, 0, 0, 0, time.UTC)
	alert := &models.AlertInstance{ID: 1, Title: "节点不可用", Target: "node-1", Severity: "P0"}
	_, payload := groupedAlertNotificationPayload(
		models.FaultCenter{ID: 7, Name: "生产故障中心"},
		models.AlertGroupState{GroupKey: "cluster=prod"},
		[]groupedAlertDelivery{{alert: alert, status: "resolved"}},
		0,
		now,
	)
	if payload["title_color"] != "green" || payload["status"] != "resolved" {
		t.Fatalf("fully resolved group must use a green card: %#v", payload)
	}
}

func TestCollapseAlertGroupRowsKeepsLatestStatePerFingerprint(t *testing.T) {
	now := time.Date(2026, 8, 31, 17, 35, 20, 0, time.Local)
	oldFiredAt := now.Add(-5 * time.Minute)
	oldResolvedAt := now.Add(-2 * time.Minute)
	rows := []models.AlertInstance{
		{ID: 748, Title: "内存使用率过高", Fingerprint: "memory", Status: alertStatusResolved, NotificationStatus: notificationStatusWaitingRecovery, FiredAt: &oldFiredAt, ResolvedAt: &oldResolvedAt},
		{ID: 749, Title: "CPU 使用率过高", Fingerprint: "cpu", Status: alertStatusResolved, NotificationStatus: notificationStatusWaitingRecovery, FiredAt: &oldFiredAt, ResolvedAt: &oldResolvedAt},
		{ID: 750, Title: "CPU 使用率过高", Fingerprint: "cpu", Status: alertStatusFiring, FiredAt: &now},
		{ID: 751, Title: "内存使用率过高", Fingerprint: "memory", Status: alertStatusFiring, FiredAt: &now},
	}

	selected, superseded := collapseAlertGroupRows(rows)

	if len(selected) != 2 || selected[0].ID != 750 || selected[1].ID != 751 {
		t.Fatalf("the latest firing incident must replace an unsent recovery with the same fingerprint: %#v", selected)
	}
	supersededIDs := map[uint]bool{}
	for _, row := range superseded {
		supersededIDs[row.ID] = true
	}
	if len(superseded) != 2 || !supersededIDs[748] || !supersededIDs[749] {
		t.Fatalf("old unsent recoveries must be marked as superseded: %#v", superseded)
	}
}

func TestCollapseAlertGroupRowsPreservesDifferentFingerprints(t *testing.T) {
	rows := []models.AlertInstance{
		{ID: 1, Fingerprint: "cpu", Status: alertStatusFiring},
		{ID: 2, Fingerprint: "memory", Status: alertStatusResolved},
	}
	selected, superseded := collapseAlertGroupRows(rows)
	if len(selected) != 2 || len(superseded) != 0 {
		t.Fatalf("different fingerprints must remain separate: selected=%#v superseded=%#v", selected, superseded)
	}
}
