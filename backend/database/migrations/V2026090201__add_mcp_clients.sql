-- Add centrally managed MCP client identities. API tokens remain the only
-- credential; one token can be linked to at most one MCP client.

CREATE TABLE `mcp_clients` (
  `id` varchar(32) NOT NULL,
  `name` varchar(100) NOT NULL,
  `client_type` varchar(32) NOT NULL,
  `api_token_id` varchar(32) NOT NULL,
  `client_version` varchar(64) DEFAULT '',
  `protocol_version` varchar(64) DEFAULT '',
  `description` varchar(500) DEFAULT '',
  `status` bigint NOT NULL DEFAULT 1,
  `last_seen_at` datetime(3) DEFAULT NULL,
  `last_ip` varchar(64) DEFAULT '',
  `last_user_agent` varchar(500) DEFAULT '',
  `created_by` bigint unsigned NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_mcp_clients_name` (`name`),
  UNIQUE KEY `idx_mcp_clients_api_token_id` (`api_token_id`),
  KEY `idx_mcp_clients_client_type` (`client_type`),
  KEY `idx_mcp_clients_status` (`status`),
  KEY `idx_mcp_clients_last_seen_at` (`last_seen_at`),
  KEY `idx_mcp_clients_created_by` (`created_by`),
  CONSTRAINT `fk_mcp_clients_api_token` FOREIGN KEY (`api_token_id`) REFERENCES `api_access_tokens` (`id`) ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO `permissions` (`name`, `code`, `description`, `status`, `created_at`, `updated_at`)
SELECT 'MCP 客户端列表', 'mcp-client:list', '', 1, CURRENT_TIMESTAMP(3), CURRENT_TIMESTAMP(3)
WHERE NOT EXISTS (SELECT 1 FROM `permissions` WHERE `code` = 'mcp-client:list');

INSERT INTO `permissions` (`name`, `code`, `description`, `status`, `created_at`, `updated_at`)
SELECT '维护 MCP 客户端', 'mcp-client:edit', '', 1, CURRENT_TIMESTAMP(3), CURRENT_TIMESTAMP(3)
WHERE NOT EXISTS (SELECT 1 FROM `permissions` WHERE `code` = 'mcp-client:edit');

INSERT IGNORE INTO `role_permissions` (`role_id`, `permission_id`)
SELECT r.id, p.id FROM `roles` r JOIN `permissions` p ON p.code IN ('mcp-client:list', 'mcp-client:edit')
WHERE r.code = 'admin';

UPDATE `menus` SET `sort` = 6, `updated_at` = CURRENT_TIMESTAMP(3)
WHERE `path` = '/nodes' AND `parent_id` = (SELECT id FROM (SELECT id FROM `menus` WHERE `path` = '/settings' LIMIT 1) AS settings_menu);

INSERT INTO `menus` (`parent_id`, `name`, `path`, `component`, `icon`, `sort`, `type`, `permission`, `status`, `created_at`, `updated_at`)
SELECT settings.id, 'MCP 客户端', '/settings/mcp-clients', 'system/MCPClientManage', 'Connection', 5, 2, 'mcp-client:list', 1, CURRENT_TIMESTAMP(3), CURRENT_TIMESTAMP(3)
FROM `menus` settings
WHERE settings.path = '/settings'
  AND NOT EXISTS (SELECT 1 FROM `menus` WHERE `path` = '/settings/mcp-clients')
LIMIT 1;

INSERT IGNORE INTO `role_menus` (`role_id`, `menu_id`)
SELECT r.id, m.id FROM `roles` r JOIN `menus` m ON m.path = '/settings/mcp-clients'
WHERE r.code = 'admin';

-- Retire the old Skill management product entry without deleting stored
-- Markdown rows, so existing content remains recoverable if needed.
DELETE rm FROM `role_menus` rm JOIN `menus` m ON m.id = rm.menu_id
WHERE m.path = '/system/ai-skills' OR m.component = 'system/AISkillManage';

DELETE FROM `menus` WHERE `path` = '/system/ai-skills' OR `component` = 'system/AISkillManage';

DELETE rp FROM `role_permissions` rp JOIN `permissions` p ON p.id = rp.permission_id
WHERE p.code IN ('ai_skill:list', 'ai_skill:create', 'ai_skill:edit', 'ai_skill:delete');

DELETE FROM `permissions` WHERE `code` IN ('ai_skill:list', 'ai_skill:create', 'ai_skill:edit', 'ai_skill:delete');

INSERT INTO `schema_migrations` (`version`, `applied_at`)
VALUES ('2026090201', CURRENT_TIMESTAMP(3));
