package handlers

import (
	"guanlan/internal/database"
	"guanlan/internal/models"
	"time"
)

func terminateTaskAlerts(source string, sourceID uint, reason string, now time.Time) {
	var alerts []models.AlertInstance
	database.DB.Where("source = ? AND source_id = ? AND status IN ?", source, sourceID, []string{alertStatusPending, alertStatusFiring, alertStatusRecovering}).Find(&alerts)
	for i := range alerts {
		alert := &alerts[i]
		previous := alert.Status
		notificationReason := "任务已停用或删除"
		if reason == "scheduled_task_retired" {
			notificationReason = "定时任务功能已下线，请使用脚本拨测"
		}
		database.DB.Model(alert).Updates(map[string]interface{}{"status": alertStatusTerminated, "resolved_at": now, "state_changed_at": now, "last_event_at": now, "notification_status": notificationStatusTerminated, "notification_reason": notificationReason, "next_notification_at": nil})
		alert.Status, alert.ResolvedAt = alertStatusTerminated, &now
		_ = recordAlertEvent(database.DB, *alert, "state_changed", previous, alertStatusTerminated, alert.Severity, alert.Severity, reason, "", alert.LastProbeStatus, now)
	}
}
