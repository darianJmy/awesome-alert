package handlers

import (
	"guanlan/internal/models"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestChatConversationScopeAlwaysFiltersUser(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{DryRun: true})
	if err != nil {
		t.Fatalf("create dry-run database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	query := chatConversationScope(db.Model(&models.ChatConversation{}), 42).
		Where("id = ?", 7).
		Find(&models.ChatConversation{})
	statement := query.Statement.SQL.String()
	if !strings.Contains(statement, "user_id") {
		t.Fatalf("conversation query is not user-scoped: %s", statement)
	}
	if len(query.Statement.Vars) != 2 || query.Statement.Vars[0] != uint(42) || query.Statement.Vars[1] != 7 {
		t.Fatalf("unexpected conversation scope vars: %#v", query.Statement.Vars)
	}
}
