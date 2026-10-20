package database

import (
	"errors"
	"path/filepath"
	"testing"

	"gorm.io/gorm"
)

func TestInitializeWithCallerModelsAndSeed(t *testing.T) {
	db, err := OpenSQLite(filepath.Join(t.TempDir(), "nested", "test.db"), nil)
	if err != nil {
		t.Fatal(err)
	}
	pool, _ := db.DB()
	defer pool.Close()
	type Widget struct {
		ID   uint `gorm:"primaryKey"`
		Name string
	}
	create := func(tx *gorm.DB) error { return AutoMigrate(tx, &Widget{}) }
	failure := errors.New("seed failed")
	if err := Initialize(db, "v1", create, func(tx *gorm.DB) error { return failure }); !errors.Is(err, failure) {
		t.Fatalf("unexpected error: %v", err)
	}
	if db.Migrator().HasTable(&Widget{}) {
		t.Fatal("failed initialization did not roll back schema")
	}
	calls := 0
	seed := func(tx *gorm.DB) error { calls++; return tx.Create(&Widget{Name: "first"}).Error }
	for i := 0; i < 2; i++ {
		if err := Initialize(db, "v1", create, seed); err != nil {
			t.Fatal(err)
		}
	}
	if calls != 1 {
		t.Fatalf("seed called %d times", calls)
	}
	if err := Initialize(db, "v2", create, seed); err == nil {
		t.Fatal("unsupported version must fail")
	}
	var count int64
	if err := db.Model(&Widget{}).Count(&count).Error; err != nil || count != 1 {
		t.Fatalf("data changed: %d %v", count, err)
	}
}
