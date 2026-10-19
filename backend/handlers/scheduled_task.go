package handlers

import (
	"aiops-backend/database"
	"aiops-backend/models"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var schedulerOnce sync.Once

// ── CRUD ──

type saveScheduledTaskReq struct {
	Name             string `json:"name" binding:"required"`
	Type             string `json:"type" binding:"required"`
	TargetID         uint   `json:"target_id"`
	CronExpr         string `json:"cron_expr" binding:"required"`
	Description      string `json:"description"`
	ExternalLabels   string `json:"external_labels"`
	Severity         string `json:"severity"`
	FailureThreshold int    `json:"failure_threshold"`
	FaultCenterID    uint   `json:"fault_center_id"`
	Status           int    `json:"status"`
}

func ListScheduledTasks(c *gin.Context) {
	var tasks []models.ScheduledTask
	database.DB.Order("id ASC").Find(&tasks)
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "成功", "data": tasks})
}

func CreateScheduledTask(c *gin.Context) {
	var req saveScheduledTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}
	if err := normalizeScheduledTaskReq(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	if err := validateScheduledTaskReq(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	t := models.ScheduledTask{
		Name: strings.TrimSpace(req.Name), Type: strings.ToLower(strings.TrimSpace(req.Type)), TargetID: req.TargetID,
		CronExpr: strings.TrimSpace(req.CronExpr), Description: strings.TrimSpace(req.Description),
		ExternalLabels: req.ExternalLabels, Severity: req.Severity, FailureThreshold: req.FailureThreshold,
		FaultCenterID: req.FaultCenterID, Status: req.Status,
	}
	database.DB.Create(&t)
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "创建成功"})
}

func UpdateScheduledTask(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req saveScheduledTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}
	if err := normalizeScheduledTaskReq(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	if err := validateScheduledTaskReq(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	database.DB.Model(&models.ScheduledTask{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name": strings.TrimSpace(req.Name), "type": strings.ToLower(strings.TrimSpace(req.Type)),
		"target_id": req.TargetID, "cron_expr": strings.TrimSpace(req.CronExpr),
		"description": strings.TrimSpace(req.Description), "external_labels": req.ExternalLabels,
		"severity": req.Severity, "failure_threshold": req.FailureThreshold,
		"fault_center_id": req.FaultCenterID, "status": req.Status,
	})
	if req.Status == 0 {
		terminateTaskAlerts("scheduled-task", uint(id), "task_disabled", time.Now())
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "更新成功"})
}

func validateScheduledTaskReq(req saveScheduledTaskReq) error {
	targetType := strings.ToLower(strings.TrimSpace(req.Type))
	if targetType != "script" && targetType != "workflow" {
		return fmt.Errorf("定时任务只支持脚本或工作流")
	}
	if req.TargetID == 0 {
		return fmt.Errorf("请选择执行对象")
	}
	if req.Status != 0 && req.Status != 1 {
		return fmt.Errorf("任务状态不正确")
	}
	if len(strings.Fields(req.CronExpr)) != 5 {
		return fmt.Errorf("Cron 表达式格式不正确")
	}
	if req.FailureThreshold < 1 || req.FailureThreshold > 100 {
		return fmt.Errorf("连续失败阈值应为 1-100 次")
	}
	if req.Severity != "P0" && req.Severity != "P1" && req.Severity != "P2" {
		return fmt.Errorf("告警等级不正确")
	}
	if req.FaultCenterID == 0 {
		return fmt.Errorf("请选择故障中心")
	}
	var centerCount int64
	database.DB.Model(&models.FaultCenter{}).Where("id = ? AND status = 1", req.FaultCenterID).Count(&centerCount)
	if centerCount == 0 {
		return fmt.Errorf("故障中心不存在或已停用")
	}
	var count int64
	if targetType == "script" {
		database.DB.Model(&models.Script{}).Where("id = ? AND status = ?", req.TargetID, 1).Count(&count)
	} else {
		database.DB.Model(&models.Workflow{}).Where("id = ? AND status = ?", req.TargetID, 1).Count(&count)
	}
	if count == 0 {
		return fmt.Errorf("执行对象不存在或已停用")
	}
	return nil
}

func normalizeScheduledTaskReq(req *saveScheduledTaskReq) error {
	labels, err := normalizeInspectionLabels(req.ExternalLabels)
	if err != nil {
		return err
	}
	req.ExternalLabels = labels
	if req.FailureThreshold == 0 {
		req.FailureThreshold = 1
	}
	if strings.TrimSpace(req.Severity) == "" {
		req.Severity = "P2"
	} else {
		req.Severity = strings.ToUpper(strings.TrimSpace(req.Severity))
	}
	return nil
}

func DeleteScheduledTask(c *gin.Context) {
	id := c.Param("id")
	parsedID, _ := strconv.ParseUint(id, 10, 64)
	terminateTaskAlerts("scheduled-task", uint(parsedID), "task_deleted", time.Now())
	database.DB.Where("task_id = ?", id).Delete(&models.ScheduledTaskExecution{})
	database.DB.Delete(&models.ScheduledTask{}, id)
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "删除成功"})
}

// ── Execute ──

func ExecuteScheduledTask(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var task models.ScheduledTask
	if err := database.DB.First(&task, id).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "定时任务不存在"})
		return
	}

	record := runScheduledTask(&task)
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "执行完成: " + record.Status, "data": record})
}

func runScheduledTask(task *models.ScheduledTask) models.ScheduledTaskExecution {
	status, output, durationMs := executeScheduledTarget(*task)
	record := models.ScheduledTaskExecution{
		TaskID: task.ID, Status: status, Output: output, DurationMs: durationMs,
	}
	database.DB.Create(&record)
	now := time.Now()
	processScheduledTaskAlert(task, status, output, now)
	database.DB.Model(task).Updates(map[string]interface{}{"last_run_at": now, "last_status": status})
	task.LastRunAt, task.LastStatus = &now, status
	return record
}

func processScheduledTaskAlert(task *models.ScheduledTask, status, output string, now time.Time) {
	if task.FaultCenterID == 0 {
		return
	}
	success := status == "success"
	failures := task.ConsecutiveFailures
	if success {
		failures = 0
	} else {
		failures++
	}
	previous, current, alert, err := syncScheduledTaskAlert(*task, status, output, failures, now)
	if err != nil {
		log.Printf("failed to sync alert for scheduled task %d: %v", task.ID, err)
		return
	}
	active := current == alertStatusFiring || current == alertStatusRecovering
	database.DB.Model(task).Updates(map[string]interface{}{"consecutive_failures": failures, "alert_active": active})
	task.ConsecutiveFailures, task.AlertActive = failures, active
	if alert == nil {
		return
	}
	if current == alertStatusFiring {
		config := loadAlertRuleNotificationConfig(task.FaultCenterID)
		due := alert.LastNotifiedAt == nil || !now.Before(alert.LastNotifiedAt.Add(time.Duration(maxInt(config.RepeatSeconds, 60))*time.Second))
		if due {
			sendScheduledTaskAlertNotification(*task, alert, false, output, now)
		}
	}
	if current == alertStatusResolved && (previous == alertStatusFiring || previous == alertStatusRecovering) {
		sendScheduledTaskAlertNotification(*task, alert, true, output, now)
	}
}

func syncScheduledTaskAlert(task models.ScheduledTask, executionStatus, output string, failures int, now time.Time) (string, string, *models.AlertInstance, error) {
	var alert models.AlertInstance
	tx := database.DB.Where("source = ? AND source_id = ? AND status IN ?", "scheduled-task", task.ID, []string{alertStatusPending, alertStatusFiring, alertStatusRecovering}).Order("id DESC").Limit(1).Find(&alert)
	if tx.Error != nil {
		return "", "", nil, tx.Error
	}
	target := scheduledTaskTargetName(task)
	runtimeLabels, snapshot := buildTaskAlertSnapshot("scheduled-task", task.Name, task.Type, target, executionStatus, truncate(output, 4000), task.Description, task.ExternalLabels)
	previous := alert.Status
	success := executionStatus == "success"
	if tx.RowsAffected == 0 {
		if success {
			return "", "", nil, nil
		}
		state := nextInspectionAlertState("", false, failures, task.FailureThreshold, false)
		alert = models.AlertInstance{FaultCenterID: task.FaultCenterID, Source: "scheduled-task", SourceID: task.ID, Title: task.Name, Target: target, Severity: task.Severity, Status: state, Labels: runtimeLabels, FailureCount: failures, FailureThreshold: task.FailureThreshold, LastProbeStatus: executionStatus, LastResult: snapshot, StartedAt: now, StateChangedAt: now, LastEventAt: now}
		if state == alertStatusFiring {
			alert.FiredAt = &now
		}
		if err := database.DB.Create(&alert).Error; err != nil {
			return previous, state, nil, err
		}
		_ = recordAlertEvent(database.DB, alert, "created", "", state, "", alert.Severity, "scheduled_task_failed", output, executionStatus, now)
		applyActiveSilenceRule(&alert, now)
		return previous, state, &alert, nil
	}
	state := nextInspectionAlertState(previous, success, failures, task.FailureThreshold, true)
	updates := map[string]interface{}{"fault_center_id": task.FaultCenterID, "title": task.Name, "target": target, "severity": task.Severity, "status": state, "labels": runtimeLabels, "failure_count": failures, "failure_threshold": task.FailureThreshold, "last_probe_status": executionStatus, "last_result": snapshot, "last_event_at": now}
	applyAlertStateTransition(updates, previous, state, now)
	if err := database.DB.Model(&alert).Updates(updates).Error; err != nil {
		return previous, state, nil, err
	}
	alert.Status, alert.Labels, alert.LastResult, alert.LastEventAt = state, runtimeLabels, snapshot, now
	if state != previous {
		alert.StateChangedAt = now
		if state == alertStatusResolved {
			alert.ResolvedAt = &now
		}
		_ = recordAlertEvent(database.DB, alert, "state_changed", previous, state, alert.Severity, task.Severity, map[bool]string{true: "scheduled_task_recovered", false: "scheduled_task_state_changed"}[state == alertStatusResolved], output, executionStatus, now)
	}
	applyActiveSilenceRule(&alert, now)
	return previous, state, &alert, nil
}

func scheduledTaskTargetName(task models.ScheduledTask) string {
	if task.Type == "script" {
		var value models.Script
		if database.DB.Select("name").First(&value, task.TargetID).Error == nil {
			return value.Name
		}
	} else if task.Type == "workflow" {
		var value models.Workflow
		if database.DB.Select("name").First(&value, task.TargetID).Error == nil {
			return value.Name
		}
	}
	return fmt.Sprintf("%s #%d", task.Type, task.TargetID)
}

func sendScheduledTaskAlertNotification(task models.ScheduledTask, alert *models.AlertInstance, recovery bool, output string, now time.Time) {
	if alert == nil || alert.ID == 0 {
		return
	}
	config := loadAlertRuleNotificationConfig(task.FaultCenterID)
	if recovery {
		if queueGroupedRecoveryAlert(alert, config, now) {
			dispatchDueAlertGroups(now)
			return
		}
	} else if queueGroupedFiringAlert(alert, config, now) {
		dispatchDueAlertGroups(now)
		return
	}
	if recovery && (!config.RecoveryNotification || alert.LastNotifiedAt == nil) {
		status, reason := resolvedAlertNotificationStatus(alert.LastNotifiedAt), "在首次通知前恢复"
		if !config.RecoveryNotification && alert.LastNotifiedAt != nil {
			status, reason = notificationStatusRecoveryDisabled, "恢复通知未启用"
		}
		updateAlertNotificationState(alert, status, reason, alert.NotificationGroupKey, nil)
		return
	}
	if !recovery {
		if status, reason := alertNotificationSuppression(alert, now); status != "" {
			updateAlertNotificationState(alert, status, reason, "", nil)
			return
		}
	}
	channelIDs := notificationRouteChannelIDs(*alert, config.Routes, config.ChannelIDs)
	if len(channelIDs) == 0 {
		updateAlertNotificationState(alert, notificationStatusNoReceiver, "没有匹配到可用通知对象", "", nil)
		return
	}
	status, eventType, titleColor := "firing", "scheduled.task.firing", alertNotificationTitleColor(task.Severity, false)
	title := "定时任务失败：" + task.Name
	content := fmt.Sprintf("任务：%s\n执行对象：%s\n连续失败：%d/%d\n输出：%s", task.Name, alert.Target, alert.FailureCount, alert.FailureThreshold, truncate(output, 800))
	if recovery {
		status, eventType, titleColor = "resolved", "scheduled.task.recovery", alertNotificationTitleColor(task.Severity, true)
		title = "定时任务已恢复：" + task.Name
		content = fmt.Sprintf("任务：%s\n执行对象：%s\n本次执行成功，故障已恢复。", task.Name, alert.Target)
	}
	payload := map[string]interface{}{
		"title": title, "content": content, "title_color": titleColor, "source": "scheduled-task",
		"fault_center_id": task.FaultCenterID, "task_id": task.ID, "task_name": task.Name,
		"task_type": task.Type, "target_id": task.TargetID, "target": alert.Target,
		"status": status, "severity": task.Severity, "level": task.Severity, "result": output, "evaluated_at": now,
	}
	namespaces := alertLabelNamespaces(*alert)
	addNotificationLabelNamespace(payload, "system_labels", namespaces["system_labels"])
	addNotificationLabelNamespace(payload, "labels", namespaces["labels"])
	addNotificationLabelNamespace(payload, "datasource_labels", namespaces["datasource_labels"])
	addNotificationLabelNamespace(payload, "external_labels", namespaces["external_labels"])
	sent, attempted := false, false
	for _, channelID := range channelIDs {
		var channel models.NotificationChannel
		if database.DB.Where("id = ? AND status = 1", channelID).First(&channel).Error != nil {
			continue
		}
		attempted = true
		eventKey := fmt.Sprintf("scheduled-task:%d:%s:%d", task.ID, status, now.Unix())
		if sendNotificationToChannel(channel, eventType, eventKey, title, payload) == nil {
			sent = true
		}
	}
	if sent {
		notificationStatus := notificationStatusSent
		if recovery {
			notificationStatus = notificationStatusRecoverySent
		}
		updateAlertNotificationState(alert, notificationStatus, "通知已发送", "", nil)
		if !recovery {
			database.DB.Model(alert).Update("last_notified_at", now)
			alert.LastNotifiedAt = &now
		}
		_ = recordAlertEvent(database.DB, *alert, "notification_sent", alert.Status, alert.Status, alert.Severity, alert.Severity, status, output, alert.LastProbeStatus, now)
	} else if attempted {
		updateAlertNotificationState(alert, notificationStatusSendFailed, "通知发送失败", "", nil)
	}
}

func terminateTaskAlerts(source string, sourceID uint, reason string, now time.Time) {
	var alerts []models.AlertInstance
	database.DB.Where("source = ? AND source_id = ? AND status IN ?", source, sourceID, []string{alertStatusPending, alertStatusFiring, alertStatusRecovering}).Find(&alerts)
	for i := range alerts {
		alert := &alerts[i]
		previous := alert.Status
		database.DB.Model(alert).Updates(map[string]interface{}{"status": alertStatusTerminated, "resolved_at": now, "state_changed_at": now, "last_event_at": now, "notification_status": notificationStatusTerminated, "notification_reason": "任务已停用或删除", "next_notification_at": nil})
		alert.Status, alert.ResolvedAt = alertStatusTerminated, &now
		_ = recordAlertEvent(database.DB, *alert, "state_changed", previous, alertStatusTerminated, alert.Severity, alert.Severity, reason, "", alert.LastProbeStatus, now)
	}
}

func executeScheduledTarget(task models.ScheduledTask) (string, string, int64) {
	switch task.Type {
	case "script":
		var script models.Script
		if task.TargetID == 0 || database.DB.First(&script, task.TargetID).Error != nil {
			return "error", "脚本不存在", 0
		}
		if script.Status != 1 {
			return "error", "脚本已停用", 0
		}
		return runScript(script)
	case "workflow":
		var workflow models.Workflow
		if task.TargetID == 0 || database.DB.First(&workflow, task.TargetID).Error != nil {
			return "error", "工作流不存在", 0
		}
		if workflow.Status != 1 {
			return "error", "工作流已停用", 0
		}
		record := runWorkflow(&workflow)
		return record.Status, record.Result, record.DurationMs
	default:
		return "error", "不支持的定时任务类型", 0
	}
}

// ── Executions ──

func ListTaskExecutions(c *gin.Context) {
	taskID := c.Query("task_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}

	var total int64
	var records []models.ScheduledTaskExecution
	tx := database.DB.Model(&models.ScheduledTaskExecution{})
	if taskID != "" {
		tx = tx.Where("task_id = ?", taskID)
	}
	tx.Count(&total)
	tx.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&records)

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "成功", "data": gin.H{
		"list": records, "total": total, "page": page, "page_size": pageSize,
	}})
}

// ── Scheduler ──

func StartScheduler() {
	schedulerOnce.Do(func() {
		go schedulerLoop()
	})
}

func schedulerLoop() {
	taskTicker := time.NewTicker(30 * time.Second)
	alertRuleTicker := time.NewTicker(5 * time.Second)
	notificationTicker := time.NewTicker(time.Second)
	defer taskTicker.Stop()
	defer alertRuleTicker.Stop()
	defer notificationTicker.Stop()
	runDueAlertRuleBindings(time.Now())
	for {
		select {
		case now := <-notificationTicker.C:
			dispatchDueAlertGroups(now)
			dispatchDueAlertActions(now)
		case <-taskTicker.C:
			runPendingTasks()
		case now := <-alertRuleTicker.C:
			runDueAlertRuleBindings(now)
		}
	}
}

func runPendingTasks() {
	runDueInspectionItems(time.Now())

	var tasks []models.ScheduledTask
	database.DB.Where("status = 1").Find(&tasks)
	now := time.Now()

	for _, task := range tasks {
		if matchCron(task.CronExpr, now) {
			// 防重复：1分钟内刚执行过的跳过
			if task.LastRunAt != nil && now.Sub(*task.LastRunAt) < 50*time.Second {
				continue
			}
			runScheduledTask(&task)
		}
	}
}

func runDueInspectionItems(now time.Time) {
	var items []models.InspectionItem
	database.DB.Where("status = ? AND type IN ?", 1, []string{"http", "icmp", "tcp", "certificate_expiry"}).Find(&items)
	for index := range items {
		item := &items[index]
		if !inspectionTaskDue(*item, now) {
			continue
		}
		executeAndRecordInspection(item)
	}
}

func inspectionTaskDue(item models.InspectionItem, now time.Time) bool {
	if item.Status != 1 {
		return false
	}
	interval := item.IntervalSeconds
	if interval < 30 {
		interval = 300
	}
	return item.LastRunAt == nil || now.Sub(*item.LastRunAt) >= time.Duration(interval)*time.Second
}

// ── 简易 Cron 匹配器 ──

func matchCron(expr string, t time.Time) bool {
	parts := strings.Fields(expr)
	if len(parts) != 5 {
		return false
	}
	return matchField(parts[0], t.Minute()) &&
		matchField(parts[1], t.Hour()) &&
		matchField(parts[2], t.Day()) &&
		matchField(parts[3], int(t.Month())) &&
		matchField(parts[4], int(t.Weekday()))
}

func matchField(field string, val int) bool {
	if field == "*" {
		return true
	}

	// */N 步进
	if strings.HasPrefix(field, "*/") {
		step, _ := strconv.Atoi(field[2:])
		return step > 0 && val%step == 0
	}

	// N-M 范围
	if parts := strings.Split(field, "-"); len(parts) == 2 {
		start, _ := strconv.Atoi(parts[0])
		end, _ := strconv.Atoi(parts[1])
		return val >= start && val <= end
	}

	// N,M 枚举
	if strings.Contains(field, ",") {
		for _, s := range strings.Split(field, ",") {
			if n, err := strconv.Atoi(s); err == nil && n == val {
				return true
			}
		}
		return false
	}

	// 精确值
	n, err := strconv.Atoi(field)
	return err == nil && n == val
}

// NextRun 计算下次执行时间（简化版）
func NextRunTime(c *gin.Context) {
	expr := c.Query("expr")
	if expr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "缺少cron表达式"})
		return
	}
	next := nextCronTime(expr, time.Now())
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "成功", "data": next.Format(time.RFC3339)})
}

func nextCronTime(expr string, from time.Time) time.Time {
	for i := 0; i < 1440; i++ {
		from = from.Add(1 * time.Minute).Truncate(time.Minute)
		if matchCron(expr, from) {
			return from
		}
	}
	return from
}

func FormatCronHelp(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "成功", "data": map[string]string{
		"*/1 * * * *":  "每1分钟",
		"*/5 * * * *":  "每5分钟",
		"*/30 * * * *": "每30分钟",
		"0 * * * *":    "每小时整点",
		"0 */2 * * *":  "每2小时",
		"0 0 * * *":    "每天0点",
		"0 9 * * 1-5":  "工作日9点",
		"0 0 1 * *":    "每月1号0点",
	},
	})
}

// For test connection - return mock success
func init() { fmt.Println() }
