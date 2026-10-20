-- Add probe rule groups while preserving task IDs, configuration and history.
CREATE TABLE IF NOT EXISTS inspection_rule_groups (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  name VARCHAR(128) NOT NULL,
  description VARCHAR(500) DEFAULT NULL,
  created_at DATETIME(3) DEFAULT NULL,
  updated_at DATETIME(3) DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY idx_inspection_rule_groups_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
INSERT INTO inspection_rule_groups (id, name, description, created_at, updated_at)
VALUES (1, '默认规则组', '默认拨测任务规则组', NOW(3), NOW(3))
ON DUPLICATE KEY UPDATE id = id;

SET @inspection_group_sql = IF(
  EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'inspection_items' AND column_name = 'group_id'),
  'SELECT 1',
  'ALTER TABLE inspection_items ADD COLUMN group_id BIGINT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX idx_inspection_items_group_id (group_id)'
);
PREPARE inspection_group_stmt FROM @inspection_group_sql;
EXECUTE inspection_group_stmt;
DEALLOCATE PREPARE inspection_group_stmt;
SET @inspection_group_sql = IF(
  EXISTS(SELECT 1 FROM information_schema.table_constraints WHERE constraint_schema = DATABASE() AND table_name = 'inspection_items' AND constraint_name = 'fk_inspection_items_group'),
  'SELECT 1',
  'ALTER TABLE inspection_items ADD CONSTRAINT fk_inspection_items_group FOREIGN KEY (group_id) REFERENCES inspection_rule_groups (id) ON UPDATE CASCADE ON DELETE RESTRICT'
);
PREPARE inspection_group_stmt FROM @inspection_group_sql;
EXECUTE inspection_group_stmt;
DEALLOCATE PREPARE inspection_group_stmt;
INSERT IGNORE INTO schema_migrations (version, applied_at) VALUES ('2026090701', NOW(3));
