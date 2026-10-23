package handlers

import (
	"guanlan/internal/database"
	"guanlan/internal/models"
	"sync"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestSQLiteInspectionTrendUpsert(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:?_pragma=foreign_keys(1)"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	pool, _ := db.DB()
	pool.SetMaxOpenConns(1)
	previous := database.DB
	database.DB = db
	t.Cleanup(func() { database.DB = previous; _ = pool.Close() })
	if err := db.AutoMigrate(&models.InspectionTrendBucket{}); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	node := inspectionExecutionNode{NodeType: "direct", NodeID: 0, NodeName: "local"}
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Go(func() { recordInspectionTrend(1, node, "success", `{"status_code":200}`, 100, now) })
		wg.Go(func() { recordInspectionTrend(1, node, "fail", `{"status_code":500}`, 300, now) })
	}
	wg.Wait()
	var buckets []models.InspectionTrendBucket
	if err := db.Find(&buckets).Error; err != nil {
		t.Fatal(err)
	}
	if len(buckets) != 1 {
		t.Fatalf("expected one bucket, got %d", len(buckets))
	}
	bucket := buckets[0]
	if bucket.TotalCount != 20 || bucket.SuccessCount != 10 || bucket.FailureCount != 10 || bucket.DurationSumMs != 4000 || bucket.DurationMinMs != 100 || bucket.DurationMaxMs != 300 || bucket.Status2xxCount != 10 || bucket.Status5xxCount != 10 {
		t.Fatalf("incorrect aggregate: %+v", bucket)
	}
	recordInspectionTrend(1, node, "success", `{}`, 50, now.Add(inspectionTrendBucketSize))
	var count int64
	if err := db.Model(&models.InspectionTrendBucket{}).Count(&count).Error; err != nil || count != 2 {
		t.Fatalf("next bucket: count=%d err=%v", count, err)
	}
}
