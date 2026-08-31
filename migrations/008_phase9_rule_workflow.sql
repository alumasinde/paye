-- Phase 9 workflow hardening.

ALTER TABLE payroll_rule_change_requests
    MODIFY status ENUM('OPEN','APPROVED','REJECTED','CANCELLED') NOT NULL DEFAULT 'OPEN';

CREATE TABLE IF NOT EXISTS payroll_rule_validation_runs (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    public_id CHAR(36) NOT NULL,
    rule_set_id BIGINT UNSIGNED NOT NULL,
    valid BOOLEAN NOT NULL,
    errors_json JSON NOT NULL,
    warnings_json JSON NOT NULL,
    validated_by BIGINT UNSIGNED NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY(id),
    UNIQUE KEY uq_validation_public_id(public_id),
    KEY idx_validation_ruleset(rule_set_id,created_at),
    CONSTRAINT fk_validation_ruleset FOREIGN KEY(rule_set_id) REFERENCES payroll_rule_sets(id) ON DELETE CASCADE,
    CONSTRAINT fk_validation_admin FOREIGN KEY(validated_by) REFERENCES admin_users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS payroll_rule_publish_events (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    public_id CHAR(36) NOT NULL,
    rule_set_id BIGINT UNSIGNED NOT NULL,
    previous_rule_set_id BIGINT UNSIGNED NULL,
    published_by BIGINT UNSIGNED NOT NULL,
    published_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY(id),
    UNIQUE KEY uq_publish_event_public_id(public_id),
    KEY idx_publish_rule(rule_set_id),
    CONSTRAINT fk_publish_rule FOREIGN KEY(rule_set_id) REFERENCES payroll_rule_sets(id),
    CONSTRAINT fk_publish_previous FOREIGN KEY(previous_rule_set_id) REFERENCES payroll_rule_sets(id),
    CONSTRAINT fk_publish_admin FOREIGN KEY(published_by) REFERENCES admin_users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
