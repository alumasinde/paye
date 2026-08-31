-- ============================================================================
-- 006_payroll_platform_foundation.sql
-- Multi-company payroll foundation.
-- All company access is resolved through memberships and database-backed
-- permissions. Roles and permissions are data, not application constants.
-- ============================================================================

CREATE TABLE permissions (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    code VARCHAR(100) NOT NULL,
    name VARCHAR(150) NOT NULL,
    description VARCHAR(500) NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_permissions_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE company_role_templates (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    code VARCHAR(100) NOT NULL,
    name VARCHAR(150) NOT NULL,
    description VARCHAR(500) NULL,
    is_system TINYINT(1) NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_company_role_templates_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE company_role_template_permissions (
    role_template_id BIGINT UNSIGNED NOT NULL,
    permission_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (role_template_id, permission_id),
    CONSTRAINT fk_company_role_template_permissions_role
        FOREIGN KEY (role_template_id) REFERENCES company_role_templates(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_company_role_template_permissions_permission
        FOREIGN KEY (permission_id) REFERENCES permissions(id)
        ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE companies (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    public_id CHAR(36) NOT NULL,
    legal_name VARCHAR(255) NOT NULL,
    trading_name VARCHAR(255) NULL,
    kra_pin VARCHAR(50) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(30) NULL,
    country_code CHAR(2) NOT NULL,
    currency_code CHAR(3) NOT NULL,
    payroll_frequency VARCHAR(30) NOT NULL,
    status ENUM('ACTIVE','SUSPENDED','ARCHIVED') NOT NULL DEFAULT 'ACTIVE',
    settings JSON NULL,
    created_by BIGINT UNSIGNED NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_companies_public_id (public_id),
    UNIQUE KEY uq_companies_kra_pin (kra_pin),
    KEY idx_companies_status (status),
    CONSTRAINT fk_companies_created_by
        FOREIGN KEY (created_by) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE company_roles (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    public_id CHAR(36) NOT NULL,
    company_id BIGINT UNSIGNED NOT NULL,
    code VARCHAR(100) NOT NULL,
    name VARCHAR(150) NOT NULL,
    description VARCHAR(500) NULL,
    is_system TINYINT(1) NOT NULL DEFAULT 0,
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_company_roles_public_id (public_id),
    UNIQUE KEY uq_company_roles_code (company_id, code),
    CONSTRAINT fk_company_roles_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE company_role_permissions (
    role_id BIGINT UNSIGNED NOT NULL,
    permission_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (role_id, permission_id),
    CONSTRAINT fk_company_role_permissions_role
        FOREIGN KEY (role_id) REFERENCES company_roles(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_company_role_permissions_permission
        FOREIGN KEY (permission_id) REFERENCES permissions(id)
        ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE company_members (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    public_id CHAR(36) NOT NULL,
    company_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    role_id BIGINT UNSIGNED NOT NULL,
    status ENUM('ACTIVE','SUSPENDED','REMOVED') NOT NULL DEFAULT 'ACTIVE',
    joined_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_company_members_public_id (public_id),
    UNIQUE KEY uq_company_members_company_user (company_id, user_id),
    KEY idx_company_members_company_status (company_id, status),
    CONSTRAINT fk_company_members_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_company_members_user
        FOREIGN KEY (user_id) REFERENCES users(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_company_members_role
        FOREIGN KEY (role_id) REFERENCES company_roles(id)
        ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO permissions (code, name, description) VALUES
('company.read', 'View company', 'View company information and membership.'),
('company.write', 'Manage company', 'Update company information and settings.'),
('members.read', 'View members', 'View company users and their roles.'),
('members.write', 'Manage members', 'Add members and change their role.'),
('roles.read', 'View roles', 'View company roles and permissions.'),
('roles.write', 'Manage roles', 'Create company roles and assign permissions.'),
('payroll.read', 'View payroll', 'Reserved for payroll access.'),
('payroll.write', 'Manage payroll', 'Reserved for payroll creation and changes.'),
('billing.read', 'View billing', 'Reserved for subscription and invoice access.'),
('billing.write', 'Manage billing', 'Reserved for subscription management.')
ON DUPLICATE KEY UPDATE
    name = VALUES(name),
    description = VALUES(description);

INSERT INTO company_role_templates (code, name, description, is_system) VALUES
('OWNER', 'Owner', 'Full control of the company payroll account.', 1),
('ADMIN', 'Administrator', 'Manages company configuration and payroll operations.', 1),
('PAYROLL_MANAGER', 'Payroll Manager', 'Manages payroll work without company ownership.', 1),
('VIEWER', 'Viewer', 'Read-only access.', 1)
ON DUPLICATE KEY UPDATE
    name = VALUES(name),
    description = VALUES(description),
    is_system = VALUES(is_system);

INSERT IGNORE INTO company_role_template_permissions (role_template_id, permission_id)
SELECT rt.id, p.id
FROM company_role_templates rt
JOIN permissions p
WHERE rt.code = 'OWNER';

INSERT IGNORE INTO company_role_template_permissions (role_template_id, permission_id)
SELECT rt.id, p.id
FROM company_role_templates rt
JOIN permissions p
WHERE rt.code = 'ADMIN'
  AND p.code IN ('company.read','company.write','members.read','members.write','roles.read','roles.write','payroll.read','payroll.write','billing.read');

INSERT IGNORE INTO company_role_template_permissions (role_template_id, permission_id)
SELECT rt.id, p.id
FROM company_role_templates rt
JOIN permissions p
WHERE rt.code = 'PAYROLL_MANAGER'
  AND p.code IN ('company.read','members.read','roles.read','payroll.read','payroll.write');

INSERT IGNORE INTO company_role_template_permissions (role_template_id, permission_id)
SELECT rt.id, p.id
FROM company_role_templates rt
JOIN permissions p
WHERE rt.code = 'VIEWER'
  AND p.code IN ('company.read','members.read','roles.read','payroll.read');

-- Existing global admin roles are intentionally left untouched. Company roles
-- are scoped to one company so one employer can never grant access to another.
