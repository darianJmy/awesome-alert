package handlers

import (
	"context"
	"errors"
	"guanlan/internal/database"
	"guanlan/internal/models"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestChatExecutionLimits(t *testing.T) {
	release, ok := acquireChatExecution(100, 1)
	if !ok {
		t.Fatal("first run rejected")
	}
	defer release()
	if _, ok := acquireChatExecution(100, 1); ok {
		t.Fatal("duplicate run accepted")
	}
	for id := uint(2); id <= 4; id++ {
		release, ok := acquireChatExecution(100, id)
		if !ok {
			t.Fatal("capacity too low")
		}
		defer release()
	}
	if _, ok := acquireChatExecution(101, 9); ok {
		t.Fatal("global capacity exceeded")
	}
	release()
	next, ok := acquireChatExecution(101, 9)
	if !ok {
		t.Fatal("capacity not released")
	}
	next()
}

func TestAIToolPermissionsAndHistory(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	previous := database.DB
	database.DB = db
	defer func() { database.DB = previous }()
	if err := db.AutoMigrate(&models.User{}, &models.Role{}, &models.Permission{}, &models.ChatMessage{}); err != nil {
		t.Fatal(err)
	}
	statements := []string{
		"INSERT INTO users (id,username,password,status) VALUES (1,'tester','unused',1)",
		"INSERT INTO roles (id,name,code,status) VALUES (1,'reader','reader',1)",
		"INSERT INTO permissions (id,name,code,status) VALUES (1,'read','datasource:list',1)",
		"INSERT INTO user_roles (user_id,role_id) VALUES (1,1)",
		"INSERT INTO role_permissions (role_id,permission_id) VALUES (1,1)",
	}
	for _, sql := range statements {
		if err := db.Exec(sql).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := authorizeAITool(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	if err := authorizeAITool(context.Background(), 2); err == nil {
		t.Fatal("unprivileged user allowed")
	}
	db.Exec("UPDATE users SET status=0 WHERE id=1")
	if err := authorizeAITool(context.Background(), 1); err == nil {
		t.Fatal("disabled user allowed")
	}
	for _, msg := range []models.ChatMessage{{ID: 1, ConversationID: 7, Role: "user", Content: "previous"}, {ID: 2, ConversationID: 8, Role: "user", Content: "other"}, {ID: 3, ConversationID: 7, Role: "user", Content: "current"}} {
		if err := db.Create(&msg).Error; err != nil {
			t.Fatal(err)
		}
	}
	if got := collectChatHistoryBefore(7, 3); got != "user: previous\n" {
		t.Fatalf("history not scoped: %q", got)
	}
}

func TestLLMRetryWaitCanBeCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	start := time.Now()
	if err := waitForLLMRetry(ctx, time.Hour); !errors.Is(err, context.Canceled) {
		t.Fatalf("unexpected error: %v", err)
	}
	if time.Since(start) > time.Second {
		t.Fatal("cancelled retry blocked")
	}
}
