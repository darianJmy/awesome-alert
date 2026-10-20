package handlers

import (
	"guanlan/internal/models"
	"testing"
)

func TestRetiredChatMenuFiltering(t *testing.T) {
	menus := []models.Menu{
		{ID: 1, Path: "/dashboard", Component: "Dashboard"},
		{ID: 17, Path: "/chat", Component: "Chat"},
		{ID: 99, Path: "/renamed-chat", Component: "Chat"},
		{ID: 100, Path: "/settings/basic", Component: "system/BasicSetting"},
	}
	result := withoutRetiredChatMenus(menus)
	if len(result) != 2 || result[0].ID != 1 || result[1].ID != 100 {
		t.Fatalf("unexpected menus: %#v", result)
	}
	if len(menus) != 4 {
		t.Fatal("source menu collection changed")
	}
}
