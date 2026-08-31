-- ============================================================================
-- Budget254 PAYE Calculator
-- Database: budget254_paye
-- Engine: MySQL 8.0+
-- Charset: utf8mb4
--
-- Design goals
-- 1. Historical calculations from January 2022 onward.
-- 2. Government rates/bands are data-driven, versioned and effective-dated.
-- 3. Rule changes do not require Go code changes when an existing calculation
--    method can express the new rule.
-- 4. Historical calculations remain reproducible through calculation snapshots.
-- 5. Government rules and user custom deductions are kept separate.
-- 6. Admin changes follow Draft -> Review -> Approved -> Published/Archived.
-- ============================================================================

-- The target database is created and selected by cmd/migrate (using
-- DB_NAME from .env) before this file runs, so this migration is
-- database-name-agnostic and does not hardcode a name here.

-- ============================================================================
-- USERS
-- ============================================================================

CREATE TABLE users (
    id                  BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    public_id           CHAR(36) NOT NULL,
    first_name          VARCHAR(100) NULL,
    last_name           VARCHAR(100) NULL,
    email               VARCHAR(255) NULL,
    phone               VARCHAR(30) NULL,
    password_hash       VARCHAR(255) NULL,
    status              ENUM('ACTIVE','SUSPENDED','PENDING') NOT NULL DEFAULT 'ACTIVE',
    email_verified_at   DATETIME NULL,
    last_login_at       DATETIME NULL,
    created_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    UNIQUE KEY uq_users_public_id (public_id),
    UNIQUE KEY uq_users_email (email),
    UNIQUE KEY uq_users_phone (phone),
    KEY idx_users_status (status)
) ENGINE=InnoDB;

-- ============================================================================
-- ADMIN ROLES AND USERS
-- ============================================================================

CREATE TABLE roles (
    id                  BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    code                VARCHAR(100) NOT NULL,
    name                VARCHAR(150) NOT NULL,
    description         VARCHAR(500) NULL,
    created_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    UNIQUE KEY uq_roles_code (code)
) ENGINE=InnoDB;

CREATE TABLE user_roles (
    user_id             BIGINT UNSIGNED NOT NULL,
    role_id             BIGINT UNSIGNED NOT NULL,
    assigned_at         DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (user_id, role_id),
    CONSTRAINT fk_user_roles_user
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_user_roles_role
        FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
) ENGINE=InnoDB;

-- ============================================================================
-- GOVERNMENT RULE CATALOG
--
-- One row represents a stable rule concept, e.g. PAYE, NSSF, NHIF, SHIF,
-- AHL or PERSONAL_RELIEF. Historical changes are stored in rule_versions.
-- ============================================================================

CREATE TABLE rule_definitions (
    id                  BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    code                VARCHAR(100) NOT NULL,
    name                VARCHAR(200) NOT NULL,
    category            ENUM(
                            'INCOME_TAX',
                            'STATUTORY_DEDUCTION',
                            'STATUTORY_CONTRIBUTION',
                            'RELIEF',
                            'ALLOWABLE_DEDUCTION',
                            'OTHER'
                        ) NOT NULL,
    description         TEXT NULL,
    is_active           TINYINT(1) NOT NULL DEFAULT 1,
    created_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    UNIQUE KEY uq_rule_definitions_code (code),
    KEY idx_rule_definitions_category (category, is_active)
) ENGINE=InnoDB;

-- ============================================================================
-- GOVERNMENT SOURCE DOCUMENTS
--
-- Every published rule should have an official source reference where possible.
-- ============================================================================

CREATE TABLE rule_sources (
    id                  BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    authority           VARCHAR(200) NOT NULL,
    title               VARCHAR(500) NOT NULL,
    source_url          VARCHAR(1000) NULL,
    reference_number    VARCHAR(255) NULL,
    published_on        DATE NULL,
    accessed_on         DATE NULL,
    notes               TEXT NULL,
    created_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    KEY idx_rule_sources_authority (authority),
    KEY idx_rule_sources_published_on (published_on)
) ENGINE=InnoDB;

-- ============================================================================
-- CALCULATION METHODS
--
-- These are generic engine capabilities. Admins select a method and configure
-- its parameters; admins do not upload executable formulas/code.
-- ============================================================================

CREATE TABLE calculation_methods (
    id                  BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    code                VARCHAR(100) NOT NULL,
    name                VARCHAR(150) NOT NULL,
    description         VARCHAR(500) NULL,
    is_active           TINYINT(1) NOT NULL DEFAULT 1,
    created_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    UNIQUE KEY uq_calculation_methods_code (code)
) ENGINE=InnoDB;

-- Recommended seed method codes:
-- FIXED_AMOUNT
-- PERCENTAGE
-- PERCENTAGE_WITH_MINIMUM
-- PERCENTAGE_WITH_MIN_MAX
-- CAPPED_PERCENTAGE
-- PROGRESSIVE_BANDS
-- TIERED_FIXED_AMOUNT
-- TIERED_PERCENTAGE

-- ============================================================================
-- RULE VERSIONS
--
-- Draft/publish workflow and effective dates live here.
-- Existing published versions should never be overwritten.
-- ============================================================================

CREATE TABLE rule_versions (
    id                      BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    rule_definition_id      BIGINT UNSIGNED NOT NULL,
    calculation_method_id   BIGINT UNSIGNED NOT NULL,

    version_code            VARCHAR(150) NOT NULL,
    version_name            VARCHAR(255) NOT NULL,

    status                  ENUM(
                                'DRAFT',
                                'IN_REVIEW',
                                'APPROVED',
                                'PUBLISHED',
                                'ARCHIVED'
                            ) NOT NULL DEFAULT 'DRAFT',

    effective_from          DATE NOT NULL,
    effective_to            DATE NULL,

    calculation_order       INT NOT NULL DEFAULT 100,

    affects_taxable_income  TINYINT(1) NOT NULL DEFAULT 0,
    affects_net_pay         TINYINT(1) NOT NULL DEFAULT 1,

    source_summary          VARCHAR(1000) NULL,
    notes                   TEXT NULL,

    created_by              BIGINT UNSIGNED NULL,
    reviewed_by             BIGINT UNSIGNED NULL,
    approved_by             BIGINT UNSIGNED NULL,
    published_by            BIGINT UNSIGNED NULL,

    created_at              DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    reviewed_at             DATETIME NULL,
    approved_at             DATETIME NULL,
    published_at            DATETIME NULL,
    updated_at              DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (id),

    UNIQUE KEY uq_rule_versions_definition_version
        (rule_definition_id, version_code),

    KEY idx_rule_versions_lookup
        (rule_definition_id, status, effective_from, effective_to),

    KEY idx_rule_versions_status (status, effective_from),

    CONSTRAINT fk_rule_versions_definition
        FOREIGN KEY (rule_definition_id)
        REFERENCES rule_definitions(id),

    CONSTRAINT fk_rule_versions_method
        FOREIGN KEY (calculation_method_id)
        REFERENCES calculation_methods(id),

    CONSTRAINT fk_rule_versions_created_by
        FOREIGN KEY (created_by)
        REFERENCES users(id) ON DELETE SET NULL,

    CONSTRAINT fk_rule_versions_reviewed_by
        FOREIGN KEY (reviewed_by)
        REFERENCES users(id) ON DELETE SET NULL,

    CONSTRAINT fk_rule_versions_approved_by
        FOREIGN KEY (approved_by)
        REFERENCES users(id) ON DELETE SET NULL,

    CONSTRAINT fk_rule_versions_published_by
        FOREIGN KEY (published_by)
        REFERENCES users(id) ON DELETE SET NULL,

    CONSTRAINT chk_rule_versions_dates
        CHECK (effective_to IS NULL OR effective_to >= effective_from)
) ENGINE=InnoDB;

-- ============================================================================
-- RULE VERSION SOURCE LINKS
-- ============================================================================

CREATE TABLE rule_version_sources (
    rule_version_id         BIGINT UNSIGNED NOT NULL,
    rule_source_id          BIGINT UNSIGNED NOT NULL,
    created_at              DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (rule_version_id, rule_source_id),

    CONSTRAINT fk_rule_version_sources_version
        FOREIGN KEY (rule_version_id)
        REFERENCES rule_versions(id) ON DELETE CASCADE,

    CONSTRAINT fk_rule_version_sources_source
        FOREIGN KEY (rule_source_id)
        REFERENCES rule_sources(id) ON DELETE CASCADE
) ENGINE=InnoDB;

-- ============================================================================
-- RULE PARAMETERS
--
-- Generic parameters such as rate, minimum_amount, upper_earnings_limit,
-- personal_relief_amount, calculation_base, etc.
-- DECIMAL values are stored as strings/numbers in value_decimal.
-- ============================================================================

CREATE TABLE rule_parameters (
    id                      BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    rule_version_id         BIGINT UNSIGNED NOT NULL,
    parameter_name          VARCHAR(150) NOT NULL,
    parameter_type          ENUM(
                                'DECIMAL',
                                'INTEGER',
                                'BOOLEAN',
                                'TEXT'
                            ) NOT NULL,

    value_decimal           DECIMAL(24,8) NULL,
    value_integer           BIGINT NULL,
    value_boolean           TINYINT(1) NULL,
    value_text              VARCHAR(1000) NULL,

    created_at              DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at              DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    UNIQUE KEY uq_rule_parameters_version_name
        (rule_version_id, parameter_name),

    CONSTRAINT fk_rule_parameters_version
        FOREIGN KEY (rule_version_id)
        REFERENCES rule_versions(id) ON DELETE CASCADE
) ENGINE=InnoDB;

-- ============================================================================
-- RULE BANDS
--
-- Used for PAYE progressive bands and NHIF-style tiered fixed bands.
--
-- Examples:
-- PROGRESSIVE_BANDS: from_amount, to_amount, rate
-- TIERED_FIXED_AMOUNT: from_amount, to_amount, fixed_amount
-- ============================================================================

CREATE TABLE rule_bands (
    id                      BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    rule_version_id         BIGINT UNSIGNED NOT NULL,

    from_amount             DECIMAL(24,2) NOT NULL DEFAULT 0.00,
    to_amount               DECIMAL(24,2) NULL,

    rate                    DECIMAL(12,8) NULL,
    fixed_amount            DECIMAL(24,2) NULL,

    display_order           INT NOT NULL DEFAULT 1,
    label                   VARCHAR(255) NULL,

    created_at              DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at              DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (id),

    KEY idx_rule_bands_version_order
        (rule_version_id, display_order),

    CONSTRAINT fk_rule_bands_version
        FOREIGN KEY (rule_version_id)
        REFERENCES rule_versions(id) ON DELETE CASCADE,

    CONSTRAINT chk_rule_bands_range
        CHECK (to_amount IS NULL OR to_amount >= from_amount)
) ENGINE=InnoDB;

-- ============================================================================
-- RULE RELATIONSHIPS
--
-- Allows the engine/admin to express relationships without hardcoding.
-- Example: a relief may apply after PAYE; an allowable deduction may affect
-- taxable income.
-- ============================================================================

CREATE TABLE rule_dependencies (
    id                      BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    rule_version_id         BIGINT UNSIGNED NOT NULL,
    depends_on_rule_code    VARCHAR(100) NOT NULL,
    dependency_type         ENUM(
                                'CALCULATION_BASE',
                                'MUST_RUN_BEFORE',
                                'MUST_RUN_AFTER',
                                'INFORMATIONAL'
                            ) NOT NULL,
    created_at              DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    UNIQUE KEY uq_rule_dependency
        (rule_version_id, depends_on_rule_code, dependency_type),

    CONSTRAINT fk_rule_dependencies_version
        FOREIGN KEY (rule_version_id)
        REFERENCES rule_versions(id) ON DELETE CASCADE
) ENGINE=InnoDB;

-- ============================================================================
-- EMPLOYEE PROFILES
--
-- Optional. Anonymous/quick calculations do not require a profile.
-- ============================================================================

CREATE TABLE employee_profiles (
    id                      BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    public_id               CHAR(36) NOT NULL,
    user_id                 BIGINT UNSIGNED NOT NULL,

    first_name              VARCHAR(100) NOT NULL,
    last_name               VARCHAR(100) NOT NULL,
    employee_number         VARCHAR(100) NULL,
    kra_pin                 VARCHAR(50) NULL,
    job_title               VARCHAR(150) NULL,
    department              VARCHAR(150) NULL,

    is_default              TINYINT(1) NOT NULL DEFAULT 0,
    is_active               TINYINT(1) NOT NULL DEFAULT 1,

    created_at              DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at              DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    UNIQUE KEY uq_employee_profiles_public_id (public_id),
    KEY idx_employee_profiles_user (user_id, is_active),

    CONSTRAINT fk_employee_profiles_user
        FOREIGN KEY (user_id)
        REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB;

-- ============================================================================
-- SAVED USER CUSTOM DEDUCTION TEMPLATES
--
-- These are normally post-tax deductions such as savings/welfare/chama.
-- Tax-allowable categories should be validated by backend rule logic.
-- ============================================================================

CREATE TABLE deduction_templates (
    id                      BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    public_id               CHAR(36) NOT NULL,
    user_id                 BIGINT UNSIGNED NOT NULL,

    name                    VARCHAR(200) NOT NULL,

    category                ENUM(
                                'SAVINGS',
                                'WELFARE',
                                'SACCO',
                                'CHAMA',
                                'LOAN',
                                'SALARY_ADVANCE',
                                'INSURANCE',
                                'PENSION',
                                'UNION',
                                'OTHER'
                            ) NOT NULL DEFAULT 'OTHER',

    calculation_type        ENUM(
                                'FIXED_AMOUNT',
                                'PERCENTAGE'
                            ) NOT NULL,

    fixed_amount            DECIMAL(24,2) NULL,
    percentage              DECIMAL(12,8) NULL,

    calculation_base        ENUM(
                                'BASIC_PAY',
                                'GROSS_PAY',
                                'NET_PAY'
                            ) NULL,

    tax_treatment           ENUM(
                                'POST_TAX',
                                'TAX_ALLOWABLE'
                            ) NOT NULL DEFAULT 'POST_TAX',

    is_active               TINYINT(1) NOT NULL DEFAULT 1,

    created_at              DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at              DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    UNIQUE KEY uq_deduction_templates_public_id (public_id),
    KEY idx_deduction_templates_user (user_id, is_active),

    CONSTRAINT fk_deduction_templates_user
        FOREIGN KEY (user_id)
        REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB;

-- ============================================================================
-- CALCULATIONS
--
-- A calculation represents one payroll calculation for one selected pay period.
-- It stores summary values for fast history queries.
-- ============================================================================

CREATE TABLE calculations (
    id                      BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    public_id               CHAR(36) NOT NULL,

    user_id                 BIGINT UNSIGNED NULL,
    employee_profile_id     BIGINT UNSIGNED NULL,

    pay_period_start        DATE NOT NULL,
    pay_period_end          DATE NOT NULL,

    calculation_status      ENUM(
                                'CALCULATED',
                                'SAVED',
                                'ARCHIVED'
                            ) NOT NULL DEFAULT 'CALCULATED',

    gross_pay               DECIMAL(24,2) NOT NULL DEFAULT 0.00,
    taxable_income          DECIMAL(24,2) NOT NULL DEFAULT 0.00,

    gross_tax               DECIMAL(24,2) NOT NULL DEFAULT 0.00,
    total_reliefs           DECIMAL(24,2) NOT NULL DEFAULT 0.00,
    paye                    DECIMAL(24,2) NOT NULL DEFAULT 0.00,

    total_statutory_deductions DECIMAL(24,2) NOT NULL DEFAULT 0.00,
    total_other_deductions     DECIMAL(24,2) NOT NULL DEFAULT 0.00,
    total_deductions           DECIMAL(24,2) NOT NULL DEFAULT 0.00,

    net_pay                 DECIMAL(24,2) NOT NULL DEFAULT 0.00,

    engine_version          VARCHAR(100) NOT NULL,
    calculated_at           DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    created_at              DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at              DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    UNIQUE KEY uq_calculations_public_id (public_id),

    KEY idx_calculations_user_period
        (user_id, pay_period_start),

    KEY idx_calculations_employee_period
        (employee_profile_id, pay_period_start),

    KEY idx_calculations_period
        (pay_period_start, pay_period_end),

    CONSTRAINT fk_calculations_user
        FOREIGN KEY (user_id)
        REFERENCES users(id) ON DELETE SET NULL,

    CONSTRAINT fk_calculations_employee
        FOREIGN KEY (employee_profile_id)
        REFERENCES employee_profiles(id) ON DELETE SET NULL,

    CONSTRAINT chk_calculations_period
        CHECK (pay_period_end >= pay_period_start)
) ENGINE=InnoDB;

-- ============================================================================
-- CALCULATION ITEMS
--
-- Full breakdown: earnings, deductions, taxes, reliefs and contributions.
-- This is what the Android app uses to render the detailed explanation.
-- ============================================================================

CREATE TABLE calculation_items (
    id                      BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    calculation_id          BIGINT UNSIGNED NOT NULL,

    item_type               ENUM(
                                'EARNING',
                                'TAXABLE_ALLOWANCE',
                                'NON_TAXABLE_ALLOWANCE',
                                'ALLOWABLE_DEDUCTION',
                                'STATUTORY_DEDUCTION',
                                'POST_TAX_DEDUCTION',
                                'TAX',
                                'RELIEF',
                                'EMPLOYER_CONTRIBUTION'
                            ) NOT NULL,

    code                    VARCHAR(100) NOT NULL,
    name                    VARCHAR(255) NOT NULL,

    amount                  DECIMAL(24,2) NOT NULL DEFAULT 0.00,

    calculation_order       INT NOT NULL DEFAULT 100,

    source_type             ENUM(
                                'SYSTEM_RULE',
                                'USER_INPUT',
                                'DEDUCTION_TEMPLATE',
                                'ENGINE'
                            ) NOT NULL DEFAULT 'ENGINE',

    source_id               BIGINT UNSIGNED NULL,

    is_statutory            TINYINT(1) NOT NULL DEFAULT 0,
    affects_taxable_income  TINYINT(1) NOT NULL DEFAULT 0,
    affects_net_pay         TINYINT(1) NOT NULL DEFAULT 1,

    metadata_json           JSON NULL,

    created_at              DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (id),

    KEY idx_calculation_items_calculation_order
        (calculation_id, calculation_order),

    KEY idx_calculation_items_type
        (calculation_id, item_type),

    CONSTRAINT fk_calculation_items_calculation
        FOREIGN KEY (calculation_id)
        REFERENCES calculations(id) ON DELETE CASCADE
) ENGINE=InnoDB;

-- ============================================================================
-- CALCULATION RULE SNAPSHOTS
--
-- The exact rule version and configuration used by a saved calculation.
-- This protects historical reproducibility even after future admin changes.
-- ============================================================================

CREATE TABLE calculation_rule_snapshots (
    id                      BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    calculation_id          BIGINT UNSIGNED NOT NULL,

    rule_version_id         BIGINT UNSIGNED NULL,
    rule_code               VARCHAR(100) NOT NULL,
    rule_version_code       VARCHAR(150) NOT NULL,

    effective_from          DATE NOT NULL,
    effective_to            DATE NULL,

    calculation_method_code VARCHAR(100) NOT NULL,

    configuration_json      JSON NOT NULL,

    created_at              DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (id),

    UNIQUE KEY uq_calculation_rule_snapshots
        (calculation_id, rule_code, rule_version_code),

    KEY idx_calculation_rule_snapshots_rule_version
        (rule_version_id),

    CONSTRAINT fk_calculation_rule_snapshots_calculation
        FOREIGN KEY (calculation_id)
        REFERENCES calculations(id) ON DELETE CASCADE,

    CONSTRAINT fk_calculation_rule_snapshots_rule_version
        FOREIGN KEY (rule_version_id)
        REFERENCES rule_versions(id) ON DELETE SET NULL
) ENGINE=InnoDB;

-- ============================================================================
-- CALCULATION STEPS
--
-- Optional detailed audit/explanation for "How was this calculated?"
-- ============================================================================

CREATE TABLE calculation_steps (
    id                      BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    calculation_id          BIGINT UNSIGNED NOT NULL,

    step_order              INT NOT NULL,
    step_code               VARCHAR(100) NOT NULL,
    title                   VARCHAR(255) NOT NULL,
    description             TEXT NULL,

    input_amount            DECIMAL(24,2) NULL,
    output_amount           DECIMAL(24,2) NULL,

    details_json            JSON NULL,

    created_at              DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (id),

    UNIQUE KEY uq_calculation_steps_order
        (calculation_id, step_order),

    CONSTRAINT fk_calculation_steps_calculation
        FOREIGN KEY (calculation_id)
        REFERENCES calculations(id) ON DELETE CASCADE
) ENGINE=InnoDB;

-- ============================================================================
-- USER SAVED CALCULATION DEDUCTION APPLICATIONS
--
-- Keeps the user/template relationship visible without relying only on
-- calculation_items metadata.
-- ============================================================================

CREATE TABLE calculation_custom_deductions (
    id                      BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    calculation_id          BIGINT UNSIGNED NOT NULL,
    deduction_template_id   BIGINT UNSIGNED NULL,

    name                    VARCHAR(200) NOT NULL,
    category                VARCHAR(100) NOT NULL,
    calculation_type        VARCHAR(50) NOT NULL,

    amount                  DECIMAL(24,2) NOT NULL,

    created_at              DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (id),

    KEY idx_calculation_custom_deductions_calculation
        (calculation_id),

    CONSTRAINT fk_calculation_custom_deductions_calculation
        FOREIGN KEY (calculation_id)
        REFERENCES calculations(id) ON DELETE CASCADE,

    CONSTRAINT fk_calculation_custom_deductions_template
        FOREIGN KEY (deduction_template_id)
        REFERENCES deduction_templates(id) ON DELETE SET NULL
) ENGINE=InnoDB;

-- ============================================================================
-- ADMIN AUDIT LOG
--
-- Every government-rule change should be auditable.
-- ============================================================================

CREATE TABLE admin_audit_logs (
    id                      BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,

    actor_user_id           BIGINT UNSIGNED NULL,

    entity_type             VARCHAR(100) NOT NULL,
    entity_id               BIGINT UNSIGNED NULL,

    action                  VARCHAR(100) NOT NULL,

    old_values_json         JSON NULL,
    new_values_json         JSON NULL,

    ip_address              VARCHAR(45) NULL,
    user_agent              VARCHAR(1000) NULL,

    created_at              DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (id),

    KEY idx_admin_audit_logs_actor
        (actor_user_id, created_at),

    KEY idx_admin_audit_logs_entity
        (entity_type, entity_id, created_at),

    CONSTRAINT fk_admin_audit_logs_actor
        FOREIGN KEY (actor_user_id)
        REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB;

-- ============================================================================
-- CALCULATION TEST CASES
--
-- Admins can define expected outputs before publishing a rule.
-- ============================================================================

CREATE TABLE rule_test_cases (
    id                      BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    rule_version_id         BIGINT UNSIGNED NOT NULL,

    name                    VARCHAR(255) NOT NULL,
    description             TEXT NULL,

    input_json              JSON NOT NULL,
    expected_output_json    JSON NOT NULL,

    is_active               TINYINT(1) NOT NULL DEFAULT 1,

    created_by              BIGINT UNSIGNED NULL,
    created_at              DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at              DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (id),

    KEY idx_rule_test_cases_version
        (rule_version_id, is_active),

    CONSTRAINT fk_rule_test_cases_version
        FOREIGN KEY (rule_version_id)
        REFERENCES rule_versions(id) ON DELETE CASCADE,

    CONSTRAINT fk_rule_test_cases_created_by
        FOREIGN KEY (created_by)
        REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB;

-- ============================================================================
-- OPTIONAL: APP CONFIGURATION
--
-- For non-tax settings that may need admin management.
-- ============================================================================

CREATE TABLE app_settings (
    id                      BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    setting_key             VARCHAR(150) NOT NULL,
    setting_value           JSON NOT NULL,
    description             VARCHAR(500) NULL,
    updated_by              BIGINT UNSIGNED NULL,
    updated_at              DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    UNIQUE KEY uq_app_settings_key (setting_key),

    CONSTRAINT fk_app_settings_updated_by
        FOREIGN KEY (updated_by)
        REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB;

-- ============================================================================
-- SEED: ADMIN ROLES
-- ============================================================================

INSERT INTO roles (code, name, description) VALUES
('SUPER_ADMIN', 'Super Admin', 'Full control of Budget254 PAYE administration'),
('RULE_EDITOR', 'Rule Editor', 'Can create and edit draft government rules'),
('RULE_REVIEWER', 'Rule Reviewer', 'Can review and approve rules'),
('SUPPORT', 'Support', 'Can assist users without publishing tax rules');

-- ============================================================================
-- SEED: GENERIC CALCULATION METHODS
-- ============================================================================

INSERT INTO calculation_methods (code, name, description) VALUES
('FIXED_AMOUNT', 'Fixed Amount', 'Returns a configured fixed amount'),
('PERCENTAGE', 'Percentage', 'Percentage of a configured calculation base'),
('PERCENTAGE_WITH_MINIMUM', 'Percentage With Minimum', 'Percentage subject to a minimum amount'),
('PERCENTAGE_WITH_MIN_MAX', 'Percentage With Minimum and Maximum', 'Percentage subject to minimum and maximum limits'),
('CAPPED_PERCENTAGE', 'Capped Percentage', 'Percentage of a calculation base capped at a configured amount'),
('PROGRESSIVE_BANDS', 'Progressive Bands', 'Applies progressive tax rates across ordered bands'),
('TIERED_FIXED_AMOUNT', 'Tiered Fixed Amount', 'Selects one fixed amount based on an income band'),
('TIERED_PERCENTAGE', 'Tiered Percentage', 'Selects a percentage based on an income band');

-- ============================================================================
-- SEED: RULE DEFINITIONS
-- ============================================================================

INSERT INTO rule_definitions (code, name, category, description) VALUES
('PAYE', 'Pay As You Earn', 'INCOME_TAX', 'Monthly employment income tax'),
('PERSONAL_RELIEF', 'Personal Relief', 'RELIEF', 'Monthly personal income tax relief'),
('NSSF', 'National Social Security Fund', 'STATUTORY_DEDUCTION', 'Employee statutory pension contribution'),
('NHIF', 'National Hospital Insurance Fund', 'STATUTORY_DEDUCTION', 'Historical employee health insurance contribution'),
('SHIF', 'Social Health Insurance Fund', 'STATUTORY_DEDUCTION', 'Employee social health insurance contribution'),
('AHL', 'Affordable Housing Levy', 'STATUTORY_DEDUCTION', 'Employee affordable housing levy'),
('EMPLOYER_AHL', 'Employer Affordable Housing Levy', 'STATUTORY_CONTRIBUTION', 'Employer affordable housing levy contribution'),
('EMPLOYER_NSSF', 'Employer NSSF Contribution', 'STATUTORY_CONTRIBUTION', 'Employer statutory NSSF contribution'),
('QUALIFYING_PENSION', 'Qualifying Pension Contribution', 'ALLOWABLE_DEDUCTION', 'Tax-allowable pension contribution subject to applicable limits'),
('QUALIFYING_MORTGAGE_INTEREST', 'Qualifying Mortgage Interest', 'ALLOWABLE_DEDUCTION', 'Tax-allowable mortgage interest subject to applicable limits'),
('POST_RETIREMENT_MEDICAL_FUND', 'Post-Retirement Medical Fund', 'ALLOWABLE_DEDUCTION', 'Tax-allowable post-retirement medical fund contribution subject to applicable limits');

-- ============================================================================
-- IMPORTANT IMPLEMENTATION NOTES
--
-- 1. Do NOT overwrite a PUBLISHED rule_version. Create a new version.
-- 2. The Go API should select PUBLISHED rules where:
--      effective_from <= selected_pay_period_end
--      AND (effective_to IS NULL OR effective_to >= selected_pay_period_start)
--    with stricter monthly validation where a rule changes inside a month.
--
-- 3. For payroll periods spanning a legal change inside one month, define the
--    exact official payroll treatment in a dedicated version/source note rather
--    than assuming a simple date comparison.
--
-- 4. All government figures should be stored as data. Go implements only the
--    approved generic calculation methods.
--
-- 5. Before publishing a rule version, run rule_test_cases and compare the
--    results with independently verified expected results.
--
-- 6. Monetary values use DECIMAL, never FLOAT/DOUBLE.
-- ============================================================================
