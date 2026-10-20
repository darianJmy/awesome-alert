package database

import (
	"fmt"
	"gorm.io/gorm"
	"time"
)

type schemaMigration struct {
	Version   string    `gorm:"size:64;primaryKey"`
	AppliedAt time.Time `gorm:"not null;autoCreateTime:milli"`
}

func (schemaMigration) TableName() string {
	return "schema_migrations"
}

func RequireSchemaVersion(db *gorm.DB, version string) error {
	if !db.Migrator().HasTable(&schemaMigration{}) {
		return fmt.Errorf("existing database has no schema version; refusing to overwrite its data")
	}
	var count int64
	if err := db.Model(&schemaMigration{}).Where("version = ?", version).Count(&count).Error; err != nil {
		return fmt.Errorf("read schema version: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("required schema version %s has not been applied", version)
	}
	return nil
}

func RecordSchemaVersion(db *gorm.DB, version string) error {
	if err := db.AutoMigrate(&schemaMigration{}); err != nil {
		return fmt.Errorf("create schema migration table: %w", err)
	}
	result := db.FirstOrCreate(&schemaMigration{Version: version}, schemaMigration{Version: version})
	if result.Error != nil {
		return fmt.Errorf("record schema version %s: %w", version, result.Error)
	}
	return nil
}

// Initialize atomically creates and seeds an empty SQLite database; existing databases are version-checked only.
func Initialize(db *gorm.DB, version string, createSchema, seed func(*gorm.DB) error) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var tableCount int64
		if err := tx.Raw(`SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%'`).Scan(&tableCount).Error; err != nil {
			return fmt.Errorf("inspect database: %w", err)
		}
		if tableCount != 0 {
			return RequireSchemaVersion(tx, version)
		}
		if err := createSchema(tx); err != nil {
			return err
		}
		if err := seed(tx); err != nil {
			return fmt.Errorf("seed initial data: %w", err)
		}
		return RecordSchemaVersion(tx, version)
	})
}

// AutoMigrate applies the caller's model definitions without importing business models.
func AutoMigrate(db *gorm.DB, models ...any) error {
	if err := db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("create database schema: %w", err)
	}
	return nil
}
