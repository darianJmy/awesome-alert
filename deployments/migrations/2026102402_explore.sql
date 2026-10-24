-- Unified Explore menu; retain legacy rows/bindings for rollback, hide stock entries.
BEGIN IMMEDIATE;
INSERT INTO permissions (name,code,status) SELECT 'SQL 查询','query:sql',1 WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code='query:sql');
INSERT OR IGNORE INTO role_permissions (role_id,permission_id) SELECT r.id,p.id FROM roles r,permissions p WHERE r.code='admin' AND p.code='query:sql';
INSERT OR IGNORE INTO role_menus (role_id,menu_id) SELECT role_id,19 FROM role_menus WHERE menu_id IN (30,31,32) AND EXISTS (SELECT 1 FROM menus WHERE id=19 AND path='/monitor/query');
UPDATE menus SET parent_id=0,name='即时查询',component='monitor/Explore',sort=2 WHERE id=19 AND path='/monitor/query' AND component IN ('monitor/InstantQuery','monitor/Explore');
UPDATE menus SET status=0 WHERE (id=30 AND path='/observability') OR (id=31 AND path='/observability/logs') OR (id=32 AND path='/observability/traces');
UPDATE menus SET sort=3 WHERE id=38 AND path='/alert-center' AND parent_id=0;
UPDATE menus SET sort=4 WHERE id=40 AND path='/alerts' AND parent_id=0;
COMMIT;
