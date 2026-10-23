package database

import (
	"guanlan/internal/models"
	"testing"
)

func TestDefaultFaultCenterSurvivesReopen(t *testing.T) {
	db := freshSQLite(t)
	var center models.FaultCenter
	if err := db.Where("is_default = ?", true).First(&center).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&center).Update("name", "renamed default").Error; err != nil {
		t.Fatal(err)
	}
	if err := prepareDatabase(db); err != nil {
		t.Fatal(err)
	}
	var count int64
	db.Model(&models.FaultCenter{}).Where("is_default = ?", true).Count(&count)
	if count != 1 {
		t.Fatal("reopen recreated the default center")
	}
	if err := db.First(&center, center.ID).Error; err != nil || center.Name != "renamed default" {
		t.Fatal("reopen changed default center")
	}
}
