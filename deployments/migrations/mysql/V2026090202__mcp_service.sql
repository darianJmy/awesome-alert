-- Promote MCP client inventory into a real MCP Streamable HTTP service.

ALTER TABLE `mcp_clients`
  ADD COLUMN `reported_name` varchar(100) DEFAULT '' AFTER `api_token_id`;

UPDATE `permissions`
SET `name` = 'MCP 服务', `code` = 'mcp-service:list', `updated_at` = CURRENT_TIMESTAMP(3)
WHERE `code` = 'mcp-client:list';

UPDATE `permissions`
SET `name` = '维护 MCP 服务', `code` = 'mcp-service:edit', `updated_at` = CURRENT_TIMESTAMP(3)
WHERE `code` = 'mcp-client:edit';

INSERT INTO `permissions` (`name`, `code`, `description`, `status`, `created_at`, `updated_at`)
SELECT 'MCP 服务', 'mcp-service:list', '', 1, CURRENT_TIMESTAMP(3), CURRENT_TIMESTAMP(3)
WHERE NOT EXISTS (SELECT 1 FROM `permissions` WHERE `code` = 'mcp-service:list');

INSERT INTO `permissions` (`name`, `code`, `description`, `status`, `created_at`, `updated_at`)
SELECT '维护 MCP 服务', 'mcp-service:edit', '', 1, CURRENT_TIMESTAMP(3), CURRENT_TIMESTAMP(3)
WHERE NOT EXISTS (SELECT 1 FROM `permissions` WHERE `code` = 'mcp-service:edit');

INSERT IGNORE INTO `role_permissions` (`role_id`, `permission_id`)
SELECT r.id, p.id
FROM `roles` r
JOIN `permissions` p ON p.code IN ('mcp-service:list', 'mcp-service:edit')
WHERE r.code = 'admin';

UPDATE `menus`
SET `name` = 'MCP 服务',
    `path` = '/settings/mcp',
    `component` = 'system/MCPServiceManage',
    `permission` = 'mcp-service:list',
    `updated_at` = CURRENT_TIMESTAMP(3)
WHERE `id` = 92 OR `path` = '/settings/mcp-clients';

INSERT INTO `schema_migrations` (`version`, `applied_at`)
VALUES ('2026090202', CURRENT_TIMESTAMP(3));
