-- admin_audit_logs was already created by migration 001 with a schema
-- meant for a different, older audit design (actor_user_id, entity_id as
-- a bigint, old_values_json/new_values_json, no public_id). Migration
-- 007's `CREATE TABLE IF NOT EXISTS admin_audit_logs` was therefore
-- silently skipped - the table it actually intended (public_id,
-- admin_user_id, entity_public_id, request_id, before_json/after_json)
-- never got created. Every audit write has been failing ever since (the
-- application code discarded the error rather than surfacing it, which
-- is how this stayed hidden until the audit log was actually read back).
--
-- No other table has a foreign key onto admin_audit_logs, and it holds
-- no meaningful data yet in a fresh install, so the safe fix is to drop
-- and recreate it with the schema migration 007 always intended.

DROP TABLE IF EXISTS admin_audit_logs;

CREATE TABLE admin_audit_logs (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    public_id CHAR(36) NOT NULL,
    admin_user_id BIGINT UNSIGNED NULL,
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(100) NOT NULL,
    entity_public_id CHAR(36) NULL,
    request_id VARCHAR(100) NULL,
    ip_address VARCHAR(64) NULL,
    before_json JSON NULL,
    after_json JSON NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_audit_public_id (public_id),
    KEY idx_audit_entity (entity_type, entity_public_id),
    KEY idx_audit_user_date (admin_user_id, created_at),
    CONSTRAINT fk_audit_admin FOREIGN KEY (admin_user_id) REFERENCES admin_users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
