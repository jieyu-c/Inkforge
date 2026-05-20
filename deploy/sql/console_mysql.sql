-- Console auth MVP (phone + password). Run against Inkforge MySQL schema.
CREATE TABLE IF NOT EXISTS console_users (
  id CHAR(36) NOT NULL PRIMARY KEY,
  phone VARCHAR(20) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  failed_login_attempts INT NOT NULL DEFAULT 0,
  locked_until TIMESTAMP NULL DEFAULT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_console_users_phone (phone)
);

CREATE TABLE IF NOT EXISTS console_sessions (
  id CHAR(36) NOT NULL PRIMARY KEY,
  user_id CHAR(36) NOT NULL,
  family_id CHAR(36) NOT NULL,
  refresh_hash BINARY(32) NOT NULL COMMENT 'SHA-256(raw refresh token)',
  revoked_at TIMESTAMP NULL DEFAULT NULL,
  replaced_by CHAR(36) NULL DEFAULT NULL,
  expires_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  last_ip VARBINARY(16) NULL COMMENT 'parsed IPv4/IPv6',
  ua_hash BINARY(32) NULL COMMENT 'SHA-256(User-Agent)',
  UNIQUE KEY uk_console_sessions_refresh (refresh_hash),
  KEY idx_console_sessions_user (user_id),
  KEY idx_console_sessions_family (family_id)
);

CREATE TABLE IF NOT EXISTS console_namespaces (
  id CHAR(36) NOT NULL,
  tenant_id CHAR(36) NOT NULL COMMENT 'account isolation domain; personal MVP equals user id',
  ns_slug VARCHAR(63) NOT NULL,
  display_name VARCHAR(255) NOT NULL,
  description TEXT,
  tags JSON COMMENT 'opaque JSON tags',
  status VARCHAR(16) NOT NULL DEFAULT 'active' COMMENT 'active|archived',
  default_channel_slug VARCHAR(128) NULL,
  archived_at TIMESTAMP NULL DEFAULT NULL,
  quota_prompts_max INT UNSIGNED NULL COMMENT 'NULL inherit account/platform merge',
  prompt_count INT UNSIGNED NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_console_ns_tenant_slug (tenant_id, ns_slug),
  KEY idx_console_ns_tenant (tenant_id),
  KEY idx_console_ns_tenant_status (tenant_id, status)
);

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
  KEY idx_console_prompt_updated (ns_id, updated_at),
  CONSTRAINT fk_console_prompt_ns FOREIGN KEY (ns_id) REFERENCES console_namespaces (id) ON DELETE CASCADE
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
  KEY idx_console_prompt_ver_ns (tenant_id, ns_id),
  CONSTRAINT fk_console_prompt_ver_prompt FOREIGN KEY (prompt_id) REFERENCES console_prompts (id) ON DELETE CASCADE
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
  KEY idx_console_prompt_ptr_ns (tenant_id, ns_id),
  CONSTRAINT fk_console_prompt_ptr_prompt FOREIGN KEY (prompt_id) REFERENCES console_prompts (id) ON DELETE CASCADE,
  CONSTRAINT fk_console_prompt_ptr_ver FOREIGN KEY (version_id) REFERENCES console_prompt_versions (id) ON DELETE RESTRICT
);
