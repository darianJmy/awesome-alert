-- Retire certificate management. Shared SSH/Agent nodes and SSL probes remain.
-- Keep certificate data and alert history for audit; do not drop business tables.
START TRANSACTION;
SET @retired_at = CURRENT_TIMESTAMP(3);

UPDATE certificate_domains SET monitor_enabled = 0;
UPDATE automation_tasks SET status = 0 WHERE runtime_profile = 'certificate_deploy';

INSERT INTO alert_events (alert_id,event_type,from_status,to_status,from_severity,to_severity,reason,value,labels,evaluation_status,created_at)
SELECT id,'state_changed',status,'terminated',severity,severity,'certificate_module_removed','',labels,last_probe_status,@retired_at
FROM alert_instances WHERE source='certificate-domain' AND status IN ('pending','firing','recovering');
UPDATE alert_instances SET status='terminated',resolved_at=@retired_at,state_changed_at=@retired_at,last_event_at=@retired_at,updated_at=@retired_at,
notification_status='terminated',notification_reason='证书管理模块已移除',next_notification_at=NULL
WHERE source='certificate-domain' AND status IN ('pending','firing','recovering');

DELETE rm FROM role_menus rm JOIN menus m ON m.id=rm.menu_id
WHERE m.path='/certificates' OR m.path LIKE '/certificates/%' OR m.component IN ('certificate/CertificateManage','certificate/DomainManage','certificate/DeployScriptManage');
DELETE FROM menus WHERE path='/certificates' OR path LIKE '/certificates/%' OR component IN ('certificate/CertificateManage','certificate/DomainManage','certificate/DeployScriptManage');
DELETE rp FROM role_permissions rp JOIN permissions p ON p.id=rp.permission_id
WHERE p.code='certificate:list' OR p.code LIKE 'certificate:asset:%' OR p.code LIKE 'certificate:domain:%' OR p.code LIKE 'certificate:deploy%';
DELETE FROM permissions WHERE code='certificate:list' OR code LIKE 'certificate:asset:%' OR code LIKE 'certificate:domain:%' OR code LIKE 'certificate:deploy%';

INSERT INTO schema_migrations (version,applied_at) VALUES ('2026090601',@retired_at)
ON DUPLICATE KEY UPDATE version=VALUES(version);
COMMIT;
