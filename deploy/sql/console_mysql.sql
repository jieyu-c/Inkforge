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
