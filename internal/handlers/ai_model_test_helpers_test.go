package handlers

import (
	"testing"

	"guanlan/internal/database"
	"guanlan/internal/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupAIModelTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	conn, _ := db.DB()
	conn.SetMaxOpenConns(1)
	old := database.DB
	database.DB = db
	t.Cleanup(func() { database.DB = old; conn.Close() })
	if err := db.AutoMigrate(&models.AIModel{}); err != nil {
		t.Fatal(err)
	}
	return db
}
