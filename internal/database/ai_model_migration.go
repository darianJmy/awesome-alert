package database

import (
	"fmt"
	"gorm.io/gorm"
	"guanlan/internal/models"
	sqldb "guanlan/pkg/database"
)

// Only the known preceding version is upgraded. Changes and version marker commit atomically.
func upgradeAIModelSchema(db *gorm.DB) error {
	if !db.Migrator().HasTable("schema_migrations") {
		return nil
	}
	if sqldb.RequireSchemaVersion(db, CurrentSchemaVersion) == nil {
		return nil
	}
	if err := sqldb.RequireSchemaVersion(db, "2026102001"); err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if sqldb.RequireSchemaVersion(tx, CurrentSchemaVersion) == nil {
			return nil
		}
		if err := tx.Migrator().CreateTable(&models.AIModel{}); err != nil {
			return fmt.Errorf("create AI models: %w", err)
		}
		if err := tx.Migrator().AddColumn(&models.ChatConversation{}, "ModelID"); err != nil {
			return err
		}
		menu := models.Menu{ParentID: 87, Name: "AI 模型", Path: "/settings/ai-models", Component: "system/AIModelManage", Icon: "Cpu", Sort: 6, Type: 2, Permission: "system-setting:list", Status: 1}
		if err := tx.Create(&menu).Error; err != nil {
			return err
		}
		if err := tx.Exec(`INSERT INTO role_menus (role_id,menu_id) SELECT id,? FROM roles WHERE code = 'admin'`, menu.ID).Error; err != nil {
			return err
		}
		return sqldb.RecordSchemaVersion(tx, CurrentSchemaVersion)
	})
}
