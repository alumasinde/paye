-- Phase 8: Dynamic payroll rule management and admin governance.

CREATE TABLE IF NOT EXISTS admin_roles (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(255) NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_admin_roles_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS admin_permissions (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    code VARCHAR(100) NOT NULL,
    description VARCHAR(255) NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_admin_permissions_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS admin_role_permissions (
    role_id BIGINT UNSIGNED NOT NULL,
    permission_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (role_id, permission_id),
    CONSTRAINT fk_arp_role FOREIGN KEY (role_id) REFERENCES admin_roles(id) ON DELETE CASCADE,
    CONSTRAINT fk_arp_permission FOREIGN KEY (permission_id) REFERENCES admin_permissions(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS admin_users (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    public_id CHAR(36) NOT NULL,
    email VARCHAR(254) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    status ENUM('ACTIVE','DISABLED') NOT NULL DEFAULT 'ACTIVE',
    last_login_at DATETIME NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_admin_public_id (public_id),
    UNIQUE KEY uq_admin_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS admin_user_roles (
    admin_user_id BIGINT UNSIGNED NOT NULL,
    role_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (admin_user_id, role_id),
    CONSTRAINT fk_aur_user FOREIGN KEY (admin_user_id) REFERENCES admin_users(id) ON DELETE CASCADE,
    CONSTRAINT fk_aur_role FOREIGN KEY (role_id) REFERENCES admin_roles(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS admin_refresh_tokens (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    admin_user_id BIGINT UNSIGNED NOT NULL,
    token_hash CHAR(64) NOT NULL,
    expires_at DATETIME NOT NULL,
    revoked_at DATETIME NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_admin_refresh_hash (token_hash),
    KEY idx_admin_refresh_user (admin_user_id),
    CONSTRAINT fk_admin_refresh_user FOREIGN KEY (admin_user_id) REFERENCES admin_users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- A rule set groups one coherent version of Kenya payroll rules.
CREATE TABLE IF NOT EXISTS payroll_rule_sets (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    public_id CHAR(36) NOT NULL,
    code VARCHAR(100) NOT NULL,
    name VARCHAR(150) NOT NULL,
    jurisdiction VARCHAR(20) NOT NULL DEFAULT 'KE',
    effective_from DATE NOT NULL,
    effective_to DATE NULL,
    status ENUM('DRAFT','IN_REVIEW','PUBLISHED','ARCHIVED') NOT NULL DEFAULT 'DRAFT',
    version_number INT UNSIGNED NOT NULL,
    source_reference VARCHAR(500) NULL,
    source_notes TEXT NULL,
    published_at DATETIME NULL,
    published_by BIGINT UNSIGNED NULL,
    created_by BIGINT UNSIGNED NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_rule_set_public_id (public_id),
    UNIQUE KEY uq_rule_set_version (code, version_number),
    KEY idx_rule_set_effective (jurisdiction, effective_from, effective_to, status),
    CONSTRAINT fk_ruleset_created_by FOREIGN KEY (created_by) REFERENCES admin_users(id),
    CONSTRAINT fk_ruleset_published_by FOREIGN KEY (published_by) REFERENCES admin_users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS payroll_rule_components (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    public_id CHAR(36) NOT NULL,
    rule_set_id BIGINT UNSIGNED NOT NULL,
    component_code VARCHAR(100) NOT NULL,
    component_type ENUM('PAYE_BANDS','STATUTORY_DEDUCTION','RELIEF','CONFIGURATION') NOT NULL,
    name VARCHAR(150) NOT NULL,
    calculation_order INT NOT NULL DEFAULT 100,
    reduces_taxable_income BOOLEAN NOT NULL DEFAULT FALSE,
    reduces_net_pay BOOLEAN NOT NULL DEFAULT FALSE,
    formula_type ENUM('BANDS','PERCENTAGE','FIXED','JSON') NOT NULL,
    payload JSON NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_component_public_id (public_id),
    UNIQUE KEY uq_component_per_ruleset (rule_set_id, component_code),
    CONSTRAINT fk_component_ruleset FOREIGN KEY (rule_set_id) REFERENCES payroll_rule_sets(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS payroll_rule_change_requests (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    public_id CHAR(36) NOT NULL,
    rule_set_id BIGINT UNSIGNED NOT NULL,
    requested_by BIGINT UNSIGNED NOT NULL,
    reviewed_by BIGINT UNSIGNED NULL,
    status ENUM('OPEN','APPROVED','REJECTED','CANCELLED') NOT NULL DEFAULT 'OPEN',
    review_comment TEXT NULL,
    requested_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    reviewed_at DATETIME NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_change_request_public_id (public_id),
    CONSTRAINT fk_change_rule_set FOREIGN KEY (rule_set_id) REFERENCES payroll_rule_sets(id) ON DELETE CASCADE,
    CONSTRAINT fk_change_requested_by FOREIGN KEY (requested_by) REFERENCES admin_users(id),
    CONSTRAINT fk_change_reviewed_by FOREIGN KEY (reviewed_by) REFERENCES admin_users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS admin_audit_logs (
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

INSERT IGNORE INTO admin_roles(code,name,description) VALUES
('SUPER_ADMIN','Super Administrator','Full administrative access'),
('RULE_EDITOR','Payroll Rule Editor','Can create and edit drafts'),
('RULE_APPROVER','Payroll Rule Approver','Can review and publish approved rules'),
('AUDITOR','Auditor','Read-only audit and rule access');

INSERT IGNORE INTO admin_permissions(code,description) VALUES
('rules.read','View payroll rules'),
('rules.write','Create and edit draft payroll rules'),
('rules.review','Review rule change requests'),
('rules.publish','Publish approved payroll rules'),
('audit.read','View audit logs'),
('admin.users','Manage admin users');

-- Roles and permissions above were previously left unlinked - no role
-- granted any permission, so every permission check would fail for every
-- admin regardless of role. Grants below match each role's stated
-- description.
INSERT IGNORE INTO admin_role_permissions(role_id, permission_id)
SELECT ar.id, ap.id FROM admin_roles ar JOIN admin_permissions ap
WHERE ar.code = 'SUPER_ADMIN';

INSERT IGNORE INTO admin_role_permissions(role_id, permission_id)
SELECT ar.id, ap.id FROM admin_roles ar JOIN admin_permissions ap
WHERE ar.code = 'RULE_EDITOR' AND ap.code IN ('rules.read','rules.write');

INSERT IGNORE INTO admin_role_permissions(role_id, permission_id)
SELECT ar.id, ap.id FROM admin_roles ar JOIN admin_permissions ap
WHERE ar.code = 'RULE_APPROVER' AND ap.code IN ('rules.read','rules.review','rules.publish');

INSERT IGNORE INTO admin_role_permissions(role_id, permission_id)
SELECT ar.id, ap.id FROM admin_roles ar JOIN admin_permissions ap
WHERE ar.code = 'AUDITOR' AND ap.code IN ('rules.read','audit.read');
