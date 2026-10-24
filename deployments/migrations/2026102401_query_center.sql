-- Optional, idempotent menu update for existing SQLite installations.
-- Preserves menu IDs, role bindings and existing URLs. Only changes stock entries.
BEGIN IMMEDIATE;
UPDATE menus SET name = '查询中心' WHERE id = 30 AND path = '/observability' AND name = '可观测中心';
UPDATE menus SET parent_id = 30, name = '指标查询' WHERE id = 19 AND path = '/monitor/query' AND parent_id = 38 AND name = '即时查询';
UPDATE menus SET component = 'monitor/LogQuery' WHERE path = '/observability/logs' AND component = 'common/FeaturePlaceholder';
COMMIT;
