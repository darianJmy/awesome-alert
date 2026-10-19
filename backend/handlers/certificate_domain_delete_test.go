package handlers

import (
	"fmt"
	"os"
	"testing"
	"time"

	"aiops-backend/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestDeleteCertificateDomainLifecycle(t *testing.T) {
	dsn := os.Getenv("CERTIFICATE_DELETE_TEST_DSN")
	if dsn == "" {
		t.Skip("set CERTIFICATE_DELETE_TEST_DSN to an isolated MySQL test database")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.CertificateDomain{}, &models.AlertInstance{}, &models.AlertEvent{}); err != nil {
		t.Fatal(err)
	}
	now := time.Now().Truncate(time.Millisecond)
	domain := models.CertificateDomain{Domain: "delete-test.invalid"}
	if err := db.Create(&domain).Error; err != nil {
		t.Fatal(err)
	}
	statuses := []string{alertStatusPending, alertStatusFiring, alertStatusRecovering, alertStatusResolved, alertStatusTerminated}
	var alerts []models.AlertInstance
	for _, status := range statuses {
		alert := models.AlertInstance{Source: "certificate-domain", SourceID: domain.ID, Title: status, Status: status, Severity: "P2", StartedAt: now, LastEventAt: now, StateChangedAt: now, NextNotificationAt: &now}
		if err := db.Create(&alert).Error; err != nil {
			t.Fatal(err)
		}
		alerts = append(alerts, alert)
	}
	unrelated := models.AlertInstance{Source: "probe-task", SourceID: domain.ID, Title: "unrelated", Status: alertStatusFiring, StartedAt: now, LastEventAt: now, StateChangedAt: now}
	if err := db.Create(&unrelated).Error; err != nil {
		t.Fatal(err)
	}
	if err := deleteCertificateDomain(db, domain.ID, now); err != nil {
		t.Fatal(err)
	}
	var count int64
	db.Model(&models.CertificateDomain{}).Where("id = ?", domain.ID).Count(&count)
	if count != 0 {
		t.Fatal("domain remains")
	}
	for i, before := range alerts {
		var after models.AlertInstance
		if err := db.First(&after, before.ID).Error; err != nil {
			t.Fatal("alert history was deleted:", err)
		}
		if i < 3 {
			if after.Status != alertStatusTerminated || after.ResolvedAt == nil || after.NextNotificationAt != nil || after.NotificationStatus != notificationStatusTerminated {
				t.Fatalf("active alert not terminated: %+v", after)
			}
		} else if after.Status != before.Status || after.ResolvedAt != nil {
			t.Fatal("historical alert changed")
		}
	}
	db.First(&unrelated, unrelated.ID)
	if unrelated.Status != alertStatusFiring {
		t.Fatal("unrelated source changed")
	}
	db.Model(&models.AlertEvent{}).Where("reason = ?", "certificate_domain_deleted").Count(&count)
	if count != 3 {
		t.Fatalf("want 3 termination events, got %d", count)
	}
	if err := deleteCertificateDomain(db, domain.ID, now); err != nil {
		t.Fatal(err)
	}
	db.Model(&models.AlertEvent{}).Count(&count)
	if count != 3 {
		t.Fatal("repeated deletion duplicated history")
	}

	// Force event persistence to fail and verify deletion and state changes roll back.
	domain.ID = 0
	domain.Domain = "rollback-test.invalid"
	if err := db.Create(&domain).Error; err != nil {
		t.Fatal(err)
	}
	alert := models.AlertInstance{Source: "certificate-domain", SourceID: domain.ID, Title: "rollback", Status: alertStatusFiring, StartedAt: now, LastEventAt: now, StateChangedAt: now}
	if err := db.Create(&alert).Error; err != nil {
		t.Fatal(err)
	}
	callback := "test:fail_alert_event"
	db.Callback().Create().Before("gorm:create").Register(callback, func(tx *gorm.DB) {
		if tx.Statement.Table == "alert_events" {
			tx.AddError(fmt.Errorf("forced event failure"))
		}
	})
	defer db.Callback().Create().Remove(callback)
	if err := deleteCertificateDomain(db, domain.ID, now); err == nil {
		t.Fatal("expected event failure")
	}
	db.Model(&models.CertificateDomain{}).Where("id = ?", domain.ID).Count(&count)
	if count != 1 {
		t.Fatal("domain deletion was not rolled back")
	}
	db.First(&alert, alert.ID)
	if alert.Status != alertStatusFiring {
		t.Fatal("alert update was not rolled back")
	}
}
