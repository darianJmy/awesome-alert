package database

import (
	"guanlan/internal/models"
	"guanlan/pkg/logger"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func seedInitialData(db *gorm.DB) error {
	logger.Info("initializing database seed data")

	password, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)

	if err != nil {
		return err
	}
	adminUser := models.User{
		Username: "admin",
		Password: string(password),
		Nickname: "管理员",
		Email:    "admin@aiops.local",
		Status:   1,
	}
	if err := db.Create(&adminUser).Error; err != nil {
		return err
	}

	permissions := []models.Permission{
		{Name: "用户列表", Code: "user:list", Status: 1},
		{Name: "创建用户", Code: "user:create", Status: 1},
		{Name: "编辑用户", Code: "user:edit", Status: 1},
		{Name: "删除用户", Code: "user:delete", Status: 1},
		{Name: "角色列表", Code: "role:list", Status: 1},
		{Name: "创建角色", Code: "role:create", Status: 1},
		{Name: "编辑角色", Code: "role:edit", Status: 1},
		{Name: "删除角色", Code: "role:delete", Status: 1},
		{Name: "权限列表", Code: "perm:list", Status: 1},
		{Name: "菜单列表", Code: "menu:list", Status: 1},
		{Name: "分配角色", Code: "user:assign:role", Status: 1},
		{Name: "分配权限", Code: "role:assign:perm", Status: 1},
		{Name: "分配菜单", Code: "role:assign:menu", Status: 1},
		{Name: "接入列表", Code: "datasource:list", Status: 1},
		{Name: "创建接入", Code: "datasource:create", Status: 1},
		{Name: "编辑接入", Code: "datasource:edit", Status: 1},
		{Name: "删除接入", Code: "datasource:delete", Status: 1},
		{Name: "仪表盘列表", Code: "dashboard:list", Status: 1},
		{Name: "创建仪表盘", Code: "dashboard:create", Status: 1},
		{Name: "编辑仪表盘", Code: "dashboard:edit", Status: 1},
		{Name: "删除仪表盘", Code: "dashboard:delete", Status: 1},
		{Name: "拨测任务列表", Code: "inspection:list", Status: 1},
		{Name: "创建拨测任务", Code: "inspection:create", Status: 1},
		{Name: "编辑拨测任务", Code: "inspection:edit", Status: 1},
		{Name: "删除拨测任务", Code: "inspection:delete", Status: 1},
		{Name: "执行拨测", Code: "inspection:run", Status: 1},
		{Name: "脚本列表", Code: "script:list", Status: 1},
		{Name: "创建脚本", Code: "script:create", Status: 1},
		{Name: "编辑脚本", Code: "script:edit", Status: 1},
		{Name: "删除脚本", Code: "script:delete", Status: 1},
		{Name: "工作流列表", Code: "workflow:list", Status: 1},
		{Name: "创建工作流", Code: "workflow:create", Status: 1},
		{Name: "编辑工作流", Code: "workflow:edit", Status: 1},
		{Name: "删除工作流", Code: "workflow:delete", Status: 1},
		{Name: "执行工作流", Code: "workflow:run", Status: 1},
		{Name: "自动化任务列表", Code: "automation-task:list", Status: 1},
		{Name: "创建自动化任务", Code: "automation-task:create", Status: 1},
		{Name: "编辑自动化任务", Code: "automation-task:edit", Status: 1},
		{Name: "删除自动化任务", Code: "automation-task:delete", Status: 1},
		{Name: "执行自动化任务", Code: "automation-task:run", Status: 1},
		{Name: "Git 认证列表", Code: "git-credential:list", Status: 1},
		{Name: "维护 Git 认证", Code: "git-credential:edit", Status: 1},
		{Name: "API Token 列表", Code: "api-token:list", Status: 1},
		{Name: "维护 API Token", Code: "api-token:edit", Status: 1},
		{Name: "MCP 服务", Code: "mcp-service:list", Status: 1},
		{Name: "维护 MCP 服务", Code: "mcp-service:edit", Status: 1},
		{Name: "系统设置", Code: "system-setting:list", Status: 1},
		{Name: "维护系统设置", Code: "system-setting:edit", Status: 1},
		{Name: "定时任务列表", Code: "task:list", Status: 1},
		{Name: "创建定时任务", Code: "task:create", Status: 1},
		{Name: "编辑定时任务", Code: "task:edit", Status: 1},
		{Name: "删除定时任务", Code: "task:delete", Status: 1},
		{Name: "执行定时任务", Code: "task:run", Status: 1},
	}
	if err := db.Create(&permissions).Error; err != nil {
		return err
	}

	menus := []models.Menu{
		{ID: 1, ParentID: 0, Name: "用户管理", Path: "/system", Component: "", Icon: "User", Sort: 98, Type: 1, Status: 1},
		{ID: 2, ParentID: 1, Name: "用户列表", Path: "/system/users", Component: "system/UserManage", Icon: "User", Sort: 1, Type: 2, Permission: "user:list", Status: 1},
		{ID: 3, ParentID: 1, Name: "角色列表", Path: "/system/roles", Component: "system/RoleManage", Icon: "Avatar", Sort: 2, Type: 2, Permission: "role:list", Status: 1},
		{ID: 4, ParentID: 1, Name: "权限列表", Path: "/system/permissions", Component: "system/PermissionManage", Icon: "Lock", Sort: 3, Type: 2, Permission: "perm:list", Status: 1},
		{ID: 5, ParentID: 1, Name: "菜单列表", Path: "/system/menus", Component: "system/MenuManage", Icon: "Menu", Sort: 4, Type: 2, Permission: "menu:list", Status: 1},
		{ID: 87, ParentID: 0, Name: "系统设置", Path: "/settings", Component: "", Icon: "Tools", Sort: 99, Type: 1, Status: 1},
		{ID: 88, ParentID: 87, Name: "基础设置", Path: "/settings/basic", Component: "system/BasicSetting", Icon: "SetUp", Sort: 1, Type: 2, Permission: "system-setting:list", Status: 0},
		{ID: 89, ParentID: 87, Name: "登录认证", Path: "/settings/authentication", Component: "system/AuthSetting", Icon: "Lock", Sort: 2, Type: 2, Permission: "system-setting:list", Status: 1},
		{ID: 86, ParentID: 87, Name: "Git 认证", Path: "/settings/git-credentials", Component: "system/GitCredentialManage", Icon: "Key", Sort: 3, Type: 2, Permission: "git-credential:list", Status: 1},
		{ID: 90, ParentID: 87, Name: "访问令牌", Path: "/settings/api-tokens", Component: "system/APITokenManage", Icon: "Key", Sort: 4, Type: 2, Permission: "api-token:list", Status: 1},
		{ID: 92, ParentID: 87, Name: "MCP 服务", Path: "/settings/mcp", Component: "system/MCPServiceManage", Icon: "Connection", Sort: 5, Type: 2, Permission: "mcp-service:list", Status: 1},
		{ID: 7, ParentID: 0, Name: "接入中心", Path: "/integrations/sources", Component: "monitor/IntegrationManage", Icon: "Connection", Sort: 10, Type: 2, Permission: "datasource:list", Status: 1},
		{ID: 8, ParentID: 0, Name: "仪表盘", Path: "/dashboards", Component: "", Icon: "DataAnalysis", Sort: 1, Type: 1, Status: 1},
		{ID: 84, ParentID: 8, Name: "监控仪表盘", Path: "/dashboards/grafana", Component: "monitor/GrafanaDashboardManage", Icon: "DataAnalysis", Sort: 1, Type: 2, Permission: "dashboard:list", Status: 1},
		{ID: 85, ParentID: 8, Name: "日志仪表盘", Path: "/dashboards/sls", Component: "monitor/SLSDashboardManage", Icon: "TrendCharts", Sort: 2, Type: 2, Permission: "dashboard:list", Status: 1},
		{ID: 91, ParentID: 8, Name: "云监控2.0", Path: "/dashboards/cloud-monitor", Component: "monitor/CloudMonitorDashboardManage", Icon: "Cloudy", Sort: 3, Type: 2, Permission: "dashboard:list", Status: 1},
		{ID: 38, ParentID: 0, Name: "告警中心", Path: "/alert-center", Component: "", Icon: "Bell", Sort: 4, Type: 1, Status: 1},
		{ID: 19, ParentID: 38, Name: "即时查询", Path: "/monitor/query", Component: "monitor/InstantQuery", Icon: "Search", Sort: 2, Type: 2, Status: 1},
		{ID: 9, ParentID: 0, Name: "拨测中心", Path: "/inspection", Component: "", Icon: "AlarmClock", Sort: 6, Type: 1, Status: 1},
		{ID: 10, ParentID: 9, Name: "即时拨测", Path: "/inspection/probe", Component: "inspection/InstantProbe", Icon: "Connection", Sort: 1, Type: 2, Status: 1},
		{ID: 11, ParentID: 9, Name: "拨测任务", Path: "/inspection/items", Component: "inspection/InspectionManage", Icon: "Checked", Sort: 2, Type: 2, Permission: "inspection:list", Status: 1},
		{ID: 15, ParentID: 0, Name: "自动化中心", Path: "/automation", Component: "", Icon: "Operation", Sort: 7, Type: 1, Status: 1},
		{ID: 24, ParentID: 87, Name: "节点管理", Path: "/nodes", Component: "certificate/ExecutorManage", Icon: "Monitor", Sort: 5, Type: 2, Permission: "certificate:executor:list", Status: 1},
		{ID: 12, ParentID: 15, Name: "脚本库", Path: "/automation/scripts", Component: "automation/ScriptLibrary", Icon: "Document", Sort: 2, Type: 2, Status: 1},
		{ID: 13, ParentID: 15, Name: "自动化任务", Path: "/automation/tasks", Component: "automation/AutomationTaskManage", Icon: "SetUp", Sort: 3, Type: 2, Permission: "automation-task:list", Status: 1},
		{ID: 18, ParentID: 15, Name: "定时任务", Path: "/automation/scheduled-tasks", Component: "inspection/ScheduledTaskManage", Icon: "Timer", Sort: 4, Type: 2, Permission: "task:list", Status: 1},
		{ID: 14, ParentID: 15, Name: "执行记录", Path: "/automation/records", Component: "inspection/ExecutionRecords", Icon: "Tickets", Sort: 5, Type: 2, Status: 1},
	}
	if err := db.Create(&menus).Error; err != nil {
		return err
	}

	adminRole := models.Role{
		Name:        "管理员",
		Code:        "admin",
		Description: "系统管理员, 拥有所有权限",
		Status:      1,
		Permissions: permissions,
		Menus:       menus,
	}
	if err := db.Create(&adminRole).Error; err != nil {
		return err
	}

	if err := db.Model(&adminUser).Association("Roles").Append(&adminRole); err != nil {
		return err
	}

	return applyInitialRoadmapData(db)
}

func applyInitialRoadmapData(db *gorm.DB) error {
	// Menu order is administrator-managed data. Roadmap migrations may repair menu
	// metadata, but must not reset sort values on every application restart.
	menuSorts, err := snapshotMenuSorts(db)
	if err != nil {
		return err
	}

	permissions := []models.Permission{
		{Name: "节点列表", Code: "certificate:executor:list", Status: 1},
		{Name: "维护节点", Code: "certificate:executor:edit", Status: 1},
		{Name: "MCP 服务", Code: "mcp-service:list", Status: 1},
		{Name: "维护 MCP 服务", Code: "mcp-service:edit", Status: 1},
		{Name: "日志查询", Code: "observability:logs", Status: 1},
		{Name: "链路追踪", Code: "observability:traces", Status: 1},
		{Name: "查看故障告警", Code: "alert:event:list", Status: 1},
		{Name: "告警规则", Code: "alert:rule:list", Status: 1},
		{Name: "创建告警规则", Code: "alert:rule:create", Status: 1},
		{Name: "编辑告警规则", Code: "alert:rule:edit", Status: 1},
		{Name: "删除告警规则", Code: "alert:rule:delete", Status: 1},
		{Name: "执行告警规则", Code: "alert:rule:run", Status: 1},
		{Name: "通知对象", Code: "notification:channel:list", Status: 1},
		{Name: "维护通知对象", Code: "notification:channel:edit", Status: 1},
		{Name: "通知模板", Code: "notification:template:list", Status: 1},
		{Name: "维护通知模板", Code: "notification:template:edit", Status: 1},
		{Name: "通知记录", Code: "notification:record:list", Status: 1},
		{Name: "操作审计", Code: "security:audit:list", Status: 1},
		{Name: "自动化任务列表", Code: "automation-task:list", Status: 1},
		{Name: "创建自动化任务", Code: "automation-task:create", Status: 1},
		{Name: "编辑自动化任务", Code: "automation-task:edit", Status: 1},
		{Name: "删除自动化任务", Code: "automation-task:delete", Status: 1},
		{Name: "执行自动化任务", Code: "automation-task:run", Status: 1},
		{Name: "Git 认证列表", Code: "git-credential:list", Status: 1},
		{Name: "维护 Git 认证", Code: "git-credential:edit", Status: 1},
		{Name: "API Token 列表", Code: "api-token:list", Status: 1},
		{Name: "维护 API Token", Code: "api-token:edit", Status: 1},
		{Name: "系统设置", Code: "system-setting:list", Status: 1},
		{Name: "维护系统设置", Code: "system-setting:edit", Status: 1},
	}
	for _, item := range permissions {
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&item).Error; err != nil {
			return err
		}
	}

	menus := []models.Menu{
		{ID: 8, ParentID: 0, Name: "仪表盘", Path: "/dashboards", Component: "", Icon: "DataAnalysis", Sort: 1, Type: 1, Status: 1},
		{ID: 84, ParentID: 8, Name: "监控仪表盘", Path: "/dashboards/grafana", Component: "monitor/GrafanaDashboardManage", Icon: "DataAnalysis", Sort: 1, Type: 2, Permission: "dashboard:list", Status: 1},
		{ID: 85, ParentID: 8, Name: "日志仪表盘", Path: "/dashboards/sls", Component: "monitor/SLSDashboardManage", Icon: "TrendCharts", Sort: 2, Type: 2, Permission: "dashboard:list", Status: 1},
		{ID: 91, ParentID: 8, Name: "云监控2.0", Path: "/dashboards/cloud-monitor", Component: "monitor/CloudMonitorDashboardManage", Icon: "Cloudy", Sort: 3, Type: 2, Permission: "dashboard:list", Status: 1},
		{ID: 7, ParentID: 0, Name: "接入中心", Path: "/integrations/sources", Component: "monitor/IntegrationManage", Icon: "Connection", Sort: 10, Type: 2, Permission: "datasource:list", Status: 1},
		{ID: 15, ParentID: 0, Name: "自动化中心", Path: "/automation", Component: "", Icon: "Operation", Sort: 7, Type: 1, Status: 1},
		{ID: 24, ParentID: 87, Name: "节点管理", Path: "/nodes", Component: "certificate/ExecutorManage", Icon: "Monitor", Sort: 6, Type: 2, Permission: "certificate:executor:list", Status: 1},
		{ID: 87, ParentID: 0, Name: "系统设置", Path: "/settings", Component: "", Icon: "Tools", Sort: 99, Type: 1, Status: 1},
		{ID: 88, ParentID: 87, Name: "基础设置", Path: "/settings/basic", Component: "system/BasicSetting", Icon: "SetUp", Sort: 1, Type: 2, Permission: "system-setting:list", Status: 0},
		{ID: 89, ParentID: 87, Name: "登录认证", Path: "/settings/authentication", Component: "system/AuthSetting", Icon: "Lock", Sort: 2, Type: 2, Permission: "system-setting:list", Status: 1},
		{ID: 86, ParentID: 87, Name: "Git 认证", Path: "/settings/git-credentials", Component: "system/GitCredentialManage", Icon: "Key", Sort: 3, Type: 2, Permission: "git-credential:list", Status: 1},
		{ID: 90, ParentID: 87, Name: "访问令牌", Path: "/settings/api-tokens", Component: "system/APITokenManage", Icon: "Key", Sort: 4, Type: 2, Permission: "api-token:list", Status: 1},
		{ID: 92, ParentID: 87, Name: "MCP 服务", Path: "/settings/mcp", Component: "system/MCPServiceManage", Icon: "Connection", Sort: 5, Type: 2, Permission: "mcp-service:list", Status: 1},
		{ID: 30, ParentID: 0, Name: "可观测中心", Path: "/observability", Component: "", Icon: "DataLine", Sort: 5, Type: 1, Status: 1},
		{ID: 31, ParentID: 30, Name: "日志查询", Path: "/observability/logs", Component: "common/FeaturePlaceholder", Icon: "Document", Sort: 1, Type: 2, Permission: "observability:logs", Status: 1},
		{ID: 32, ParentID: 30, Name: "链路追踪", Path: "/observability/traces", Component: "common/FeaturePlaceholder", Icon: "Share", Sort: 2, Type: 2, Permission: "observability:traces", Status: 1},
		{ID: 38, ParentID: 0, Name: "告警中心", Path: "/alert-center", Component: "", Icon: "Bell", Sort: 4, Type: 1, Status: 1},
		{ID: 39, ParentID: 38, Name: "告警管理", Path: "/alert-rules", Component: "alert/RuleManage", Icon: "Histogram", Sort: 1, Type: 2, Permission: "alert:rule:list", Status: 1},
		{ID: 40, ParentID: 0, Name: "故障中心", Path: "/alerts", Component: "alert/AlertCenter", Icon: "OfficeBuilding", Sort: 3, Type: 2, Permission: "alert:event:list", Status: 1},
		{ID: 70, ParentID: 0, Name: "通知管理", Path: "/notifications", Component: "", Icon: "Message", Sort: 11, Type: 1, Status: 1},
		{ID: 71, ParentID: 70, Name: "通知对象", Path: "/notifications/channels", Component: "notification/ChannelManage", Icon: "Promotion", Sort: 1, Type: 2, Permission: "notification:channel:list", Status: 1},
		{ID: 72, ParentID: 70, Name: "通知模板", Path: "/notifications/templates", Component: "notification/TemplateManage", Icon: "Tickets", Sort: 2, Type: 2, Permission: "notification:template:list", Status: 1},
		{ID: 73, ParentID: 70, Name: "通知记录", Path: "/notifications/records", Component: "notification/RecordManage", Icon: "DocumentChecked", Sort: 3, Type: 2, Permission: "notification:record:list", Status: 1},
		{ID: 81, ParentID: 0, Name: "操作审计", Path: "/security/audit", Component: "security/AuditLog", Icon: "DocumentChecked", Sort: 90, Type: 2, Permission: "security:audit:list", Status: 1},
	}
	for _, item := range menus {
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&item).Error; err != nil {
			return err
		}
	}
	if err := db.Model(&models.Menu{}).Where("id = ?", 87).Updates(map[string]interface{}{
		"parent_id": 0, "name": "系统设置", "path": "/settings", "component": "", "icon": "Tools",
		"sort": 99, "type": 1, "permission": "", "status": 1,
	}).Error; err != nil {
		return err
	}
	if err := db.Model(&models.Menu{}).Where("id = ?", 86).Updates(map[string]interface{}{
		"parent_id": 87, "name": "Git 认证", "path": "/settings/git-credentials", "component": "system/GitCredentialManage",
		"icon": "Key", "sort": 3, "type": 2, "permission": "git-credential:list", "status": 1,
	}).Error; err != nil {
		return err
	}
	if err := db.Model(&models.Menu{}).Where("id = ?", 90).Updates(map[string]interface{}{
		"parent_id": 87, "name": "访问令牌", "path": "/settings/api-tokens", "component": "system/APITokenManage",
		"icon": "Key", "sort": 4, "type": 2, "permission": "api-token:list", "status": 1,
	}).Error; err != nil {
		return err
	}
	if err := db.Model(&models.Menu{}).Where("id = ?", 92).Updates(map[string]interface{}{
		"parent_id": 87, "name": "MCP 服务", "path": "/settings/mcp", "component": "system/MCPServiceManage",
		"icon": "Connection", "sort": 5, "type": 2, "permission": "mcp-service:list", "status": 1,
	}).Error; err != nil {
		return err
	}
	if err := db.Model(&models.Menu{}).Where("id = ?", 88).Updates(map[string]interface{}{"parent_id": 87, "name": "基础设置", "path": "/settings/basic", "component": "system/BasicSetting", "icon": "SetUp", "sort": 1, "type": 2, "permission": "system-setting:list", "status": 0}).Error; err != nil {
		return err
	}
	if err := db.Model(&models.Menu{}).Where("id = ?", 89).Updates(map[string]interface{}{"parent_id": 87, "name": "登录认证", "path": "/settings/authentication", "component": "system/AuthSetting", "icon": "Lock", "sort": 2, "type": 2, "permission": "system-setting:list", "status": 1}).Error; err != nil {
		return err
	}
	if err := db.Exec("INSERT OR IGNORE INTO role_menus (role_id, menu_id) SELECT role_id, 87 FROM role_menus WHERE menu_id = 86").Error; err != nil {
		return err
	}
	if err := db.Model(&models.Menu{}).Where("id = ?", 8).Updates(map[string]interface{}{
		"parent_id": 0, "name": "仪表盘", "path": "/dashboards", "component": "",
		"icon": "DataAnalysis", "sort": 1, "type": 1, "permission": "", "status": 1,
	}).Error; err != nil {
		return err
	}
	if err := db.Model(&models.Menu{}).Where("id = ?", 84).Updates(map[string]interface{}{
		"parent_id": 8, "name": "监控仪表盘", "path": "/dashboards/grafana",
		"component": "monitor/GrafanaDashboardManage", "icon": "DataAnalysis", "sort": 1,
		"type": 2, "permission": "dashboard:list", "status": 1,
	}).Error; err != nil {
		return err
	}
	if err := db.Model(&models.Menu{}).Where("id = ?", 85).Updates(map[string]interface{}{
		"parent_id": 8, "name": "日志仪表盘", "path": "/dashboards/sls",
		"component": "monitor/SLSDashboardManage", "icon": "TrendCharts", "sort": 2,
		"type": 2, "permission": "dashboard:list", "status": 1,
	}).Error; err != nil {
		return err
	}
	if err := db.Model(&models.Menu{}).Where("id = ?", 91).Updates(map[string]interface{}{
		"parent_id": 8, "name": "云监控2.0", "path": "/dashboards/cloud-monitor",
		"component": "monitor/CloudMonitorDashboardManage", "icon": "Cloudy", "sort": 3,
		"type": 2, "permission": "dashboard:list", "status": 1,
	}).Error; err != nil {
		return err
	}
	if err := db.Exec("INSERT OR IGNORE INTO role_menus (role_id, menu_id) SELECT role_id, 84 FROM role_menus WHERE menu_id = 8").Error; err != nil {
		return err
	}
	if err := db.Exec("INSERT OR IGNORE INTO role_menus (role_id, menu_id) SELECT role_id, 85 FROM role_menus WHERE menu_id = 8").Error; err != nil {
		return err
	}
	if err := db.Exec("INSERT OR IGNORE INTO role_menus (role_id, menu_id) SELECT role_id, 91 FROM role_menus WHERE menu_id = 8").Error; err != nil {
		return err
	}
	if err := db.Exec("INSERT OR IGNORE INTO role_menus (role_id, menu_id) SELECT DISTINCT role_id, 8 FROM role_menus WHERE menu_id IN (84, 85, 91)").Error; err != nil {
		return err
	}
	if err := db.Model(&models.Dashboard{}).Where("provider = '' OR provider IS NULL").Update("provider", "grafana").Error; err != nil {
		return err
	}
	if err := db.Model(&models.Menu{}).Where("id = ?", 38).Updates(map[string]interface{}{
		"parent_id": 0, "name": "告警中心", "path": "/alert-center", "component": "",
		"icon": "Bell", "sort": 4, "type": 1, "permission": "", "status": 1,
	}).Error; err != nil {
		return err
	}
	if err := db.Model(&models.Menu{}).Where("id = ?", 19).Updates(map[string]interface{}{
		"parent_id": 38, "name": "即时查询", "path": "/monitor/query", "component": "monitor/InstantQuery",
		"icon": "Search", "sort": 2, "type": 2, "status": 1,
	}).Error; err != nil {
		return err
	}
	if err := db.Model(&models.Menu{}).Where("id = ?", 39).Updates(map[string]interface{}{
		"parent_id": 38, "name": "告警管理", "path": "/alert-rules", "component": "alert/RuleManage",
		"icon": "Histogram", "sort": 1, "type": 2, "permission": "alert:rule:list", "status": 1,
	}).Error; err != nil {
		return err
	}
	if err := db.Exec("INSERT OR IGNORE INTO role_menus (role_id, menu_id) SELECT DISTINCT role_id, 38 FROM role_menus WHERE menu_id IN (19, 39)").Error; err != nil {
		return err
	}
	if err := ensureFaultCenters(db); err != nil {
		return err
	}
	if err := db.Exec("DELETE FROM role_menus WHERE menu_id IN (SELECT id FROM menus WHERE path = ? OR path LIKE ?)", "/assets", "/assets/%").Error; err != nil {
		return err
	}
	if err := db.Where("path = ? OR path LIKE ?", "/assets", "/assets/%").Delete(&models.Menu{}).Error; err != nil {
		return err
	}
	if err := db.Exec("DELETE FROM role_permissions WHERE permission_id IN (SELECT id FROM permissions WHERE code IN ?)", []string{"asset:list", "asset:topology"}).Error; err != nil {
		return err
	}
	if err := db.Where("code IN ?", []string{"asset:list", "asset:topology"}).Delete(&models.Permission{}).Error; err != nil {
		return err
	}
	if err := db.Exec("DELETE FROM role_menus WHERE menu_id IN (SELECT id FROM menus WHERE path = ? OR component = ?)", "/tunnels", "system/TunnelManage").Error; err != nil {
		return err
	}
	if err := db.Where("path = ? OR component = ?", "/tunnels", "system/TunnelManage").Delete(&models.Menu{}).Error; err != nil {
		return err
	}
	if err := db.Exec("DELETE FROM role_permissions WHERE permission_id IN (SELECT id FROM permissions WHERE code LIKE ?)", "tunnel:%").Error; err != nil {
		return err
	}
	if err := db.Where("code LIKE ?", "tunnel:%").Delete(&models.Permission{}).Error; err != nil {
		return err
	}
	if err := db.Exec("DELETE FROM role_menus WHERE menu_id IN (SELECT id FROM menus WHERE path = ? OR component = ?)", "/certificates/aliyun", "certificate/AliyunCertificate").Error; err != nil {
		return err
	}
	if err := db.Where("path = ? OR component = ?", "/certificates/aliyun", "certificate/AliyunCertificate").Delete(&models.Menu{}).Error; err != nil {
		return err
	}
	if err := db.Model(&models.Menu{}).Where("path = ?", "/certificates").Update("name", "证书管理").Error; err != nil {
		return err
	}
	if err := db.Model(&models.Menu{}).Where("path = ?", "/certificates/assets").Update("name", "证书资产").Error; err != nil {
		return err
	}
	if err := db.Model(&models.Menu{}).Where("path = ?", "/security/audit").Updates(map[string]interface{}{
		"parent_id": 0, "name": "操作审计", "component": "security/AuditLog", "permission": "security:audit:list", "sort": 90, "status": 1,
	}).Error; err != nil {
		return err
	}
	if err := db.Exec("INSERT OR IGNORE INTO role_menus (role_id, menu_id) SELECT role_id, 81 FROM role_menus WHERE menu_id = 80").Error; err != nil {
		return err
	}
	if err := db.Exec("DELETE FROM role_menus WHERE menu_id = ?", 80).Error; err != nil {
		return err
	}
	if err := db.Where("id = ?", 80).Delete(&models.Menu{}).Error; err != nil {
		return err
	}
	if err := db.Model(&models.Menu{}).Where("id = ?", 1).Updates(map[string]interface{}{"name": "用户管理", "icon": "User", "sort": 98}).Error; err != nil {
		return err
	}
	if err := db.Model(&models.Menu{}).Where("id = ?", 2).Update("name", "用户列表").Error; err != nil {
		return err
	}
	if err := db.Model(&models.Menu{}).Where("id = ?", 3).Update("name", "角色列表").Error; err != nil {
		return err
	}
	if err := db.Model(&models.Menu{}).Where("id = ?", 4).Update("name", "权限列表").Error; err != nil {
		return err
	}
	if err := db.Model(&models.Menu{}).Where("id = ?", 5).Update("name", "菜单列表").Error; err != nil {
		return err
	}
	if err := db.Model(&models.Menu{}).Where("id = ?", 40).Updates(map[string]interface{}{
		"parent_id": 0, "name": "故障中心", "path": "/alerts", "component": "alert/AlertCenter",
		"icon": "OfficeBuilding", "sort": 3, "type": 2, "permission": "alert:event:list", "status": 1,
	}).Error; err != nil {
		return err
	}
	if err := db.Exec("DELETE FROM role_menus WHERE menu_id = ?", 41).Error; err != nil {
		return err
	}
	if err := db.Where("id = ?", 41).Delete(&models.Menu{}).Error; err != nil {
		return err
	}
	if err := db.Model(&models.Permission{}).Where("code = ?", "alert:event:list").Update("name", "查看故障告警").Error; err != nil {
		return err
	}
	if err := db.Model(&models.Permission{}).Where("code = ?", "certificate:list").Update("name", "证书轮转记录").Error; err != nil {
		return err
	}
	if err := db.Model(&models.Permission{}).Where("code = ?", "certificate:asset:list").Update("name", "证书资产").Error; err != nil {
		return err
	}
	if err := db.Model(&models.Permission{}).Where("code = ?", "certificate:asset:create").Update("name", "上传证书").Error; err != nil {
		return err
	}
	if err := db.Model(&models.Permission{}).Where("code = ?", "certificate:asset:delete").Update("name", "删除证书").Error; err != nil {
		return err
	}
	if err := db.Model(&models.Menu{}).Where("path = ?", "/certificates/domains").Updates(map[string]interface{}{
		"name":      "域名管理",
		"component": "certificate/DomainManage",
		"icon":      "Link",
		"sort":      2,
		"status":    1,
	}).Error; err != nil {
		return err
	}
	if err := db.Model(&models.Permission{}).Where("code = ?", "certificate:domain:list").Update("name", "域名管理").Error; err != nil {
		return err
	}
	if err := db.Model(&models.Permission{}).Where("code = ?", "certificate:domain:edit").Update("name", "维护域名").Error; err != nil {
		return err
	}
	if err := db.Model(&models.Menu{}).Where("id = ?", 15).Updates(map[string]interface{}{
		"parent_id": 0,
		"name":      "自动化中心",
		"path":      "/automation",
		"component": "",
		"icon":      "Operation",
		"sort":      7,
		"type":      1,
		"status":    1,
	}).Error; err != nil {
		return err
	}
	if err := db.Model(&models.Menu{}).Where("id = ?", 24).Updates(map[string]interface{}{
		"parent_id": 87,
		"name":      "节点管理",
		"path":      "/nodes",
		"component": "certificate/ExecutorManage",
		"icon":      "Monitor",
		"sort":      5,
		"type":      2,
		"status":    1,
	}).Error; err != nil {
		return err
	}
	if err := db.Exec("INSERT OR IGNORE INTO role_menus (role_id, menu_id) SELECT role_id, 87 FROM role_menus WHERE menu_id = 24").Error; err != nil {
		return err
	}
	if err := db.Exec("INSERT OR IGNORE INTO role_menus (role_id, menu_id) SELECT role_id, 12 FROM role_menus WHERE menu_id = 25").Error; err != nil {
		return err
	}
	if err := db.Exec("DELETE FROM role_menus WHERE menu_id = 25").Error; err != nil {
		return err
	}
	if err := db.Delete(&models.Menu{}, 25).Error; err != nil {
		return err
	}
	if err := db.Model(&models.Menu{}).Where("id = ?", 12).Updates(map[string]interface{}{
		"parent_id":  15,
		"name":       "脚本库",
		"path":       "/automation/scripts",
		"component":  "automation/ScriptLibrary",
		"icon":       "Document",
		"sort":       2,
		"type":       2,
		"permission": "",
		"status":     1,
	}).Error; err != nil {
		return err
	}
	if err := db.Model(&models.Menu{}).Where("id = ?", 13).Updates(map[string]interface{}{
		"parent_id": 15, "name": "自动化任务", "path": "/automation/tasks",
		"component": "automation/AutomationTaskManage", "icon": "SetUp", "sort": 3,
		"type": 2, "permission": "automation-task:list", "status": 1,
	}).Error; err != nil {
		return err
	}
	if err := db.Model(&models.Menu{}).Where("id = ?", 18).Updates(map[string]interface{}{"parent_id": 15, "path": "/automation/scheduled-tasks", "sort": 4}).Error; err != nil {
		return err
	}
	if err := db.Model(&models.Menu{}).Where("id = ?", 14).Updates(map[string]interface{}{"parent_id": 15, "name": "执行记录", "path": "/automation/records", "sort": 5}).Error; err != nil {
		return err
	}
	if err := db.Model(&models.Menu{}).Where("id = ?", 9).Updates(map[string]interface{}{"name": "拨测中心", "sort": 6}).Error; err != nil {
		return err
	}
	if err := db.Model(&models.Menu{}).Where("id = ?", 11).Updates(map[string]interface{}{"name": "拨测任务", "sort": 2}).Error; err != nil {
		return err
	}
	if err := db.Exec("DELETE FROM role_menus WHERE menu_id = ?", 28).Error; err != nil {
		return err
	}
	if err := db.Where("id = ? OR path = ? OR component = ?", 28, "/certificates/deploy-scripts", "certificate/DeployScriptManage").Delete(&models.Menu{}).Error; err != nil {
		return err
	}
	if err := db.Exec("INSERT OR IGNORE INTO role_menus (role_id, menu_id) SELECT DISTINCT role_id, 15 FROM role_menus WHERE menu_id IN (12, 13, 14, 18)").Error; err != nil {
		return err
	}
	if err := db.Exec("DELETE FROM role_menus WHERE menu_id = 26").Error; err != nil {
		return err
	}
	if err := db.Delete(&models.Menu{}, 26).Error; err != nil {
		return err
	}
	if err := db.Exec("DELETE FROM role_menus WHERE menu_id = 27").Error; err != nil {
		return err
	}
	if err := db.Where("id = ? OR path = ? OR component = ?", 27, "/inspection/scripts", "inspection/ProbeScriptManage").Delete(&models.Menu{}).Error; err != nil {
		return err
	}
	if err := db.Exec("DELETE FROM role_menus WHERE menu_id IN (60, 61)").Error; err != nil {
		return err
	}
	if err := db.Where("id IN ? OR path LIKE ?", []uint{60, 61}, "/changes%").Delete(&models.Menu{}).Error; err != nil {
		return err
	}
	if err := db.Exec("DELETE FROM role_permissions WHERE permission_id IN (SELECT id FROM permissions WHERE code = ?)", "change:list").Error; err != nil {
		return err
	}
	if err := db.Where("code = ?", "change:list").Delete(&models.Permission{}).Error; err != nil {
		return err
	}

	if err := db.Model(&models.Menu{}).Where("id = ?", 20).Update("sort", 9).Error; err != nil {
		return err
	}
	if err := db.Model(&models.Menu{}).Where("id = ?", 30).Update("sort", 5).Error; err != nil {
		return err
	}
	if err := db.Model(&models.Menu{}).Where("id = ?", 6).Update("sort", 12).Error; err != nil {
		return err
	}
	if err := db.Model(&models.Menu{}).Where("id = ?", 70).Update("sort", 11).Error; err != nil {
		return err
	}
	if err := db.Model(&models.Permission{}).Where("code = ?", "script:list").Update("name", "脚本列表").Error; err != nil {
		return err
	}
	if err := db.Model(&models.Permission{}).Where("code = ?", "inspection:list").Update("name", "拨测任务列表").Error; err != nil {
		return err
	}
	if err := db.Model(&models.Permission{}).Where("code = ?", "inspection:create").Update("name", "创建拨测任务").Error; err != nil {
		return err
	}
	if err := db.Model(&models.Permission{}).Where("code = ?", "inspection:edit").Update("name", "编辑拨测任务").Error; err != nil {
		return err
	}
	if err := db.Model(&models.Permission{}).Where("code = ?", "inspection:delete").Update("name", "删除拨测任务").Error; err != nil {
		return err
	}
	if err := db.Model(&models.Permission{}).Where("code = ?", "inspection:run").Update("name", "执行拨测").Error; err != nil {
		return err
	}
	if err := db.Model(&models.Permission{}).Where("code = ?", "certificate:executor:list").Update("name", "节点列表").Error; err != nil {
		return err
	}
	if err := db.Model(&models.Permission{}).Where("code = ?", "certificate:executor:edit").Update("name", "维护节点").Error; err != nil {
		return err
	}
	if err := db.Exec("DELETE FROM role_permissions WHERE permission_id IN (SELECT id FROM permissions WHERE code IN ?)", []string{"certificate:deploy-script:list", "certificate:deploy-script:edit", "certificate:deploy:linux"}).Error; err != nil {
		return err
	}
	if err := db.Where("code IN ?", []string{"certificate:deploy-script:list", "certificate:deploy-script:edit", "certificate:deploy:linux"}).Delete(&models.Permission{}).Error; err != nil {
		return err
	}

	var adminRole models.Role
	if err := db.Where("code = ?", "admin").First(&adminRole).Error; err != nil {
		return err
	}

	permissionCodes := make([]string, 0, len(permissions))
	for _, item := range permissions {
		permissionCodes = append(permissionCodes, item.Code)
	}
	var newPerms []models.Permission
	if err := db.Where("code IN ?", permissionCodes).Find(&newPerms).Error; err != nil {
		return err
	}
	for i := range newPerms {
		if err := db.Exec("INSERT OR IGNORE INTO role_permissions (role_id, permission_id) VALUES (?, ?)", adminRole.ID, newPerms[i].ID).Error; err != nil {
			return err
		}
	}

	menuPaths := make([]string, 0, len(menus))
	for _, item := range menus {
		menuPaths = append(menuPaths, item.Path)
	}
	var newMenus []models.Menu
	if err := db.Where("path IN ?", menuPaths).Find(&newMenus).Error; err != nil {
		return err
	}
	for i := range newMenus {
		if err := db.Exec("INSERT OR IGNORE INTO role_menus (role_id, menu_id) VALUES (?, ?)", adminRole.ID, newMenus[i].ID).Error; err != nil {
			return err
		}
	}

	if err := db.Exec("INSERT OR IGNORE INTO role_menus (role_id, menu_id) SELECT role_id, 7 FROM role_menus WHERE menu_id IN (82, 83)").Error; err != nil {
		return err
	}
	if err := db.Exec("DELETE FROM role_menus WHERE menu_id IN (82, 83)").Error; err != nil {
		return err
	}
	if err := db.Where("id IN ? OR path = ? OR component = ?", []uint{82, 83}, "/integrations/workspaces", "monitor/DatasourceGroupManage").Delete(&models.Menu{}).Error; err != nil {
		return err
	}
	if err := db.Model(&models.Menu{}).Where("id = ?", 7).Updates(map[string]interface{}{
		"parent_id":  0,
		"name":       "接入中心",
		"path":       "/integrations/sources",
		"component":  "monitor/IntegrationManage",
		"icon":       "Connection",
		"sort":       10,
		"type":       2,
		"permission": "datasource:list",
		"status":     1,
	}).Error; err != nil {
		return err
	}
	if err := db.Exec("DELETE FROM role_menus WHERE menu_id = ?", 16).Error; err != nil {
		return err
	}
	if err := db.Where("id = ? OR path = ? OR component = ?", 16, "/workspaces", "system/WorkspaceManage").Delete(&models.Menu{}).Error; err != nil {
		return err
	}
	if err := db.Exec("DELETE FROM role_permissions WHERE permission_id IN (SELECT id FROM permissions WHERE code LIKE ?)", "workspace:%").Error; err != nil {
		return err
	}
	if err := db.Where("code LIKE ?", "workspace:%").Delete(&models.Permission{}).Error; err != nil {
		return err
	}
	if err := db.Model(&models.Permission{}).Where("code = ?", "datasource:list").Update("name", "接入列表").Error; err != nil {
		return err
	}
	if err := db.Model(&models.Permission{}).Where("code = ?", "datasource:create").Update("name", "创建接入").Error; err != nil {
		return err
	}
	if err := db.Model(&models.Permission{}).Where("code = ?", "datasource:edit").Update("name", "编辑接入").Error; err != nil {
		return err
	}
	if err := db.Model(&models.Permission{}).Where("code = ?", "datasource:delete").Update("name", "删除接入").Error; err != nil {
		return err
	}
	if err := db.Model(&models.Menu{}).Where("path = ?", "/notifications").Updates(map[string]interface{}{
		"name":   "通知管理",
		"icon":   "Message",
		"sort":   11,
		"status": 1,
	}).Error; err != nil {
		return err
	}
	if err := db.Model(&models.Menu{}).Where("path = ?", "/notifications/channels").Updates(map[string]interface{}{
		"name":      "通知对象",
		"component": "notification/ChannelManage",
		"icon":      "Promotion",
		"sort":      1,
		"status":    1,
	}).Error; err != nil {
		return err
	}
	if err := db.Model(&models.Permission{}).Where("code = ?", "notification:channel:list").Update("name", "通知对象").Error; err != nil {
		return err
	}
	if err := db.Model(&models.Permission{}).Where("code = ?", "notification:channel:edit").Update("name", "维护通知对象").Error; err != nil {
		return err
	}
	if err := db.Model(&models.Menu{}).Where("path = ?", "/notifications/templates").Updates(map[string]interface{}{
		"name":      "通知模板",
		"component": "notification/TemplateManage",
		"icon":      "Tickets",
		"sort":      2,
		"status":    1,
	}).Error; err != nil {
		return err
	}
	if err := db.Model(&models.Permission{}).Where("code = ?", "notification:template:list").Update("name", "通知模板").Error; err != nil {
		return err
	}
	if err := db.Model(&models.Permission{}).Where("code = ?", "notification:template:edit").Update("name", "维护通知模板").Error; err != nil {
		return err
	}
	if err := db.Model(&models.Menu{}).Where("path = ?", "/notifications/records").Updates(map[string]interface{}{
		"name":      "通知记录",
		"component": "notification/RecordManage",
		"icon":      "DocumentChecked",
		"sort":      3,
		"status":    1,
	}).Error; err != nil {
		return err
	}
	if err := db.Model(&models.Permission{}).Where("code = ?", "notification:record:list").Update("name", "通知记录").Error; err != nil {
		return err
	}

	var datasources []models.Datasource
	if err := db.Where("category = '' OR category IS NULL").Find(&datasources).Error; err != nil {
		return err
	}
	for i := range datasources {
		if err := db.Model(&datasources[i]).Update("category", datasourceCategoryByType(datasources[i].Type)).Error; err != nil {
			return err
		}
	}
	return restoreMenuSorts(db, menuSorts)
}

type menuSortSnapshot struct {
	ID       uint
	ParentID uint
	Sort     int
}

func snapshotMenuSorts(db *gorm.DB) ([]menuSortSnapshot, error) {
	var sorts []menuSortSnapshot
	if err := db.Model(&models.Menu{}).Select("id", "parent_id", "sort").Scan(&sorts).Error; err != nil {
		return nil, err
	}

	// Correct the short-lived reversed defaults once. Any other administrator
	// configuration is preserved exactly as stored.
	values := make(map[uint]int, len(sorts))
	for _, item := range sorts {
		values[item.ID] = item.Sort
	}
	if values[1] == 99 && values[87] == 98 {
		for i := range sorts {
			switch sorts[i].ID {
			case 1:
				sorts[i].Sort = 98
			case 87:
				sorts[i].Sort = 99
			}
		}
	}
	for i := range sorts {
		if sorts[i].ID == 81 && sorts[i].ParentID == 80 {
			sorts[i].Sort = 90
		}
		_, hasAPITokenMenu := values[90]
		if sorts[i].ID == 24 && ((sorts[i].ParentID == 0) || (!hasAPITokenMenu && sorts[i].ParentID == 87 && sorts[i].Sort == 4)) {
			sorts[i].Sort = 5
		}
	}
	return sorts, nil
}

func restoreMenuSorts(db *gorm.DB, sorts []menuSortSnapshot) error {
	for _, item := range sorts {
		if err := db.Model(&models.Menu{}).Where("id = ?", item.ID).Update("sort", item.Sort).Error; err != nil {
			return err
		}
	}
	return nil
}

func ensureFaultCenters(db *gorm.DB) error {
	var defaultCenter models.FaultCenter
	result := db.Where("name = ?", "默认故障中心").Limit(1).Find(&defaultCenter)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		defaultCenter = models.FaultCenter{Name: "默认故障中心", Description: "接收未单独分配的拨测告警", Status: 1}
		if err := db.Create(&defaultCenter).Error; err != nil {
			return err
		}
	}

	var items []models.InspectionItem
	if err := db.Find(&items).Error; err != nil {
		return err
	}
	for i := range items {
		centerID := items[i].FaultCenterID
		if centerID == 0 {
			centerID = defaultCenter.ID
			if err := db.Model(&items[i]).Update("fault_center_id", centerID).Error; err != nil {
				return err
			}
		}
		if err := db.Model(&models.AlertInstance{}).Where("source = ? AND source_id = ? AND fault_center_id = 0", "probe-task", items[i].ID).Update("fault_center_id", centerID).Error; err != nil {
			return err
		}
	}
	return nil
}

func datasourceCategoryByType(dsType string) string {
	switch dsType {
	case "prometheus", "loki", "elasticsearch", "jaeger", "opentelemetry", "sls", "cls", "trace":
		return "observability"
	case "mysql", "clickhouse", "postgresql", "redis", "influxdb":
		return "database"
	case "aliyun_ecs", "aliyun_cdn", "aliyun_dcdn", "aliyun_slb", "aliyun_alb", "aliyun_nlb":
		return "cloud"
	case "k8s":
		return "kubernetes"
	case "http", "grpc", "tcp", "icmp", "dns":
		return "network"
	case "kafka", "rabbitmq", "nacos", "consul":
		return "middleware"
	default:
		return "other"
	}
}
