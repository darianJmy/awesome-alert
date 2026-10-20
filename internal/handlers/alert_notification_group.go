package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"guanlan/internal/database"
	"guanlan/internal/models"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm/clause"
)

const (
	notificationStatusWaitingGroup       = "waiting_group"
	notificationStatusWaitingUpdate      = "waiting_update"
	notificationStatusWaitingRecovery    = "waiting_recovery"
	notificationStatusSent               = "sent"
	notificationStatusRecoverySent       = "recovery_sent"
	notificationStatusSilenced           = "silenced"
	notificationStatusInhibited          = "inhibited"
	notificationStatusNoReceiver         = "no_receiver"
	notificationStatusSendFailed         = "send_failed"
	notificationStatusResolvedUnnotified = "resolved_unnotified"
	notificationStatusRecoveryDisabled   = "recovery_disabled"
	notificationStatusTerminated         = "terminated"
)

type groupedAlertDelivery struct {
	alert    *models.AlertInstance
	status   string
	channels []uint
}

func alertNotificationGroupKey(groupBy string, namespaces map[string]map[string]string) string {
	parts := make([]string, 0)
	for _, selector := range decodeAlertGroupingSelectors(groupBy) {
		labels := namespaces[selector.Namespace]
		parts = append(parts, selector.Namespace+"."+selector.Key+"="+labels[selector.Key])
	}
	return strings.Join(parts, "|")
}

func alertGroupingRuleKey(rule models.AlertGroupingRule, namespaces map[string]map[string]string) string {
	parts := []string{fmt.Sprintf("rule:%d", rule.ID)}
	for _, selector := range decodeAlertGroupingSelectors(rule.GroupByJSON) {
		parts = append(parts, selector.Namespace+"."+selector.Key+"="+namespaces[selector.Namespace][selector.Key])
	}
	return strings.Join(parts, "|")
}

func matchingAlertGroupingRule(alert models.AlertInstance) (models.AlertGroupingRule, bool) {
	var rules []models.AlertGroupingRule
	if database.DB.Where("fault_center_id = ? AND status = 1", alert.FaultCenterID).Order("priority ASC, id ASC").Find(&rules).Error != nil {
		return models.AlertGroupingRule{}, false
	}
	return selectAlertGroupingRule(rules, alert)
}

func selectAlertGroupingRule(rules []models.AlertGroupingRule, alert models.AlertInstance) (models.AlertGroupingRule, bool) {
	for start := 0; start < len(rules); {
		end := start + 1
		for end < len(rules) && rules[end].Priority == rules[start].Priority {
			end++
		}
		// At the same priority, a conditional rule is more specific than the
		// catch-all rule. This keeps a default policy from shadowing a newly
		// added conditional policy that uses the same priority.
		for index := start; index < end; index++ {
			matchers := decodeAlertLabelMatchers(rules[index].MatchersJSON)
			if len(matchers) > 0 && alertMatchesLabelMatchers(alert, matchers) {
				return rules[index], true
			}
		}
		for index := start; index < end; index++ {
			if len(decodeAlertLabelMatchers(rules[index].MatchersJSON)) == 0 {
				return rules[index], true
			}
		}
		start = end
	}
	return models.AlertGroupingRule{}, false
}

func nextAlertGroupNotification(state models.AlertGroupState, waitSeconds, intervalSeconds int, now time.Time) (string, time.Time) {
	if state.ClosedAt != nil {
		state.FirstSeenAt = time.Time{}
		state.LastNotifiedAt = nil
	}
	if waitSeconds < 0 {
		waitSeconds = 30
	}
	if intervalSeconds < 60 {
		intervalSeconds = 300
	}
	if state.LastNotifiedAt == nil {
		firstSeen := state.FirstSeenAt
		if firstSeen.IsZero() {
			firstSeen = now
		}
		return notificationStatusWaitingGroup, firstSeen.Add(time.Duration(waitSeconds) * time.Second)
	}
	return notificationStatusWaitingUpdate, state.LastNotifiedAt.Add(time.Duration(intervalSeconds) * time.Second)
}

func resolvedAlertNotificationStatus(lastNotifiedAt *time.Time) string {
	if lastNotifiedAt == nil {
		return notificationStatusResolvedUnnotified
	}
	return notificationStatusWaitingRecovery
}

func alertNotificationSuppression(alert *models.AlertInstance, now time.Time) (string, string) {
	applyActiveSilenceRule(alert, now)
	if alert.SilencedUntil != nil && alert.SilencedUntil.After(now) {
		return notificationStatusSilenced, "命中告警静默策略"
	}
	var rules []models.AlertInhibitionRule
	database.DB.Where("fault_center_id = ? AND status = 1", alert.FaultCenterID).Find(&rules)
	for _, rule := range rules {
		targetMatchers := decodeAlertLabelMatchers(rule.TargetMatchersJSON)
		if !alertMatchesLabelMatchers(*alert, targetMatchers) {
			continue
		}
		sourceMatchers := decodeAlertLabelMatchers(rule.SourceMatchersJSON)
		equalLabels := decodeAlertLabelSelectors(rule.EqualLabelsJSON)
		var sources []models.AlertInstance
		database.DB.Where("fault_center_id = ? AND status = ? AND id <> ?", alert.FaultCenterID, alertStatusFiring, alert.ID).Find(&sources)
		for _, source := range sources {
			if (source.SilencedUntil == nil || !source.SilencedUntil.After(now)) && alertMatchesLabelMatchers(source, sourceMatchers) && inhibitionEqualLabelsMatch(source, *alert, equalLabels) {
				return notificationStatusInhibited, "被抑制规则“" + rule.Name + "”拦截"
			}
		}
	}
	return "", ""
}

func updateAlertNotificationState(alert *models.AlertInstance, status, reason, groupKey string, next *time.Time) {
	if alert == nil || alert.ID == 0 {
		return
	}
	updates := map[string]interface{}{
		"notification_status": status, "notification_reason": reason,
		"notification_group_key": groupKey, "next_notification_at": next,
	}
	database.DB.Model(&models.AlertInstance{}).Where("id = ?", alert.ID).Updates(updates)
	alert.NotificationStatus, alert.NotificationReason, alert.NotificationGroupKey, alert.NextNotificationAt = status, reason, groupKey, next
}

// queueGroupedFiringAlert returns true when grouping owns notification delivery.
func queueGroupedFiringAlert(alert *models.AlertInstance, config alertRuleNotificationConfig, now time.Time) bool {
	if alert == nil || alert.ID == 0 || alert.FaultCenterID == 0 {
		return false
	}
	var center models.FaultCenter
	if database.DB.First(&center, alert.FaultCenterID).Error != nil || center.Status != 1 {
		return false
	}
	rule, matched := matchingAlertGroupingRule(*alert)
	if !matched {
		return false
	}
	if status, reason := alertNotificationSuppression(alert, now); status != "" {
		updateAlertNotificationState(alert, status, reason, "", nil)
		return true
	}
	if len(notificationRouteChannelIDs(*alert, config.Routes, config.ChannelIDs)) == 0 {
		updateAlertNotificationState(alert, notificationStatusNoReceiver, "没有匹配到可用通知对象", "", nil)
		return true
	}
	groupKey := alertGroupingRuleKey(rule, alertLabelNamespaces(*alert))
	// An unchanged alert is repeated by the group scheduler, not by rule evaluation.
	stableStatus := alert.NotificationStatus == notificationStatusSent || alert.NotificationStatus == notificationStatusWaitingGroup || alert.NotificationStatus == notificationStatusWaitingUpdate
	if alert.LastNotifiedAt != nil && !alert.StateChangedAt.After(*alert.LastNotifiedAt) && alert.NotificationGroupKey == groupKey && stableStatus {
		return true
	}

	state, created := findOrCreateAlertGroupState(alert.FaultCenterID, rule.ID, groupKey, now)
	if state.ID == 0 {
		// Do not leave the alert stuck in a grouping status when the durable
		// scheduler state could not be created. The caller can fall back to
		// direct delivery for this evaluation.
		return false
	}
	status, due := nextAlertGroupNotification(state, rule.WaitSeconds, rule.IntervalSeconds, now)
	if due.Before(now) {
		due = now
	}
	stateDue := due
	if state.Pending && state.NextNotificationAt != nil && state.NextNotificationAt.Before(stateDue) {
		stateDue = *state.NextNotificationAt
	}
	updates := map[string]interface{}{"pending": true, "next_notification_at": stateDue, "lease_until": nil, "closed_at": nil}
	if created || state.FirstSeenAt.IsZero() || state.ClosedAt != nil {
		updates["first_seen_at"] = now
	}
	database.DB.Model(&models.AlertGroupState{}).Where("id = ?", state.ID).Updates(updates)
	updateAlertNotificationState(alert, status, "等待同组告警聚合发送", groupKey, &stateDue)
	return true
}

func queueGroupedRecoveryAlert(alert *models.AlertInstance, config alertRuleNotificationConfig, now time.Time) bool {
	if alert == nil || alert.ID == 0 {
		return false
	}
	status := resolvedAlertNotificationStatus(alert.LastNotifiedAt)
	if status == notificationStatusResolvedUnnotified {
		updateAlertNotificationState(alert, status, "在首次分组通知前恢复", alert.NotificationGroupKey, nil)
		_ = recordAlertEvent(database.DB, *alert, "notification_skipped", alert.Status, alert.Status, alert.Severity, alert.Severity, "resolved_before_first_notification", "", alert.LastProbeStatus, now)
		return true
	}
	if !config.RecoveryNotification {
		updateAlertNotificationState(alert, notificationStatusRecoveryDisabled, "恢复通知未启用", alert.NotificationGroupKey, nil)
		return true
	}
	var center models.FaultCenter
	if database.DB.First(&center, alert.FaultCenterID).Error != nil || center.Status != 1 {
		return false
	}
	groupKey := alert.NotificationGroupKey
	if groupKey == "" {
		rule, matched := matchingAlertGroupingRule(*alert)
		if !matched {
			return false
		}
		groupKey = alertGroupingRuleKey(rule, alertLabelNamespaces(*alert))
	}
	var state models.AlertGroupState
	if groupKey == "" || database.DB.Where("fault_center_id = ? AND group_key = ?", alert.FaultCenterID, groupKey).First(&state).Error != nil {
		return false
	}
	var rule models.AlertGroupingRule
	if state.GroupingRuleID == 0 || database.DB.Where("id = ? AND fault_center_id = ? AND status = 1", state.GroupingRuleID, alert.FaultCenterID).First(&rule).Error != nil {
		return false
	}
	if state.LastNotifiedAt == nil {
		return false
	}
	due := state.LastNotifiedAt.Add(time.Duration(maxInt(rule.IntervalSeconds, 60)) * time.Second)
	if due.Before(now) {
		due = now
	}
	if state.Pending && state.NextNotificationAt != nil && state.NextNotificationAt.Before(due) {
		due = *state.NextNotificationAt
	}
	database.DB.Model(&state).Updates(map[string]interface{}{"pending": true, "next_notification_at": due, "lease_until": nil})
	updateAlertNotificationState(alert, notificationStatusWaitingRecovery, "等待分组恢复通知", groupKey, &due)
	return true
}

func findOrCreateAlertGroupState(centerID, ruleID uint, groupKey string, now time.Time) (models.AlertGroupState, bool) {
	var state models.AlertGroupState
	if database.DB.Where("fault_center_id = ? AND group_key = ?", centerID, groupKey).First(&state).Error == nil {
		return state, false
	}
	state = models.AlertGroupState{FaultCenterID: centerID, GroupingRuleID: ruleID, GroupKey: groupKey, FirstSeenAt: now}
	if err := database.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&state).Error; err == nil && state.ID != 0 {
		return state, true
	}
	database.DB.Where("fault_center_id = ? AND group_key = ?", centerID, groupKey).First(&state)
	return state, false
}

func dispatchDueAlertGroups(now time.Time) {
	var states []models.AlertGroupState
	database.DB.Where("pending = ? AND next_notification_at IS NOT NULL AND next_notification_at <= ? AND (lease_until IS NULL OR lease_until < ?)", true, now, now).
		Order("next_notification_at ASC").Limit(100).Find(&states)
	for _, state := range states {
		leaseUntil := now.Add(2 * time.Minute)
		result := database.DB.Model(&models.AlertGroupState{}).
			Where("id = ? AND pending = ? AND next_notification_at <= ? AND (lease_until IS NULL OR lease_until < ?)", state.ID, true, now, now).
			Update("lease_until", leaseUntil)
		if result.RowsAffected == 1 {
			dispatchAlertGroup(state.ID, now)
		}
	}
}

func dispatchAlertGroup(stateID uint, now time.Time) {
	var state models.AlertGroupState
	if database.DB.First(&state, stateID).Error != nil {
		return
	}
	var center models.FaultCenter
	if database.DB.First(&center, state.FaultCenterID).Error != nil || center.Status != 1 {
		finishAlertGroupDispatch(state, false, nil, now)
		return
	}
	var groupingRule models.AlertGroupingRule
	if state.GroupingRuleID == 0 || database.DB.Where("id = ? AND fault_center_id = ? AND status = 1", state.GroupingRuleID, center.ID).First(&groupingRule).Error != nil {
		finishAlertGroupDispatch(state, false, nil, now)
		return
	}
	config := loadAlertRuleNotificationConfig(center.ID)
	var rows []models.AlertInstance
	database.DB.Where("fault_center_id = ? AND notification_group_key = ? AND (status IN ? OR notification_status = ?)", center.ID, state.GroupKey, []string{alertStatusFiring, alertStatusRecovering}, notificationStatusWaitingRecovery).
		Order("fired_at ASC, id ASC").Find(&rows)
	rows, superseded := collapseAlertGroupRows(rows)
	for i := range superseded {
		alert := &superseded[i]
		if alert.Status != alertStatusResolved || alert.NotificationStatus != notificationStatusWaitingRecovery {
			continue
		}
		updateAlertNotificationState(alert, notificationStatusResolvedUnnotified, "恢复通知发送前同一告警已重新触发", state.GroupKey, nil)
		_ = recordAlertEvent(database.DB, *alert, "notification_skipped", alert.Status, alert.Status, alert.Severity, alert.Severity, "retriggered_before_recovery_notification", "", alert.LastProbeStatus, now)
	}

	deliveries := make([]groupedAlertDelivery, 0, len(rows))
	suppressedCount := 0
	for i := range rows {
		alert := &rows[i]
		if alert.Status == alertStatusResolved {
			if alert.LastNotifiedAt == nil {
				updateAlertNotificationState(alert, notificationStatusResolvedUnnotified, "在首次分组通知前恢复", state.GroupKey, nil)
				continue
			}
		} else if status, reason := alertNotificationSuppression(alert, now); status != "" {
			updateAlertNotificationState(alert, status, reason, state.GroupKey, nil)
			suppressedCount++
			continue
		}
		channels := notificationRouteChannelIDs(*alert, config.Routes, config.ChannelIDs)
		if len(channels) == 0 {
			updateAlertNotificationState(alert, notificationStatusNoReceiver, "没有匹配到可用通知对象", state.GroupKey, nil)
			continue
		}
		deliveryStatus := "firing"
		if alert.Status == alertStatusResolved {
			deliveryStatus = "resolved"
		}
		deliveries = append(deliveries, groupedAlertDelivery{alert: alert, status: deliveryStatus, channels: channels})
	}
	if len(deliveries) == 0 {
		finishAlertGroupDispatch(state, false, nil, now)
		return
	}

	byChannel := map[uint][]groupedAlertDelivery{}
	for _, item := range deliveries {
		for _, channelID := range item.channels {
			byChannel[channelID] = append(byChannel[channelID], item)
		}
	}
	succeeded := map[uint]bool{}
	failed := false
	for channelID, items := range byChannel {
		var channel models.NotificationChannel
		if database.DB.Where("id = ? AND status = 1", channelID).First(&channel).Error != nil {
			failed = true
			continue
		}
		title, payload := groupedAlertNotificationPayload(center, state, items, suppressedCount, now)
		eventKey := alertGroupEventKey(center.ID, state.GroupKey, channelID, now)
		if err := sendNotificationToChannel(channel, "alert.group", eventKey, title, payload); err != nil {
			failed = true
			continue
		}
		for _, item := range items {
			succeeded[item.alert.ID] = true
		}
	}

	activeSent := false
	for _, item := range deliveries {
		if !succeeded[item.alert.ID] {
			updateAlertNotificationState(item.alert, notificationStatusSendFailed, "分组通知发送失败，将自动重试", state.GroupKey, timePointer(now.Add(time.Minute)))
			continue
		}
		if item.status == "resolved" {
			updateAlertNotificationState(item.alert, notificationStatusRecoverySent, "恢复通知已随分组发送", state.GroupKey, nil)
		} else {
			activeSent = true
			repeatAt := now.Add(time.Duration(maxInt(config.RepeatSeconds, 60)) * time.Second)
			updateAlertNotificationState(item.alert, notificationStatusSent, "分组通知已发送", state.GroupKey, &repeatAt)
			database.DB.Model(item.alert).Update("last_notified_at", now)
			item.alert.LastNotifiedAt = &now
		}
		_ = recordAlertEvent(database.DB, *item.alert, "notification_sent", item.alert.Status, item.alert.Status, item.alert.Severity, item.alert.Severity, item.status, "", item.alert.LastProbeStatus, now)
	}
	var next *time.Time
	pending := false
	if failed {
		next, pending = timePointer(now.Add(time.Minute)), true
	} else if activeSent {
		next, pending = timePointer(now.Add(time.Duration(maxInt(config.RepeatSeconds, 60))*time.Second)), true
	}
	finishAlertGroupDispatch(state, len(succeeded) > 0, next, now)
	if pending {
		database.DB.Model(&models.AlertGroupState{}).Where("id = ?", state.ID).Update("pending", true)
	}
}

// collapseAlertGroupRows prevents multiple lifecycle records for the same
// alert fingerprint from appearing in one notification. This happens when a
// recovery is still waiting for the group interval and the same alert fires
// again; only the newest incident represents the current state.
func collapseAlertGroupRows(rows []models.AlertInstance) ([]models.AlertInstance, []models.AlertInstance) {
	selected := make([]models.AlertInstance, 0, len(rows))
	superseded := make([]models.AlertInstance, 0)
	indexes := make(map[string]int, len(rows))
	for _, row := range rows {
		key := strings.TrimSpace(row.Fingerprint)
		if key == "" {
			key = fmt.Sprintf("id:%d", row.ID)
		}
		index, exists := indexes[key]
		if !exists {
			indexes[key] = len(selected)
			selected = append(selected, row)
			continue
		}
		if alertGroupRowIsNewer(row, selected[index]) {
			superseded = append(superseded, selected[index])
			selected[index] = row
		} else {
			superseded = append(superseded, row)
		}
	}
	sort.SliceStable(selected, func(i, j int) bool {
		left, right := selected[i].FiredAt, selected[j].FiredAt
		if left != nil && right != nil && !left.Equal(*right) {
			return left.Before(*right)
		}
		if left == nil && right != nil {
			return true
		}
		if left != nil && right == nil {
			return false
		}
		return selected[i].ID < selected[j].ID
	})
	return selected, superseded
}

func alertGroupRowIsNewer(candidate, current models.AlertInstance) bool {
	if candidate.FiredAt != nil && current.FiredAt != nil && !candidate.FiredAt.Equal(*current.FiredAt) {
		return candidate.FiredAt.After(*current.FiredAt)
	}
	if candidate.FiredAt != nil && current.FiredAt == nil {
		return true
	}
	if candidate.FiredAt == nil && current.FiredAt != nil {
		return false
	}
	return candidate.ID > current.ID
}

func groupedAlertNotificationPayload(center models.FaultCenter, state models.AlertGroupState, items []groupedAlertDelivery, suppressedCount int, now time.Time) (string, map[string]interface{}) {
	firingCount, resolvedCount := 0, 0
	highestActiveSeverity, highestSeverity := "", ""
	lines := make([]string, 0, len(items)+3)
	alerts := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		if alertSeverityRank(item.alert.Severity) > alertSeverityRank(highestSeverity) {
			highestSeverity = item.alert.Severity
		}
		statusText := "告警中"
		if item.status == "resolved" {
			statusText = "已恢复"
			resolvedCount++
		} else {
			firingCount++
			if alertSeverityRank(item.alert.Severity) > alertSeverityRank(highestActiveSeverity) {
				highestActiveSeverity = item.alert.Severity
			}
		}
		lines = append(lines, fmt.Sprintf("- [%s][%s] %s · %s", statusText, item.alert.Severity, item.alert.Title, item.alert.Target))
		namespaces := alertLabelNamespaces(*item.alert)
		var snapshot struct {
			AlertName   string            `json:"alertname"`
			Datasource  string            `json:"datasource"`
			Value       string            `json:"value"`
			Condition   string            `json:"condition"`
			Description string            `json:"description"`
			Annotations map[string]string `json:"annotations"`
		}
		_ = json.Unmarshal([]byte(item.alert.LastResult), &snapshot)
		if strings.TrimSpace(snapshot.AlertName) == "" {
			snapshot.AlertName = item.alert.Title
		}
		if strings.TrimSpace(snapshot.Datasource) == "" {
			snapshot.Datasource = namespaces["system_labels"]["datasource"]
		}
		alerts = append(alerts, map[string]interface{}{
			"id": item.alert.ID, "status": item.status, "severity": item.alert.Severity,
			"title": item.alert.Title, "alertname": snapshot.AlertName, "target": item.alert.Target,
			"datasource": snapshot.Datasource, "value": snapshot.Value, "condition": snapshot.Condition,
			"description": snapshot.Description, "annotations": snapshot.Annotations, "fingerprint": item.alert.Fingerprint,
			"system_labels": namespaces["system_labels"], "labels": namespaces["labels"], "datasource_labels": namespaces["datasource_labels"],
			"external_labels": namespaces["external_labels"], "fired_at": item.alert.FiredAt, "resolved_at": item.alert.ResolvedAt, "recovered_at": item.alert.ResolvedAt,
			"first_trigger_time_format": formatNotificationTime(item.alert.FiredAt), "recover_time_format": formatNotificationTime(item.alert.ResolvedAt),
		})
	}
	titleColor := alertNotificationTitleColor(highestActiveSeverity, false)
	title := fmt.Sprintf("告警分组：%s（%d 条）", center.Name, len(items))
	if firingCount == 0 {
		titleColor = alertNotificationTitleColor(highestSeverity, true)
		title = fmt.Sprintf("告警分组已恢复：%s（%d 条）", center.Name, resolvedCount)
	}
	header := fmt.Sprintf("分组：%s\n告警中：%d 条，已恢复：%d 条", state.GroupKey, firingCount, resolvedCount)
	if suppressedCount > 0 {
		header += fmt.Sprintf("，已抑制/静默：%d 条", suppressedCount)
	}
	content := header + "\n\n" + strings.Join(lines, "\n")
	severity := highestActiveSeverity
	if firingCount == 0 {
		severity = highestSeverity
	}
	return title, map[string]interface{}{
		"title": title, "content": content, "title_color": titleColor, "source": "alert-group",
		"fault_center_id": center.ID, "fault_center_name": center.Name, "group_key": state.GroupKey,
		"firing_count": firingCount, "resolved_count": resolvedCount, "suppressed_count": suppressedCount,
		"severity": severity, "level": severity,
		"alerts": alerts, "fingerprint": state.GroupKey, "status": map[bool]string{true: "resolved", false: "firing"}[firingCount == 0], "evaluated_at": now,
	}
}

func finishAlertGroupDispatch(state models.AlertGroupState, notified bool, next *time.Time, now time.Time) {
	updates := map[string]interface{}{"pending": next != nil, "next_notification_at": next, "lease_until": nil}
	if notified {
		updates["last_notified_at"] = now
	}
	if next == nil {
		// A closed group must use group_wait again when a future incident with
		// the same labels starts.
		updates["closed_at"] = now
	}
	database.DB.Model(&models.AlertGroupState{}).Where("id = ?", state.ID).Updates(updates)
}

func alertGroupEventKey(centerID uint, groupKey string, channelID uint, now time.Time) string {
	sum := sha256.Sum256([]byte(groupKey))
	return fmt.Sprintf("alert-group:%d:%s:%d:%d", centerID, hex.EncodeToString(sum[:8]), channelID, now.Unix())
}

func timePointer(value time.Time) *time.Time { return &value }

func formatNotificationTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Format("2006-01-02 15:04:05")
}

func maxInt(first, second int) int {
	if first > second {
		return first
	}
	return second
}
