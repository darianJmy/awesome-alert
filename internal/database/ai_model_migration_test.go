package database

import (
	"guanlan/internal/models"
	sqldb "guanlan/pkg/database"
	"testing"
)

func TestAIModelSchemaUpgradePreservesData(t *testing.T) {
	db, err := openDatabaseDSN(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	conn, _ := db.DB()
	defer conn.Close()
	if err := db.Exec(`CREATE TABLE chat_conversations (id INTEGER PRIMARY KEY,user_id INTEGER,title TEXT,workspace_id INTEGER,created_at DATETIME,updated_at DATETIME)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.Menu{}, &models.Role{}); err != nil {
		t.Fatal(err)
	}
	db.Exec("INSERT INTO chat_conversations (id,user_id,title) VALUES (1,7,'保留会话')")
	db.Create(&models.Role{Name: "admin", Code: "admin", Status: 1})
	if err := sqldb.RecordSchemaVersion(db, "2026102001"); err != nil {
		t.Fatal(err)
	}
	if err := upgradeAIModelSchema(db); err != nil {
		t.Fatal(err)
	}
	if err := upgradeAIModelSchema(db); err != nil {
		t.Fatal(err)
	}
	var convo models.ChatConversation
	if err := db.First(&convo, 1).Error; err != nil || convo.Title != "保留会话" || convo.ModelID != 0 {
		t.Fatalf("data changed: %+v %v", convo, err)
	}
	if err := sqldb.RequireSchemaVersion(db, CurrentSchemaVersion); err != nil {
		t.Fatal(err)
	}
	var count int64
	db.Model(&models.Menu{}).Where("path = ?", "/settings/ai-models").Count(&count)
	if count != 1 {
		t.Fatal("menu not idempotent")
	}
}
