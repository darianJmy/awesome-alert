package handlers

import (
	"guanlan/internal/models"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestDashboardOverviewTimeBucketsAndCounts(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err := db.AutoMigrate(&models.DatasourceGroup{}, &models.Datasource{}, &models.FaultCenter{}, &models.AlertEvent{}, &models.AlertInstance{}, &models.NotificationRecord{}, &models.AgentConnection{}, &models.InspectionItem{}, &models.InspectionRecord{}); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 10, 20, 3, 30, 0, 0, time.UTC) // 11:30 Beijing
	out, err := loadDashboardOverview(db, now, "today")
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Points) != 12 || len(out.Severity) != 0 || out.Summary.TodayAlerts != 0 {
		t.Fatalf("bad empty response: %#v", out)
	}
	fresh, stale := now.Add(-time.Minute), now.Add(-3*time.Minute)
	rows := []interface{}{
		&models.AgentConnection{Name: "fresh", TokenHash: "a", Status: "online", LastSeenAt: &fresh},
		&models.AgentConnection{Name: "stale", TokenHash: "b", Status: "online", LastSeenAt: &stale},
		&models.AgentConnection{Name: "offline", TokenHash: "c", Status: "offline"},
		&models.AgentConnection{Name: "deleted", TokenHash: "d", Status: "deleted"},
		&models.InspectionRuleGroup{ID: 1, Name: "probe-group"},
		&models.InspectionItem{Name: "enabled", Type: "http", Target: "example.com", Status: 1, LastStatus: "fail"},
		&models.InspectionItem{Name: "disabled", Type: "tcp", Target: "example.com", Status: 1, LastStatus: "error"},
		&models.InspectionItem{Name: "retired", Type: "mqtt", Target: "example.com", Status: 1, LastStatus: "fail"},
		&models.InspectionRecord{ItemID: 1, Status: "success", CreatedAt: now},
		&models.InspectionRecord{ItemID: 1, Status: "fail", CreatedAt: now},
		&models.InspectionRecord{ItemID: 1, Status: "error", CreatedAt: now},
		&models.InspectionRecord{ItemID: 1, Status: "fail", CreatedAt: now.Add(-24 * time.Hour)},
		&models.Datasource{Name: "vm", Type: "victoriametrics", Status: 1},
		&models.Datasource{Name: "archived mysql", Type: "mysql", Status: 1},
		&models.DatasourceGroup{Name: "华东"}, &models.Datasource{Name: "prom", Type: "prometheus"}, &models.FaultCenter{Name: "生产"},
		&models.AlertInstance{Source: "test", Title: "active", Status: "firing", Severity: "P1", StartedAt: now, LastEventAt: now},
		&models.AlertInstance{Source: "test", Title: "resolved", Status: "resolved", Severity: "P1", StartedAt: now, LastEventAt: now},
		&models.AlertEvent{AlertID: 1, EventType: "created", ToStatus: "firing", CreatedAt: time.Date(2026, 10, 19, 16, 0, 0, 0, time.UTC)},
		&models.AlertEvent{AlertID: 1, EventType: "state_changed", FromStatus: "recovering", ToStatus: "firing", CreatedAt: now},
		&models.AlertEvent{AlertID: 1, EventType: "notification", FromStatus: "firing", ToStatus: "firing", CreatedAt: now},
		&models.AlertEvent{AlertID: 1, EventType: "created", ToStatus: "firing", CreatedAt: time.Date(2026, 10, 19, 15, 59, 59, 0, time.UTC)},
		&models.AlertEvent{AlertID: 1, EventType: "created", ToStatus: "firing", CreatedAt: now.Add(time.Hour)},
		&models.NotificationRecord{EventType: "test", Status: "success", CreatedAt: now},
		&models.NotificationRecord{EventType: "test", Status: "failed", CreatedAt: now},
	}
	for _, row := range rows {
		if err := db.Create(row).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Model(&models.InspectionItem{}).Where("name = ?", "disabled").Update("status", 0).Error; err != nil {
		t.Fatal(err)
	}
	out, err = loadDashboardOverview(db, now, "today")
	if err != nil {
		t.Fatal(err)
	}
	s := out.Summary
	if s.Prometheus != 1 || s.PrometheusEnabled != 1 || s.ProbeAgents != 3 || s.ProbeAgentsOnline != 1 || s.ProbeTasks != 2 || s.ProbeTasksEnabled != 1 || s.ProbeTasksUnhealthy != 1 || s.TodayProbeRuns != 3 || s.TodayProbeFailures != 2 {
		t.Fatalf("bad operations counts: %#v", s)
	}
	if s.DataCenters != 1 || s.Datasources != 2 || s.DatasourcesEnabled != 2 || s.FaultCenters != 1 || s.ActiveAlerts != 1 || s.TodayAlerts != 2 || s.TodayNotificationsSuccess != 1 || s.TodayNotificationsFailed != 1 {
		t.Fatalf("bad totals: %#v", s)
	}
	if out.Points[0].Alerts != 1 || out.Points[11].Alerts != 1 || out.Points[11].NotificationsSuccess != 1 || out.Points[11].NotificationsFailed != 1 {
		t.Fatalf("bad hourly buckets: %#v", out.Points)
	}
	for _, tc := range []struct {
		key   string
		count int
	}{{"7d", 7}, {"30d", 30}} {
		out, err = loadDashboardOverview(db, now, tc.key)
		if err != nil {
			t.Fatal(err)
		}
		sum := int64(0)
		for _, p := range out.Points {
			sum += p.Alerts
		}
		if len(out.Points) != tc.count || sum != 3 || out.Points[tc.count-1].Alerts != 2 {
			t.Fatalf("bad daily buckets: %#v", out.Points)
		}
	}
	if _, err := loadDashboardOverview(db, now, "invalid"); err == nil {
		t.Fatal("invalid range accepted")
	}
}
