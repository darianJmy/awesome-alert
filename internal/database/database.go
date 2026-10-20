package database

import (
	"encoding/json"
	"fmt"
	"guanlan/internal/config"
	"guanlan/internal/models"
	"guanlan/internal/utils"
	"guanlan/pkg/logger"
	"log/slog"
	"net/url"
	"os"

	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	sqldb "guanlan/pkg/database"
)

var DB *gorm.DB

const CurrentSchemaVersion = "2026102001"

// Init is the legacy startup entry point. New callers should use New so errors
// and connection cleanup can be managed by the application.
func Init() {
	if _, _, err := New(config.AppConfig); err != nil {
		logger.Error("initialize database failed", "error", err)
		os.Exit(1)
	}
}

// New opens SQLite and atomically initializes a fresh database on first startup.
// DB remains available to legacy handlers until they accept injected dependencies.
func New(cfg config.Config) (*gorm.DB, func(), error) {
	db, err := openDatabaseDSN(cfg.DBDSN)
	if err != nil {
		return nil, nil, fmt.Errorf("connect database: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, fmt.Errorf("access database pool: %w", err)
	}
	cleanup := func() { _ = sqlDB.Close() }
	if err := prepareDatabase(db); err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("initialize database: %w", err)
	}
	DB = db
	logger.Info("database ready", "schema_version", CurrentSchemaVersion)
	return db, cleanup, nil
}

func prepareDatabase(db *gorm.DB) error {
	return sqldb.Initialize(db, CurrentSchemaVersion, createCurrentSchema, func(tx *gorm.DB) error {
		if err := tx.Create(&models.InspectionRuleGroup{ID: 1, Name: "默认规则组", Description: "默认拨测任务规则组"}).Error; err != nil {
			return err
		}
		return seedInitialData(tx)
	})
}

func openDatabaseDSN(dsn string) (*gorm.DB, error) {
	return sqldb.OpenSQLite(dsn, slog.Default())
}

func createCurrentSchema(db *gorm.DB) error {
	if err := sqldb.AutoMigrate(db,
		&models.User{},
		&models.Role{},
		&models.Permission{},
		&models.Menu{},
		&models.Datasource{},
		&models.DatasourceGroup{},
		&models.DatasourceGroupMember{},
		&models.Dashboard{},
		&models.SLSDashboardProject{},
		&models.CloudMonitorWorkspace{},
		&models.InspectionRuleGroup{},
		&models.InspectionItem{},
		&models.InspectionRecord{},
		&models.InspectionNodeState{},
		&models.InspectionTrendBucket{},
		&models.Script{},
		&models.ScriptExecution{},
		&models.Workflow{},
		&models.WorkflowExecution{},
		&models.AutomationTask{},
		&models.AutomationExecution{},
		&models.GitCredential{},
		&models.GitRepository{},
		&models.SystemSetting{},
		&models.AuthSetting{},
		&models.APIAccessToken{},
		&models.MCPClient{},
		&models.Tunnel{},
		&models.ChatConversation{},
		&models.ChatMessage{},
		&models.ChatMemory{},
		&models.ScheduledTask{},
		&models.ScheduledTaskExecution{},
		&models.Certificate{},
		&models.CertificateVersion{},
		&models.CertificateDomain{},
		&models.AgentConnection{},
		&models.CertificateDeployScript{},
		&models.CertificateExecutor{},
		&models.CertificateDeploymentRecord{},
		&models.CertificateRotationRecord{},
		&models.NotificationChannel{},
		&models.NotificationTemplate{},
		&models.NotificationRecord{},
		&models.PromQueryHistory{},
		&models.FaultCenter{},
		&models.AlertInstance{},
		&models.AlertEvent{},
		&models.AlertSilenceRule{},
		&models.AlertInhibitionRule{},
		&models.AlertNotificationConfig{},
		&models.AlertNotificationRoute{},
		&models.AlertGroupingRule{},
		&models.AlertGroupState{},
		&models.AlertActionRule{},
		&models.AlertActionExecution{},
		&models.AlertRuleGroup{},
		&models.AlertRule{},
		&models.AlertRuleDatasource{},
		&models.AlertRuleDatasourceGroup{},
		&models.AlertRuleCondition{},
		&models.AlertRuleBinding{},
		&models.AlertEvaluationRecord{},
		&models.AuditLog{},
	); err != nil {
		return fmt.Errorf("create database schema: %w", err)
	}
	return nil
}

func migrateCloudMonitorWorkspaceNames() error {
	if !DB.Migrator().HasTable(&models.CloudMonitorWorkspace{}) || !DB.Migrator().HasColumn(&models.CloudMonitorWorkspace{}, "name") {
		return nil
	}
	return DB.Model(&models.CloudMonitorWorkspace{}).
		Where("name = '' OR name IS NULL").
		Update("name", gorm.Expr("workspace")).Error
}

func migrateUniqueScriptNames() error {
	if !DB.Migrator().HasTable(&models.Script{}) {
		return nil
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		var scripts []models.Script
		if err := tx.Order("id ASC").Find(&scripts).Error; err != nil {
			return err
		}
		for i := range scripts {
			var count int64
			if err := tx.Model(&models.Script{}).Where("name = ? AND id < ?", scripts[i].Name, scripts[i].ID).Count(&count).Error; err != nil {
				return err
			}
			if count == 0 {
				continue
			}
			for suffix := 2; ; suffix++ {
				candidate := suffixedScriptName(scripts[i].Name, suffix)
				if err := tx.Model(&models.Script{}).Where("name = ?", candidate).Count(&count).Error; err != nil {
					return err
				}
				if count == 0 {
					if err := tx.Model(&scripts[i]).Update("name", candidate).Error; err != nil {
						return err
					}
					break
				}
			}
		}
		return nil
	})
}

func suffixedScriptName(name string, number int) string {
	suffix := " (" + strconv.Itoa(number) + ")"
	base := []rune(strings.TrimSpace(name))
	limit := 100 - len([]rune(suffix))
	if len(base) > limit {
		base = base[:limit]
	}
	return string(base) + suffix
}

func migrateGitCredentials() error {
	return DB.Transaction(func(tx *gorm.DB) error {
		var repositories []models.GitRepository
		if err := tx.Where("credential_id = 0 OR credential_id IS NULL").Order("id ASC").Find(&repositories).Error; err != nil {
			return err
		}
		for i := range repositories {
			repository := repositories[i]
			name := legacyGitCredentialName(repository.Name, 0)
			for suffix := 2; ; suffix++ {
				var count int64
				if err := tx.Model(&models.GitCredential{}).Where("name = ?", name).Count(&count).Error; err != nil {
					return err
				}
				if count == 0 {
					break
				}
				name = legacyGitCredentialName(repository.Name, suffix)
			}
			authType := strings.TrimSpace(repository.CredentialType)
			if authType == "" {
				authType = "none"
			}
			credential := models.GitCredential{
				Name: name, Provider: legacyGitProvider(repository.RepositoryURL), BaseURL: legacyGitBaseURL(repository.RepositoryURL),
				AuthType: authType, Username: repository.CredentialUser, SecretCipher: repository.CredentialCipher,
				Description: "由 Git 仓库「" + repository.Name + "」的原凭据自动迁移", Status: 1,
			}
			if err := tx.Create(&credential).Error; err != nil {
				return err
			}
			if err := tx.Model(&repository).Update("credential_id", credential.ID).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func legacyGitCredentialName(repositoryName string, suffixNumber int) string {
	suffix := " 认证"
	if suffixNumber > 0 {
		suffix += " (" + strconv.Itoa(suffixNumber) + ")"
	}
	base := []rune(strings.TrimSpace(repositoryName))
	limit := 128 - len([]rune(suffix))
	if len(base) > limit {
		base = base[:limit]
	}
	return string(base) + suffix
}

func legacyGitBaseURL(repositoryURL string) string {
	value := strings.TrimSpace(repositoryURL)
	if strings.HasPrefix(value, "git@") {
		host := strings.SplitN(strings.TrimPrefix(value, "git@"), ":", 2)[0]
		return "https://" + host
	}
	if parsed, err := url.Parse(value); err == nil && parsed.Host != "" {
		return parsed.Scheme + "://" + parsed.Host
	}
	return value
}

func legacyGitProvider(repositoryURL string) string {
	host := strings.ToLower(legacyGitBaseURL(repositoryURL))
	switch {
	case strings.Contains(host, "github"):
		return "github"
	case strings.Contains(host, "gitlab"):
		return "gitlab"
	case strings.Contains(host, "gitee"):
		return "gitee"
	default:
		return "other"
	}
}

func migrateCertificateAutomation() error {
	return DB.Transaction(func(tx *gorm.DB) error {
		var legacyScripts []models.CertificateDeployScript
		if err := tx.Find(&legacyScripts).Error; err != nil {
			return err
		}
		for i := range legacyScripts {
			legacy := legacyScripts[i]
			var script models.Script
			result := tx.Where("name = ? AND type = ? AND content = ?", legacy.Name, legacy.Type, legacy.Content).First(&script)
			if result.Error != nil {
				script = models.Script{Name: legacy.Name, Type: legacy.Type, Content: legacy.Content, SourceType: "inline", Description: legacy.Description, Status: legacy.Status}
				if script.Status != 1 {
					script.Status = 0
				}
				if err := tx.Create(&script).Error; err != nil {
					return err
				}
			}
			if err := tx.Model(&models.AutomationTask{}).
				Where("runtime_profile = ? AND source_type = ? AND script_id = ?", "certificate_deploy", "script", legacy.ID).
				Update("script_id", script.ID).Error; err != nil {
				return err
			}
		}
		if len(legacyScripts) > 0 {
			if err := tx.Delete(&models.CertificateDeployScript{}, legacyScripts).Error; err != nil {
				return err
			}
		}
		var domainCertificateIDs []uint
		if err := tx.Model(&models.CertificateDomain{}).Distinct("certificate_id").Where("certificate_id > 0").Pluck("certificate_id", &domainCertificateIDs).Error; err != nil {
			return err
		}
		if len(domainCertificateIDs) == 1 {
			if err := tx.Model(&models.AutomationTask{}).
				Where("runtime_profile = ? AND certificate_id = 0", "certificate_deploy").
				Update("certificate_id", domainCertificateIDs[0]).Error; err != nil {
				return err
			}
		}
		if err := tx.Model(&models.AutomationTask{}).Where("runtime_profile <> ?", "certificate_deploy").Update("certificate_id", 0).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.AutomationTask{}).Where("runtime_profile = ? AND certificate_id = 0", "certificate_deploy").Update("status", 0).Error; err != nil {
			return err
		}
		return tx.Model(&models.CertificateDomain{}).Where("1 = 1").Updates(map[string]interface{}{
			"certificate_id": 0, "last_observed_version_id": 0,
			"executor_id": 0, "agent_id": 0, "deploy_script_id": 0, "notification_channel_id": 0,
			"deploy_params": "", "deploy_cert_path": "", "deploy_key_path": "", "deploy_reload_command": "",
		}).Error
	})
}

func migrateInspectionTrendHistory() error {
	var trendCount int64
	if err := DB.Model(&models.InspectionTrendBucket{}).Count(&trendCount).Error; err != nil {
		return err
	}
	if trendCount > 0 {
		if err := DB.Where("bucket_at < ?", time.Now().Add(-30*24*time.Hour)).Delete(&models.InspectionTrendBucket{}).Error; err != nil {
			return err
		}
		return pruneInspectionRecords(12)
	}
	rows, err := DB.Model(&models.InspectionRecord{}).Where("created_at >= ?", time.Now().Add(-30*24*time.Hour)).Rows()
	if err != nil {
		return err
	}
	defer rows.Close()
	type bucketKey struct {
		ItemID   uint
		NodeType string
		NodeID   uint
		BucketAt int64
	}
	bucketMap := make(map[bucketKey]*models.InspectionTrendBucket)
	for rows.Next() {
		var record models.InspectionRecord
		if err := DB.ScanRows(rows, &record); err != nil {
			return err
		}
		bucketAt := record.CreatedAt.Truncate(5 * time.Minute)
		key := bucketKey{ItemID: record.ItemID, NodeType: record.NodeType, NodeID: record.NodeID, BucketAt: bucketAt.Unix()}
		bucket := bucketMap[key]
		if bucket == nil {
			bucket = &models.InspectionTrendBucket{ItemID: record.ItemID, NodeType: record.NodeType, NodeID: record.NodeID, NodeName: record.NodeName, BucketAt: bucketAt, DurationMinMs: record.DurationMs, DurationMaxMs: record.DurationMs}
			bucketMap[key] = bucket
		}
		bucket.TotalCount++
		if record.Status == "success" {
			bucket.SuccessCount++
		} else {
			bucket.FailureCount++
		}
		bucket.DurationSumMs += record.DurationMs
		if record.DurationMs < bucket.DurationMinMs {
			bucket.DurationMinMs = record.DurationMs
		}
		if record.DurationMs > bucket.DurationMaxMs {
			bucket.DurationMaxMs = record.DurationMs
		}
		switch inspectionRecordStatusGroup(record.Result) {
		case "2xx":
			bucket.Status2xxCount++
		case "3xx":
			bucket.Status3xxCount++
		case "4xx":
			bucket.Status4xxCount++
		case "5xx":
			bucket.Status5xxCount++
		default:
			bucket.StatusOther++
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}
	if len(bucketMap) == 0 {
		return pruneInspectionRecords(12)
	}
	buckets := make([]models.InspectionTrendBucket, 0, len(bucketMap))
	for _, bucket := range bucketMap {
		buckets = append(buckets, *bucket)
	}
	if err := DB.CreateInBatches(&buckets, 200).Error; err != nil {
		return err
	}
	return pruneInspectionRecords(12)
}

func pruneInspectionRecords(limit int) error {
	type recordGroup struct {
		ItemID   uint
		NodeType string
		NodeID   uint
	}
	var groups []recordGroup
	if err := DB.Model(&models.InspectionRecord{}).Distinct("item_id", "node_type", "node_id").Find(&groups).Error; err != nil {
		return err
	}
	for _, group := range groups {
		var staleIDs []uint
		if err := DB.Model(&models.InspectionRecord{}).Where("item_id = ? AND node_type = ? AND node_id = ?", group.ItemID, group.NodeType, group.NodeID).Order("id DESC").Offset(limit).Limit(1000000).Pluck("id", &staleIDs).Error; err != nil {
			return err
		}
		if len(staleIDs) > 0 {
			if err := DB.Where("id IN ?", staleIDs).Delete(&models.InspectionRecord{}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func inspectionRecordStatusGroup(result string) string {
	var payload struct {
		StatusCode int `json:"status_code"`
		Response   struct {
			StatusCode int `json:"status_code"`
		} `json:"response"`
	}
	if json.Unmarshal([]byte(result), &payload) != nil {
		return "other"
	}
	statusCode := payload.StatusCode
	if statusCode == 0 {
		statusCode = payload.Response.StatusCode
	}
	if statusCode >= 200 && statusCode < 600 {
		return string(rune('0'+statusCode/100)) + "xx"
	}
	return "other"
}

// migrateAlertGroupingRules moves the former singleton configuration from
// fault_centers into the rule table. The legacy columns are removed after the
// data has been copied so the runtime has one authoritative model.
func migrateAlertGroupingRules() error {
	if !DB.Migrator().HasColumn("fault_centers", "group_name") {
		return nil
	}
	type legacyGrouping struct {
		ID                   uint
		GroupEnabled         bool
		GroupName            string
		GroupBy              string
		GroupWaitSeconds     int
		GroupIntervalSeconds int
	}
	var rows []legacyGrouping
	if err := DB.Table("fault_centers").Select("id", "group_enabled", "group_name", "group_by", "group_wait_seconds", "group_interval_seconds").Scan(&rows).Error; err != nil {
		return err
	}
	for _, row := range rows {
		groupBy := strings.TrimSpace(row.GroupBy)
		if strings.TrimSpace(row.GroupName) == "" || (!row.GroupEnabled && (groupBy == "" || groupBy == "[]")) {
			continue
		}
		var count int64
		if err := DB.Model(&models.AlertGroupingRule{}).Where("fault_center_id = ?", row.ID).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		status := 0
		if row.GroupEnabled {
			status = 1
		}
		waitSeconds, intervalSeconds := row.GroupWaitSeconds, row.GroupIntervalSeconds
		if waitSeconds < 0 {
			waitSeconds = 30
		}
		if intervalSeconds < 60 {
			intervalSeconds = 300
		}
		rule := models.AlertGroupingRule{
			FaultCenterID: row.ID, Name: strings.TrimSpace(row.GroupName), MatchersJSON: "[]", GroupByJSON: groupBy,
			WaitSeconds: waitSeconds, IntervalSeconds: intervalSeconds, Priority: 100, Status: status,
		}
		if rule.GroupByJSON == "" {
			rule.GroupByJSON = "[]"
		}
		if err := DB.Create(&rule).Error; err != nil {
			return err
		}
	}
	for _, column := range []string{"group_enabled", "group_name", "group_by", "group_wait_seconds", "group_interval_seconds"} {
		if DB.Migrator().HasColumn("fault_centers", column) {
			if err := DB.Migrator().DropColumn("fault_centers", column); err != nil {
				return err
			}
		}
	}
	return nil
}

func migrateSLSDashboardProjects() error {
	var datasources []models.Datasource
	if err := DB.Where("type = ?", "sls").Order("id ASC").Find(&datasources).Error; err != nil {
		return err
	}
	for _, datasource := range datasources {
		var configData struct {
			Project string `json:"project"`
			Region  string `json:"region"`
		}
		var authData struct {
			AccessKeyID     string `json:"access_key_id"`
			AccessKeySecret string `json:"access_key_secret"`
		}
		if json.Unmarshal([]byte(datasource.Config), &configData) != nil || json.Unmarshal([]byte(datasource.AuthConfig), &authData) != nil {
			continue
		}
		if strings.TrimSpace(configData.Project) == "" || strings.TrimSpace(configData.Region) == "" || strings.TrimSpace(authData.AccessKeyID) == "" || strings.TrimSpace(authData.AccessKeySecret) == "" {
			continue
		}
		var count int64
		if err := DB.Model(&models.SLSDashboardProject{}).Where("legacy_datasource_id = ?", datasource.ID).Count(&count).Error; err != nil || count > 0 {
			if err != nil {
				return err
			}
			continue
		}
		accessKeyIDCipher, err := utils.EncryptSecret(authData.AccessKeyID)
		if err != nil {
			return err
		}
		accessKeySecretCipher, err := utils.EncryptSecret(authData.AccessKeySecret)
		if err != nil {
			return err
		}
		project := models.SLSDashboardProject{
			Name:                  datasource.Name,
			Project:               strings.TrimSpace(configData.Project),
			Region:                strings.TrimSpace(configData.Region),
			AccessKeyIDCipher:     accessKeyIDCipher,
			AccessKeySecretCipher: accessKeySecretCipher,
			Description:           datasource.Remark,
			LegacyDatasourceID:    &datasource.ID,
		}
		if err := DB.Create(&project).Error; err != nil {
			return err
		}
	}
	return nil
}

func migrateInspectionRecoveryThreshold() error {
	if !DB.Migrator().HasColumn("inspection_items", "recovery_wait_seconds") {
		return nil
	}
	var rows []struct {
		ID                  uint
		IntervalSeconds     int
		RecoveryWaitSeconds int
	}
	if err := DB.Table("inspection_items").Select("id", "interval_seconds", "recovery_wait_seconds").Scan(&rows).Error; err != nil {
		return err
	}
	for _, row := range rows {
		interval := row.IntervalSeconds
		if interval < 1 {
			interval = 300
		}
		threshold := (row.RecoveryWaitSeconds + interval - 1) / interval
		if threshold < 1 {
			threshold = 1
		} else if threshold > 100 {
			threshold = 100
		}
		if err := DB.Table("inspection_items").Where("id = ?", row.ID).Update("recovery_threshold", threshold).Error; err != nil {
			return err
		}
	}
	for _, column := range []string{"recovery_wait_seconds", "recovery_started_at"} {
		if DB.Migrator().HasColumn("inspection_items", column) {
			if err := DB.Migrator().DropColumn("inspection_items", column); err != nil {
				return err
			}
		}
	}
	return nil
}

func removeLegacyInspectionNotificationConfig() error {
	for _, column := range []string{"repeat_interval_seconds", "recovery_notification", "notification_channel_id", "last_alert_at"} {
		if DB.Migrator().HasColumn("inspection_items", column) {
			if err := DB.Migrator().DropColumn("inspection_items", column); err != nil {
				return err
			}
		}
	}
	return nil
}

func removeLegacyScriptProbes() error {
	return DB.Transaction(func(tx *gorm.DB) error {
		var taskIDs []uint
		if tx.Migrator().HasTable("inspection_items") {
			if err := tx.Table("inspection_items").Where("type = ?", "script").Pluck("id", &taskIDs).Error; err != nil {
				return err
			}
		}
		if len(taskIDs) > 0 {
			var alertIDs []uint
			if err := tx.Model(&models.AlertInstance{}).Where("source = ? AND source_id IN ?", "probe-task", taskIDs).Pluck("id", &alertIDs).Error; err != nil {
				return err
			}
			if len(alertIDs) > 0 {
				if err := tx.Where("alert_id IN ?", alertIDs).Delete(&models.AlertEvent{}).Error; err != nil {
					return err
				}
				if err := tx.Where("id IN ?", alertIDs).Delete(&models.AlertInstance{}).Error; err != nil {
					return err
				}
			}
			if err := tx.Where("item_id IN ?", taskIDs).Delete(&models.InspectionRecord{}).Error; err != nil {
				return err
			}
			if err := tx.Where("id IN ?", taskIDs).Delete(&models.InspectionItem{}).Error; err != nil {
				return err
			}
		}
		if tx.Migrator().HasColumn(&models.Script{}, "purpose") {
			return tx.Migrator().DropColumn(&models.Script{}, "purpose")
		}
		return nil
	})
}

func removeLegacyWorkspaces() error {
	if !DB.Migrator().HasTable("workspaces") {
		return nil
	}
	for _, table := range []string{"chat_conversations", "chat_memories"} {
		if DB.Migrator().HasTable(table) && DB.Migrator().HasColumn(table, "workspace_id") {
			if err := DB.Table(table).Where("workspace_id <> 0").Update("workspace_id", 0).Error; err != nil {
				return err
			}
		}
	}
	if DB.Migrator().HasTable("workspace_datasources") {
		if err := DB.Migrator().DropTable("workspace_datasources"); err != nil {
			return err
		}
	}
	return DB.Migrator().DropTable("workspaces")
}

func migrateAlertSilenceLifecycle() error {
	if DB.Migrator().HasColumn("alert_silence_rules", "match_all") {
		if err := DB.Where("match_all = ?", true).Delete(&models.AlertSilenceRule{}).Error; err != nil {
			return err
		}
		if err := DB.Migrator().DropColumn(&models.AlertSilenceRule{}, "match_all"); err != nil {
			return err
		}
	}
	columns, err := DB.Migrator().ColumnTypes(&models.AlertSilenceRule{})
	if err != nil {
		return err
	}
	fields := map[string]string{"start_at": "StartAt", "end_at": "EndAt"}
	for _, column := range columns {
		if field, exists := fields[column.Name()]; exists {
			if nullable, known := column.Nullable(); known && !nullable {
				if err := DB.Migrator().AlterColumn(&models.AlertSilenceRule{}, field); err != nil {
					return err
				}
			}
		}
	}
	var rows []models.AlertSilenceRule
	if err := DB.Where("silence_type = ?", "daily").Find(&rows).Error; err != nil {
		return err
	}
	for _, row := range rows {
		startTime, endTime := "00:00", "23:59"
		if row.StartAt != nil {
			startTime = row.StartAt.Format("15:04")
		}
		if row.EndAt != nil {
			endTime = row.EndAt.Format("15:04")
		}
		updates := map[string]interface{}{
			"silence_type": "weekly",
			"weekdays":     "[1,2,3,4,5,6,7]",
			"start_time":   startTime,
			"end_time":     endTime,
		}
		if err := DB.Model(&models.AlertSilenceRule{}).Where("id = ?", row.ID).Updates(updates).Error; err != nil {
			return err
		}
	}
	return DB.Model(&models.AlertSilenceRule{}).Where("silence_type = ?", "weekly").Updates(map[string]interface{}{"start_at": nil, "end_at": nil}).Error
}

func migrateAlertLabelMatchers() error {
	if DB.Migrator().HasColumn("alert_silence_rules", "labels") {
		var rows []struct {
			ID       uint
			Labels   string
			Matchers string
		}
		if err := DB.Table("alert_silence_rules").Select("id", "labels", "matchers").Scan(&rows).Error; err != nil {
			return err
		}
		for _, row := range rows {
			if strings.TrimSpace(row.Matchers) != "" {
				continue
			}
			labels := map[string]string{}
			if json.Unmarshal([]byte(row.Labels), &labels) != nil {
				continue
			}
			matchers := make([]models.AlertLabelMatcher, 0, len(labels))
			for key, value := range labels {
				matchers = append(matchers, models.AlertLabelMatcher{Namespace: "labels", Key: key, Value: value})
			}
			encoded, _ := json.Marshal(matchers)
			if err := DB.Table("alert_silence_rules").Where("id = ?", row.ID).Update("matchers", string(encoded)).Error; err != nil {
				return err
			}
		}
		if err := DB.Migrator().DropColumn("alert_silence_rules", "labels"); err != nil {
			return err
		}
	}

	if DB.Migrator().HasColumn("alert_inhibition_rules", "source_match") {
		var rows []struct {
			ID             uint
			SourceMatch    string
			TargetMatch    string
			EqualLabels    string
			SourceMatchers string
			TargetMatchers string
		}
		if err := DB.Table("alert_inhibition_rules").Select("id", "source_match", "target_match", "equal_labels", "source_matchers", "target_matchers").Scan(&rows).Error; err != nil {
			return err
		}
		for _, row := range rows {
			if strings.TrimSpace(row.SourceMatchers) != "" || strings.TrimSpace(row.TargetMatchers) != "" {
				continue
			}
			source, sourceOK := legacyLabelMatchers(row.SourceMatch)
			target, targetOK := legacyLabelMatchers(row.TargetMatch)
			equal := legacyLabelSelectors(row.EqualLabels)
			updates := map[string]interface{}{"source_matchers": encodeMigrationJSON(source), "target_matchers": encodeMigrationJSON(target), "equal_labels_v2": encodeMigrationJSON(equal)}
			if !sourceOK || !targetOK {
				updates["status"] = 0
			}
			if err := DB.Table("alert_inhibition_rules").Where("id = ?", row.ID).Updates(updates).Error; err != nil {
				return err
			}
		}
		for _, column := range []string{"source_match", "target_match", "equal_labels"} {
			if err := DB.Migrator().DropColumn("alert_inhibition_rules", column); err != nil {
				return err
			}
		}
	}
	return DB.Model(&models.AlertNotificationRoute{}).Where("label_namespace = '' OR label_namespace IS NULL").Update("label_namespace", "labels").Error
}

func legacyLabelMatchers(value string) ([]models.AlertLabelMatcher, bool) {
	parts := strings.Split(value, ",")
	result := make([]models.AlertLabelMatcher, 0, len(parts))
	for _, part := range parts {
		key, labelValue, found := strings.Cut(strings.TrimSpace(part), "=")
		key, labelValue = strings.TrimSpace(key), strings.TrimSpace(labelValue)
		if !found || key == "" || labelValue == "" {
			return []models.AlertLabelMatcher{}, false
		}
		result = append(result, models.AlertLabelMatcher{Namespace: "labels", Key: key, Value: labelValue})
	}
	return result, len(result) > 0
}

func legacyLabelSelectors(value string) []models.AlertLabelSelector {
	result := make([]models.AlertLabelSelector, 0)
	for _, key := range strings.Split(value, ",") {
		if key = strings.TrimSpace(key); key != "" {
			result = append(result, models.AlertLabelSelector{Namespace: "labels", Key: key})
		}
	}
	return result
}

func encodeMigrationJSON(value interface{}) string {
	encoded, _ := json.Marshal(value)
	return string(encoded)
}

func migrateAlertLifecycle() error {
	if err := DB.Model(&models.AlertRule{}).Where("no_data_policy = '' OR no_data_policy IS NULL").Update("no_data_policy", "keep").Error; err != nil {
		return err
	}
	if err := migrateTerminatedAlertStates(); err != nil {
		return err
	}
	var alerts []models.AlertInstance
	if err := DB.Where("NOT EXISTS (SELECT 1 FROM alert_events WHERE alert_events.alert_id = alert_instances.id)").Find(&alerts).Error; err != nil {
		return err
	}
	events := make([]models.AlertEvent, 0, len(alerts))
	for _, alert := range alerts {
		createdAt := alert.CreatedAt
		if createdAt.IsZero() {
			createdAt = alert.StateChangedAt
		}
		events = append(events, models.AlertEvent{
			AlertID: alert.ID, EventType: "snapshot", FromStatus: alert.Status, ToStatus: alert.Status,
			FromSeverity: alert.Severity, ToSeverity: alert.Severity, Reason: "migration_snapshot",
			Labels: alert.Labels, EvaluationStatus: alert.LastProbeStatus, CreatedAt: createdAt,
		})
	}
	if len(events) == 0 {
		return nil
	}
	return DB.CreateInBatches(events, 200).Error
}

func migrateTerminatedAlertStates() error {
	var alerts []models.AlertInstance
	forcedReasons := []string{"rule_disabled", "rule_deleted", "datasource_unbound"}
	if err := DB.Where(
		"status = ? AND EXISTS (SELECT 1 FROM alert_events WHERE alert_events.alert_id = alert_instances.id AND alert_events.event_type = ? AND alert_events.reason IN ?)",
		"resolved", "state_changed", forcedReasons,
	).Find(&alerts).Error; err != nil {
		return err
	}
	reasons := map[string]string{
		"rule_disabled":      "告警规则已停用",
		"rule_deleted":       "告警规则已删除",
		"datasource_unbound": "数据源已解除绑定",
	}
	for i := range alerts {
		var event models.AlertEvent
		result := DB.Where("alert_id = ? AND event_type = ?", alerts[i].ID, "state_changed").Order("created_at DESC, id DESC").First(&event)
		reason, forced := reasons[event.Reason]
		if result.Error != nil || !forced {
			continue
		}
		if err := DB.Model(&alerts[i]).Updates(map[string]interface{}{
			"status": "terminated", "notification_status": "terminated",
			"notification_reason": reason, "next_notification_at": nil,
		}).Error; err != nil {
			return err
		}
		if err := DB.Model(&event).Update("to_status", "terminated").Error; err != nil {
			return err
		}
	}
	return nil
}

func migrateAlertRuleGroups() error {
	// Older versions enforced rule names globally. Names are now unique only inside a group.
	for _, indexName := range []string{"uni_alert_rules_name", "idx_alert_rules_name"} {
		if DB.Migrator().HasIndex(&models.AlertRule{}, indexName) {
			if err := DB.Migrator().DropIndex(&models.AlertRule{}, indexName); err != nil {
				return err
			}
		}
	}
	if err := DB.Model(&models.AlertRule{}).Where("datasource_mode = '' OR datasource_mode IS NULL").Update("datasource_mode", "custom").Error; err != nil {
		return err
	}
	// Data source groups used to belong to rule groups. Preserve existing
	// inherited links by moving them onto each rule during the transition.
	if DB.Migrator().HasColumn("alert_rule_groups", "datasource_group_id") {
		if err := DB.Exec(`UPDATE alert_rules AS r
            SET datasource_group_id = (SELECT g.datasource_group_id FROM alert_rule_groups AS g WHERE g.id = r.group_id)
            WHERE r.datasource_mode = 'inherit' AND r.datasource_group_id = 0
            AND EXISTS (SELECT 1 FROM alert_rule_groups AS g WHERE g.id = r.group_id AND g.datasource_group_id > 0)`).Error; err != nil {
			return err
		}
	}
	if err := migrateAlertRuleDatasourceSelections(); err != nil {
		return err
	}
	var orphanCount int64
	if err := DB.Model(&models.AlertRule{}).Where("group_id = 0").Count(&orphanCount).Error; err != nil || orphanCount == 0 {
		return err
	}
	var group models.AlertRuleGroup
	if err := DB.Where("name = ?", "默认规则组").FirstOrCreate(&group, models.AlertRuleGroup{
		Name: "默认规则组", Description: "系统迁移的已有告警规则",
	}).Error; err != nil {
		return err
	}
	return DB.Model(&models.AlertRule{}).Where("group_id = 0").Update("group_id", group.ID).Error
}

func migrateAlertRuleDatasourceSelections() error {
	var rules []models.AlertRule
	if err := DB.Find(&rules).Error; err != nil {
		return err
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		for _, rule := range rules {
			var groupSelectionCount, directSelectionCount int64
			tx.Model(&models.AlertRuleDatasourceGroup{}).Where("rule_id = ?", rule.ID).Count(&groupSelectionCount)
			tx.Model(&models.AlertRuleDatasource{}).Where("rule_id = ?", rule.ID).Count(&directSelectionCount)
			if groupSelectionCount == 0 && rule.DatasourceGroupID > 0 {
				if err := tx.Create(&models.AlertRuleDatasourceGroup{RuleID: rule.ID, DatasourceGroupID: rule.DatasourceGroupID}).Error; err != nil {
					return err
				}
			}
			if directSelectionCount > 0 || rule.DatasourceMode != "custom" {
				continue
			}
			var bindings []models.AlertRuleBinding
			if err := tx.Where("rule_id = ?", rule.ID).Find(&bindings).Error; err != nil {
				return err
			}
			for _, binding := range bindings {
				if err := tx.Create(&models.AlertRuleDatasource{RuleID: rule.ID, DatasourceID: binding.DatasourceID}).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}
