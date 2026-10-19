-- Existing installations only.
-- This migration records that all schema/data migrations bundled before
-- 2026-09-01 have already been applied. It does not alter business data.

CREATE TABLE IF NOT EXISTS `schema_migrations` (
  `version` varchar(64) NOT NULL,
  `applied_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`version`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT IGNORE INTO `schema_migrations` (`version`, `applied_at`)
VALUES ('2026090101', CURRENT_TIMESTAMP(3));
