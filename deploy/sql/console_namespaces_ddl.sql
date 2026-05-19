-- Namespace MVP (personal account). Matches deploy/sql/console_mysql.sql appendix.
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
