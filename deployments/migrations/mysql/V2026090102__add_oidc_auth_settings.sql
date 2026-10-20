-- Normalize the GORM-generated OIDC column names to the API's oidc_* names.
-- CHANGE COLUMN preserves the values already stored in existing installations.

ALTER TABLE `auth_settings`
  CHANGE COLUMN `o_id_c_enabled` `oidc_enabled` tinyint(1) NOT NULL DEFAULT 0,
  CHANGE COLUMN `o_id_c_issuer_url` `oidc_issuer_url` varchar(500) NULL,
  CHANGE COLUMN `o_id_c_client_id` `oidc_client_id` varchar(255) NULL,
  CHANGE COLUMN `o_id_c_client_secret` `oidc_client_secret` longtext NULL,
  CHANGE COLUMN `o_id_c_redirect_url` `oidc_redirect_url` varchar(500) NULL,
  CHANGE COLUMN `o_id_c_scopes` `oidc_scopes` varchar(500) NULL,
  CHANGE COLUMN `o_id_c_username_claim` `oidc_username_claim` varchar(100) NULL;

INSERT INTO `schema_migrations` (`version`, `applied_at`)
VALUES ('2026090102', CURRENT_TIMESTAMP(3));
