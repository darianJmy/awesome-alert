-- Make every AI conversation belong to one platform user. Existing rows are
-- assigned to the earliest user so pre-upgrade history remains recoverable.

ALTER TABLE `chat_conversations`
  ADD COLUMN `user_id` bigint unsigned NOT NULL DEFAULT 0 AFTER `id`;

UPDATE `chat_conversations`
SET `user_id` = COALESCE((SELECT MIN(`id`) FROM `users`), 0)
WHERE `user_id` = 0;

CREATE INDEX `idx_chat_user_updated`
  ON `chat_conversations` (`user_id`, `updated_at`);

INSERT INTO `schema_migrations` (`version`, `applied_at`)
VALUES ('2026090203', CURRENT_TIMESTAMP(3));
