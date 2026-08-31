-- ============================================================================
-- 003_improve_budget254_paye_schema.sql
-- Budget254 PAYE Calculator - Schema Improvements
-- MySQL 8.0+
--
-- Run after:
--   001 budget254_paye.sql
--   002 002_seed_kenya_payroll_rules.sql
--
-- This migration is additive. It does not delete historical calculations.
-- AUDIT PATCH: calculation snapshots/rule sets use the BIGINT calculation IDs
-- defined by 001, and exact resolved rule membership is stored in set items.
-- ============================================================================

START TRANSACTION;

-- ============================================================================
-- 1. RULE VERSION EFFECTIVE DATE GRANULARITY
-- ============================================================================

ALTER TABLE rule_versions
    ADD COLUMN IF NOT EXISTS effective_granularity ENUM('DAY','MONTH','YEAR')
    NOT NULL DEFAULT 'MONTH'
    AFTER effective_to;

CREATE INDEX idx_rule_versions_effective_resolution
    ON rule_versions (status, effective_from, effective_to, effective_granularity);

-- ============================================================================
-- 2. CALCULATION FORMULAS
--
-- Go owns formula implementation.
-- MySQL chooses which formula a rule version uses and supplies its parameters.
-- ============================================================================

CREATE TABLE IF NOT EXISTS calculation_formulas (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    formula_code VARCHAR(100) NOT NULL,
    formula_name VARCHAR(150) NOT NULL,
    description TEXT NULL,
    input_schema JSON NOT NULL,
    output_schema JSON NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_calculation_formula_code (formula_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO calculation_formulas
(formula_code, formula_name, description, input_schema, output_schema, is_active)
VALUES
(
    'PROGRESSIVE_TAX',
    'Progressive Tax Bands',
    'Calculates tax progressively across ordered percentage bands.',
    JSON_OBJECT('base','DECIMAL','bands','ARRAY'),
    JSON_OBJECT('amount','DECIMAL','breakdown','ARRAY'),
    TRUE
),
(
    'TIERED_PERCENTAGE',
    'Tiered Percentage',
    'Calculates contributions across ordered statutory tiers.',
    JSON_OBJECT('base','DECIMAL','tiers','ARRAY'),
    JSON_OBJECT('employee_amount','DECIMAL','employer_amount','DECIMAL','breakdown','ARRAY'),
    TRUE
),
(
    'PERCENTAGE',
    'Percentage',
    'Calculates a percentage of a configured calculation base.',
    JSON_OBJECT('base','DECIMAL','rate','DECIMAL'),
    JSON_OBJECT('amount','DECIMAL'),
    TRUE
),
(
    'PERCENTAGE_WITH_MINIMUM',
    'Percentage With Minimum',
    'Calculates a percentage of a base subject to a minimum amount.',
    JSON_OBJECT('base','DECIMAL','rate','DECIMAL','minimum','DECIMAL'),
    JSON_OBJECT('amount','DECIMAL'),
    TRUE
),
(
    'PERCENTAGE_WITH_CAP',
    'Percentage With Cap',
    'Calculates a percentage subject to a maximum contribution.',
    JSON_OBJECT('base','DECIMAL','rate','DECIMAL','maximum','DECIMAL'),
    JSON_OBJECT('amount','DECIMAL'),
    TRUE
),
(
    'FIXED_AMOUNT',
    'Fixed Amount',
    'Returns a configured fixed amount.',
    JSON_OBJECT('amount','DECIMAL'),
    JSON_OBJECT('amount','DECIMAL'),
    TRUE
),
(
    'TIERED_FIXED_AMOUNT',
    'Tiered Fixed Amount',
    'Selects a fixed amount from the band containing the calculation base.',
    JSON_OBJECT('base','DECIMAL','bands','ARRAY'),
    JSON_OBJECT('amount','DECIMAL','selected_band','OBJECT'),
    TRUE
)
ON DUPLICATE KEY UPDATE
    formula_name = VALUES(formula_name),
    description = VALUES(description),
    input_schema = VALUES(input_schema),
    output_schema = VALUES(output_schema),
    is_active = VALUES(is_active);

-- Link rule versions to canonical formula implementations.
ALTER TABLE rule_versions
    ADD COLUMN IF NOT EXISTS calculation_formula_id BIGINT UNSIGNED NULL
    AFTER calculation_method_id,
    ADD CONSTRAINT fk_rule_versions_formula
        FOREIGN KEY (calculation_formula_id)
        REFERENCES calculation_formulas(id)
        ON DELETE SET NULL;

CREATE INDEX idx_rule_versions_formula
    ON rule_versions (calculation_formula_id);

-- ============================================================================
-- 3. RULE COMPONENTS
--
-- Used where one statutory rule has visible sub-components, such as:
-- NSSF -> Tier I + Tier II.
-- ============================================================================

CREATE TABLE IF NOT EXISTS rule_components (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    rule_version_id BIGINT UNSIGNED NOT NULL,
    component_code VARCHAR(100) NOT NULL,
    component_name VARCHAR(150) NOT NULL,
    component_type ENUM(
        'TIER',
        'BAND',
        'RELIEF',
        'DEDUCTION',
        'EMPLOYER_CONTRIBUTION'
    ) NOT NULL,
    applies_to ENUM('EMPLOYEE','EMPLOYER','BOTH')
        NOT NULL DEFAULT 'EMPLOYEE',
    calculation_order INT NOT NULL DEFAULT 100,
    effective_from DATE NOT NULL,
    effective_to DATE NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uq_rule_component_version_code (
        rule_version_id,
        component_code,
        effective_from
    ),

    KEY idx_rule_components_effective (
        rule_version_id,
        effective_from,
        effective_to,
        is_active
    ),

    CONSTRAINT fk_rule_components_rule_version
        FOREIGN KEY (rule_version_id)
        REFERENCES rule_versions(id)
        ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS rule_component_bands (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    rule_component_id BIGINT UNSIGNED NOT NULL,

    from_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
    to_amount DECIMAL(18,2) NULL,

    employee_rate DECIMAL(12,10) NULL,
    employer_rate DECIMAL(12,10) NULL,

    employee_fixed_amount DECIMAL(18,2) NULL,
    employer_fixed_amount DECIMAL(18,2) NULL,

    display_order INT NOT NULL DEFAULT 1,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,

    KEY idx_rule_component_bands_component (
        rule_component_id,
        display_order
    ),

    CONSTRAINT fk_rule_component_bands_component
        FOREIGN KEY (rule_component_id)
        REFERENCES rule_components(id)
        ON DELETE CASCADE,

    CONSTRAINT chk_component_band_range
        CHECK (
            to_amount IS NULL
            OR to_amount >= from_amount
        ),

    CONSTRAINT chk_component_band_has_value
        CHECK (
            employee_rate IS NOT NULL
            OR employer_rate IS NOT NULL
            OR employee_fixed_amount IS NOT NULL
            OR employer_fixed_amount IS NOT NULL
        )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- 4. TAX TREATMENT HISTORY
--
-- Contribution calculation and PAYE tax treatment are separate concepts.
-- ============================================================================

CREATE TABLE IF NOT EXISTS rule_tax_treatments (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    rule_version_id BIGINT UNSIGNED NOT NULL,

    treatment_type ENUM(
        'NONE',
        'DEDUCT_FROM_TAXABLE_INCOME',
        'TAX_RELIEF',
        'TAX_CREDIT'
    ) NOT NULL DEFAULT 'NONE',

    maximum_amount DECIMAL(18,2) NULL,

    maximum_period ENUM(
        'MONTHLY',
        'ANNUAL',
        'NONE'
    ) NOT NULL DEFAULT 'NONE',

    effective_from DATE NOT NULL,
    effective_to DATE NULL,

    notes TEXT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,

    KEY idx_rule_tax_treatments_effective (
        rule_version_id,
        effective_from,
        effective_to
    ),

    CONSTRAINT fk_rule_tax_treatments_version
        FOREIGN KEY (rule_version_id)
        REFERENCES rule_versions(id)
        ON DELETE CASCADE,

    CONSTRAINT chk_tax_treatment_dates
        CHECK (
            effective_to IS NULL
            OR effective_to >= effective_from
        )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- 5. CALCULATION PERIODS
--
-- Removes ambiguity when users select a historical month.
-- ============================================================================

CREATE TABLE IF NOT EXISTS calculation_periods (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,

    period_year SMALLINT UNSIGNED NOT NULL,
    period_month TINYINT UNSIGNED NOT NULL,

    period_start DATE NOT NULL,
    period_end DATE NOT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE KEY uq_calculation_period_year_month (
        period_year,
        period_month
    ),

    UNIQUE KEY uq_calculation_period_dates (
        period_start,
        period_end
    ),

    CONSTRAINT chk_calculation_period_month
        CHECK (period_month BETWEEN 1 AND 12),

    CONSTRAINT chk_calculation_period_dates
        CHECK (period_end >= period_start)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- 6. RULE VALIDATION RESULTS
-- ============================================================================

CREATE TABLE IF NOT EXISTS rule_validation_results (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,

    rule_version_id BIGINT UNSIGNED NOT NULL,

    validation_code VARCHAR(100) NOT NULL,

    validation_status ENUM('PASS','WARNING','FAIL')
        NOT NULL,

    message TEXT NOT NULL,

    details JSON NULL,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    KEY idx_rule_validation_version (
        rule_version_id,
        validation_status
    ),

    CONSTRAINT fk_rule_validation_rule_version
        FOREIGN KEY (rule_version_id)
        REFERENCES rule_versions(id)
        ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- 7. USER / CUSTOM DEDUCTION CATEGORIES
-- ============================================================================

CREATE TABLE IF NOT EXISTS deduction_categories (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,

    code VARCHAR(100) NOT NULL,
    name VARCHAR(150) NOT NULL,
    description TEXT NULL,

    affects_net_pay BOOLEAN NOT NULL DEFAULT TRUE,
    affects_taxable_income BOOLEAN NOT NULL DEFAULT FALSE,
    is_statutory BOOLEAN NOT NULL DEFAULT FALSE,

    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uq_deduction_category_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO deduction_categories
(code, name, description, affects_net_pay, affects_taxable_income, is_statutory)
VALUES
('STATUTORY', 'Statutory Deduction', 'Government or legally required deduction.', TRUE, FALSE, TRUE),
('SACCO', 'SACCO Contribution', 'Savings or contribution to a SACCO.', TRUE, FALSE, FALSE),
('WELFARE', 'Welfare Contribution', 'Employee welfare contribution.', TRUE, FALSE, FALSE),
('CHAMA', 'Chama Contribution', 'Savings or contribution to a Chama.', TRUE, FALSE, FALSE),
('SAVINGS', 'Savings', 'Personal savings deduction.', TRUE, FALSE, FALSE),
('LOAN', 'Loan Repayment', 'Loan repayment deducted from salary.', TRUE, FALSE, FALSE),
('SALARY_ADVANCE', 'Salary Advance', 'Recovery of a salary advance.', TRUE, FALSE, FALSE),
('INSURANCE', 'Insurance', 'Insurance deduction; tax treatment must be explicitly configured where applicable.', TRUE, FALSE, FALSE),
('UNION', 'Union Contribution', 'Union membership or contribution.', TRUE, FALSE, FALSE),
('OTHER', 'Other Deduction', 'Other user-defined deduction.', TRUE, FALSE, FALSE)
ON DUPLICATE KEY UPDATE
    name = VALUES(name),
    description = VALUES(description);

-- Optional category relationship for existing deduction_templates.
ALTER TABLE deduction_templates
    ADD COLUMN IF NOT EXISTS deduction_category_id BIGINT UNSIGNED NULL
    AFTER id;

ALTER TABLE deduction_templates
    ADD CONSTRAINT fk_deduction_templates_category
        FOREIGN KEY (deduction_category_id)
        REFERENCES deduction_categories(id)
        ON DELETE SET NULL;

CREATE INDEX idx_deduction_templates_category
    ON deduction_templates (deduction_category_id);

-- ============================================================================
-- 8. CALCULATION RULE SETS
--
-- Records exactly which rule versions were resolved for a calculation.
-- ============================================================================

CREATE TABLE IF NOT EXISTS calculation_rule_sets (
    id CHAR(36) NOT NULL PRIMARY KEY,

    calculation_id BIGINT UNSIGNED NOT NULL,

    calculation_period_id BIGINT UNSIGNED NULL,

    resolved_for_date DATE NOT NULL,

    rules_hash CHAR(64) NOT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE KEY uq_calculation_rule_set_calculation (
        calculation_id
    ),

    KEY idx_calculation_rule_sets_period (
        calculation_period_id
    ),

    CONSTRAINT fk_calculation_rule_sets_calculation
        FOREIGN KEY (calculation_id)
        REFERENCES calculations(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_calculation_rule_sets_period
        FOREIGN KEY (calculation_period_id)
        REFERENCES calculation_periods(id)
        ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Exact rule-version membership for reproducible historical calculations.
CREATE TABLE IF NOT EXISTS calculation_rule_set_items (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,

    calculation_rule_set_id CHAR(36) NOT NULL,

    rule_version_id BIGINT UNSIGNED NOT NULL,

    rule_definition_id BIGINT UNSIGNED NOT NULL,

    resolved_order INT NOT NULL DEFAULT 100,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE KEY uq_calculation_rule_set_rule (
        calculation_rule_set_id,
        rule_version_id
    ),

    KEY idx_calculation_rule_set_items_definition (
        rule_definition_id
    ),

    CONSTRAINT fk_calculation_rule_set_items_set
        FOREIGN KEY (calculation_rule_set_id)
        REFERENCES calculation_rule_sets(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_calculation_rule_set_items_version
        FOREIGN KEY (rule_version_id)
        REFERENCES rule_versions(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_calculation_rule_set_items_definition
        FOREIGN KEY (rule_definition_id)
        REFERENCES rule_definitions(id)
        ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- 9. FULL CALCULATION SNAPSHOTS
-- ============================================================================

CREATE TABLE IF NOT EXISTS calculation_snapshots (
    id CHAR(36) NOT NULL PRIMARY KEY,

    calculation_id BIGINT UNSIGNED NOT NULL,

    engine_version VARCHAR(50) NOT NULL,

    input_snapshot JSON NOT NULL,

    resolved_rules_snapshot JSON NOT NULL,

    calculation_steps JSON NOT NULL,

    output_snapshot JSON NOT NULL,

    checksum CHAR(64) NOT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE KEY uq_calculation_snapshot_calculation (
        calculation_id
    ),

    KEY idx_calculation_snapshots_checksum (
        checksum
    ),

    CONSTRAINT fk_calculation_snapshots_calculation
        FOREIGN KEY (calculation_id)
        REFERENCES calculations(id)
        ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- 10. MAP EXISTING CALCULATION METHODS TO CANONICAL FORMULAS
-- ============================================================================

UPDATE rule_versions rv
JOIN calculation_formulas cf
    ON cf.formula_code =
        CASE
            WHEN rv.calculation_method_id = (
                SELECT id FROM calculation_methods
                WHERE code = 'PROGRESSIVE_BANDS'
                LIMIT 1
            ) THEN 'PROGRESSIVE_TAX'

            WHEN rv.calculation_method_id = (
                SELECT id FROM calculation_methods
                WHERE code = 'TIERED_FIXED_AMOUNT'
                LIMIT 1
            ) THEN 'TIERED_FIXED_AMOUNT'

            WHEN rv.calculation_method_id = (
                SELECT id FROM calculation_methods
                WHERE code = 'PERCENTAGE_WITH_MINIMUM'
                LIMIT 1
            ) THEN 'PERCENTAGE_WITH_MINIMUM'

            WHEN rv.calculation_method_id = (
                SELECT id FROM calculation_methods
                WHERE code = 'CAPPED_PERCENTAGE'
                LIMIT 1
            ) THEN 'PERCENTAGE_WITH_CAP'

            WHEN rv.calculation_method_id = (
                SELECT id FROM calculation_methods
                WHERE code = 'PERCENTAGE'
                LIMIT 1
            ) THEN 'PERCENTAGE'

            WHEN rv.calculation_method_id = (
                SELECT id FROM calculation_methods
                WHERE code = 'FIXED_AMOUNT'
                LIMIT 1
            ) THEN 'FIXED_AMOUNT'

            ELSE NULL
        END
SET rv.calculation_formula_id = cf.id
WHERE rv.calculation_formula_id IS NULL;

COMMIT;
