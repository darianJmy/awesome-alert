package handlers

import (
	"guanlan/internal/models"
	"testing"
)

func TestRetiredMCPServiceMenuFiltering(t *testing.T) {
	menus := []models.Menu{
		{ID: 92, Path: "/settings/mcp", Status: 1},
		{ID: 101, Path: "/custom-mcp", Component: "system/MCPServiceManage", Status: 1},
		{ID: 102, Type: 3, Permission: "mcp-service:edit", Status: 1},
		{ID: 103, Path: "/settings/ai-models", Permission: "ai-model:list", Status: 1},
		{ID: 90, Path: "/settings/api-tokens", Permission: "api-token:list", Status: 1},
	}
	filtered := withoutRetiredMenus(menus)
	if len(filtered) != 2 || filtered[0].ID != 103 || filtered[1].ID != 90 {
		t.Fatalf("unexpected menus: %+v", filtered)
	}
}

func TestRetiredChatMenuFiltering(t *testing.T) {
	menus := []models.Menu{
		{ID: 1, Path: "/dashboard", Component: "Dashboard"},
		{ID: 17, Path: "/chat", Component: "Chat"},
		{ID: 99, Path: "/renamed-chat", Component: "Chat"},
		{ID: 100, Path: "/settings/basic", Component: "system/BasicSetting"},
	}
	result := withoutRetiredMenus(menus)
	if len(result) != 2 || result[0].ID != 1 || result[1].ID != 100 {
		t.Fatalf("unexpected menus: %#v", result)
	}
	if len(menus) != 4 {
		t.Fatal("source menu collection changed")
	}
}

func TestRetiredScheduledTaskMenuFiltering(t *testing.T) {
	menus := []models.Menu{
		{ID: 1, Path: "/automation/tasks", Component: "automation/AutomationTaskManage"},
		{ID: 2, Path: "/automation/scheduled-tasks"},
		{ID: 3, Path: "/inspection/scheduled-tasks"},
		{ID: 4, Path: "/custom", Component: "inspection/ScheduledTaskManage"},
		{ID: 5, Permission: "task:run", Type: 3},
		{ID: 6, Path: "/inspection/items"},
	}
	filtered := withoutRetiredMenus(menus)
	if len(filtered) != 2 || filtered[0].ID != 1 || filtered[1].ID != 6 {
		t.Fatalf("unexpected menus: %+v", filtered)
	}
}

func TestStandaloneProbeMenuFiltering(t *testing.T) {
	menus := []models.Menu{
		{ID: 10, Path: "/inspection/probe", Component: "inspection/InstantProbe"},
		{ID: 101, Path: "/custom-probe", Component: "inspection/InstantProbe"},
		{ID: 11, Path: "/inspection/items", Component: "inspection/InspectionRuleGroupList"},
	}
	filtered := withoutRetiredMenus(menus)
	if len(filtered) != 1 || filtered[0].ID != 11 {
		t.Fatalf("unexpected menus: %+v", filtered)
	}
}

func TestProbeTaskMenuIsTopLevel(t *testing.T) {
	menus := []models.Menu{
		{ID: 9, Type: 1, Path: "/inspection", Name: "拨测中心"},
		{ID: 11, ParentID: 9, Type: 2, Path: "/inspection/items", Permission: "inspection:list"},
	}
	filtered := withoutRetiredMenus(menus)
	if len(filtered) != 1 || filtered[0].ID != 11 || filtered[0].ParentID != 0 || filtered[0].Name != "拨测任务" || filtered[0].Permission != "inspection:list" {
		t.Fatalf("unexpected probe navigation: %+v", filtered)
	}
	tree := buildMenuTree(map[uint]models.Menu{11: monitoringMenu(menus[1])}, 0)
	if len(tree) != 1 || tree[0].ID != 11 {
		t.Fatalf("task role lost its menu: %+v", tree)
	}
}
