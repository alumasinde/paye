-- Phase 11: link the admin governance rule set to the immutable live rule versions.
CREATE TABLE IF NOT EXISTS payroll_rule_set_live_versions (
    rule_set_id BIGINT UNSIGNED NOT NULL,
    rule_version_id BIGINT UNSIGNED NOT NULL,
    component_code VARCHAR(100) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (rule_set_id, component_code),
    UNIQUE KEY uq_prslv_live_version (rule_version_id),
    CONSTRAINT fk_prslv_rule_set FOREIGN KEY (rule_set_id) REFERENCES payroll_rule_sets(id) ON DELETE CASCADE,
    CONSTRAINT fk_prslv_rule_version FOREIGN KEY (rule_version_id) REFERENCES rule_versions(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
