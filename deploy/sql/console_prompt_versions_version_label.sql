-- Add semver-style version_label (e.g. 1.2.3). Safe to re-run: checks column existence via procedure pattern omitted; run once.
ALTER TABLE console_prompt_versions
  ADD COLUMN version_label VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'semver label without v prefix' AFTER version_num;

UPDATE console_prompt_versions
SET version_label = CAST(version_num AS CHAR)
WHERE version_label = '' OR version_label IS NULL;

ALTER TABLE console_prompt_versions
  ADD UNIQUE KEY uk_console_prompt_ver_label (prompt_id, version_label);
