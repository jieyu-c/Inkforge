-- Subset for goctl model generation only (goctl mysql ddl does not parse FK).
CREATE TABLE IF NOT EXISTS console_prompts (
  id CHAR(36) NOT NULL,
  tenant_id CHAR(36) NOT NULL COMMENT 'account isolation domain',
  ns_id CHAR(36) NOT NULL,
  prompt_key VARCHAR(128) NOT NULL,
  title VARCHAR(255) NULL,
  tags JSON NULL,
  owner_user_id CHAR(36) NULL,
  draft_body MEDIUMTEXT NOT NULL,
  draft_schema JSON NULL COMMENT 'JSON array of variable definitions',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_console_prompt_ns_key (ns_id, prompt_key),
  KEY idx_console_prompt_tenant_ns (tenant_id, ns_id),
  KEY idx_console_prompt_updated (ns_id, updated_at)
);

CREATE TABLE IF NOT EXISTS console_prompt_versions (
  id CHAR(36) NOT NULL,
  tenant_id CHAR(36) NOT NULL,
  ns_id CHAR(36) NOT NULL,
  prompt_id CHAR(36) NOT NULL,
  version_num BIGINT UNSIGNED NOT NULL COMMENT 'monotonic per prompt_id',
  version_label VARCHAR(64) NOT NULL COMMENT 'semver label e.g. 1.2.3',
  body MEDIUMTEXT NOT NULL,
  schema_json JSON NULL,
  change_note TEXT NULL,
  created_by_user_id CHAR(36) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_console_prompt_ver (prompt_id, version_num),
  UNIQUE KEY uk_console_prompt_ver_label (prompt_id, version_label),
  KEY idx_console_prompt_ver_ns (tenant_id, ns_id)
);

CREATE TABLE IF NOT EXISTS console_prompt_channel_pointers (
  id CHAR(36) NOT NULL,
  tenant_id CHAR(36) NOT NULL,
  ns_id CHAR(36) NOT NULL,
  prompt_id CHAR(36) NOT NULL,
  channel_slug VARCHAR(128) NOT NULL,
  version_id CHAR(36) NOT NULL,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_console_prompt_ptr_chan (prompt_id, channel_slug),
  KEY idx_console_prompt_ptr_ns (tenant_id, ns_id)
);
