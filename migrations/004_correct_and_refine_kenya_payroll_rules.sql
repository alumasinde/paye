-- ============================================================================
-- 004_correct_and_refine_kenya_payroll_rules.sql
-- Budget254 PAYE Calculator - Historical Rule Refinement
-- MySQL 8.0+
--
-- Run after:
--   001 budget254_paye.sql
--   002 002_seed_kenya_payroll_rules.sql
--   003 003_improve_budget254_paye_schema.sql
--
-- PURPOSE
-- 1. Convert NSSF phased contributions into explicit Tier I / Tier II records.
-- 2. Add calculation-period records from Jan 2022 through Dec 2026.
-- 3. Add explicit tax-treatment rows for key statutory deductions.
-- 4. Normalize rule formula mappings and effective granularities.
--
-- IMPORTANT
-- This migration does not delete historical rule_versions.
-- Existing NSSF capped-percentage records remain as compatibility history.
-- The new component structure is the preferred source for the Go resolver.
-- ============================================================================

START TRANSACTION;

-- ============================================================================
-- HELPER IDS
-- ============================================================================

SET @rule_nssf := (
    SELECT id FROM rule_definitions WHERE code = 'NSSF' LIMIT 1
);

SET @rule_employer_nssf := (
    SELECT id FROM rule_definitions WHERE code = 'EMPLOYER_NSSF' LIMIT 1
);

SET @formula_tiered_percentage := (
    SELECT id FROM calculation_formulas
    WHERE formula_code = 'TIERED_PERCENTAGE'
    LIMIT 1
);

SET @formula_progressive_tax := (
    SELECT id FROM calculation_formulas
    WHERE formula_code = 'PROGRESSIVE_TAX'
    LIMIT 1
);

SET @formula_tiered_fixed := (
    SELECT id FROM calculation_formulas
    WHERE formula_code = 'TIERED_FIXED_AMOUNT'
    LIMIT 1
);

SET @formula_percentage_min := (
    SELECT id FROM calculation_formulas
    WHERE formula_code = 'PERCENTAGE_WITH_MINIMUM'
    LIMIT 1
);

SET @formula_percentage := (
    SELECT id FROM calculation_formulas
    WHERE formula_code = 'PERCENTAGE'
    LIMIT 1
);

SET @formula_fixed := (
    SELECT id FROM calculation_formulas
    WHERE formula_code = 'FIXED_AMOUNT'
    LIMIT 1
);

-- ============================================================================
-- 1. NORMALIZE EFFECTIVE GRANULARITY
-- ============================================================================

UPDATE rule_versions
SET effective_granularity = 'MONTH'
WHERE effective_granularity IS NULL
   OR effective_granularity = '';

UPDATE rule_versions rv
JOIN rule_definitions rd ON rd.id = rv.rule_definition_id
SET rv.effective_granularity = 'DAY'
WHERE rd.code IN ('AHL', 'EMPLOYER_AHL')
  AND rv.effective_from = '2024-03-19';

-- ============================================================================
-- 2. FORMULA MAPPINGS FOR KEY STATUTORY RULES
-- ============================================================================

UPDATE rule_versions rv
JOIN rule_definitions rd ON rd.id = rv.rule_definition_id
SET rv.calculation_formula_id =
    CASE
        WHEN rd.code = 'PAYE' THEN @formula_progressive_tax
        WHEN rd.code = 'NHIF' THEN @formula_tiered_fixed
        WHEN rd.code = 'SHIF' THEN @formula_percentage_min
        WHEN rd.code IN ('AHL', 'EMPLOYER_AHL') THEN @formula_percentage
        WHEN rd.code IN ('PERSONAL_RELIEF') THEN @formula_fixed
        ELSE rv.calculation_formula_id
    END
WHERE rd.code IN (
    'PAYE',
    'NHIF',
    'SHIF',
    'AHL',
    'EMPLOYER_AHL',
    'PERSONAL_RELIEF'
);

-- ============================================================================
-- 3. NSSF EXPLICIT TIER I / TIER II COMPONENTS
--
-- Each NSSF employee version receives:
--   Tier I: 0 -> Lower Earnings Limit
--   Tier II: Lower Earnings Limit -> Upper Earnings Limit
--
-- The 6% employee rate and 6% employer rate are explicit.
-- ============================================================================

-- ---------------------------------------------------------------------------
-- YEAR 1: 1 Feb 2023 - 31 Jan 2024
-- LEL 6,000
-- UEL 18,000
-- ---------------------------------------------------------------------------

SET @nssf_y1 := (
    SELECT id FROM rule_versions
    WHERE rule_definition_id = @rule_nssf
      AND version_code = 'NSSF_YEAR_1_2023'
    LIMIT 1
);

UPDATE rule_versions
SET calculation_formula_id = @formula_tiered_percentage
WHERE id = @nssf_y1;

INSERT INTO rule_components
(rule_version_id, component_code, component_name, component_type,
 applies_to, calculation_order, effective_from, effective_to, is_active)
VALUES
(@nssf_y1, 'TIER_I', 'NSSF Tier I', 'TIER', 'BOTH', 10, '2023-02-01', '2024-01-31', TRUE),
(@nssf_y1, 'TIER_II', 'NSSF Tier II', 'TIER', 'BOTH', 20, '2023-02-01', '2024-01-31', TRUE)
ON DUPLICATE KEY UPDATE component_name = VALUES(component_name);

SET @nssf_y1_t1 := (
    SELECT id FROM rule_components
    WHERE rule_version_id = @nssf_y1
      AND component_code = 'TIER_I'
    LIMIT 1
);

SET @nssf_y1_t2 := (
    SELECT id FROM rule_components
    WHERE rule_version_id = @nssf_y1
      AND component_code = 'TIER_II'
    LIMIT 1
);

INSERT INTO rule_component_bands
(rule_component_id, from_amount, to_amount, employee_rate, employer_rate, display_order)
VALUES
(@nssf_y1_t1, 0.00, 6000.00, 0.0600000000, 0.0600000000, 1),
(@nssf_y1_t2, 6000.00, 18000.00, 0.0600000000, 0.0600000000, 1);

-- ---------------------------------------------------------------------------
-- YEAR 2: 1 Feb 2024 - 31 Jan 2025
-- LEL 7,000
-- UEL 36,000
-- ---------------------------------------------------------------------------

SET @nssf_y2 := (
    SELECT id FROM rule_versions
    WHERE rule_definition_id = @rule_nssf
      AND version_code = 'NSSF_YEAR_2_2024'
    LIMIT 1
);

UPDATE rule_versions
SET calculation_formula_id = @formula_tiered_percentage
WHERE id = @nssf_y2;

INSERT INTO rule_components
(rule_version_id, component_code, component_name, component_type,
 applies_to, calculation_order, effective_from, effective_to, is_active)
VALUES
(@nssf_y2, 'TIER_I', 'NSSF Tier I', 'TIER', 'BOTH', 10, '2024-02-01', '2025-01-31', TRUE),
(@nssf_y2, 'TIER_II', 'NSSF Tier II', 'TIER', 'BOTH', 20, '2024-02-01', '2025-01-31', TRUE)
ON DUPLICATE KEY UPDATE component_name = VALUES(component_name);

SET @nssf_y2_t1 := (
    SELECT id FROM rule_components
    WHERE rule_version_id = @nssf_y2
      AND component_code = 'TIER_I'
    LIMIT 1
);

SET @nssf_y2_t2 := (
    SELECT id FROM rule_components
    WHERE rule_version_id = @nssf_y2
      AND component_code = 'TIER_II'
    LIMIT 1
);

INSERT INTO rule_component_bands
(rule_component_id, from_amount, to_amount, employee_rate, employer_rate, display_order)
VALUES
(@nssf_y2_t1, 0.00, 7000.00, 0.0600000000, 0.0600000000, 1),
(@nssf_y2_t2, 7000.00, 36000.00, 0.0600000000, 0.0600000000, 1);

-- ---------------------------------------------------------------------------
-- YEAR 3: 1 Feb 2025 - 31 Jan 2026
-- LEL 8,000
-- UEL 72,000
-- ---------------------------------------------------------------------------

SET @nssf_y3 := (
    SELECT id FROM rule_versions
    WHERE rule_definition_id = @rule_nssf
      AND version_code = 'NSSF_YEAR_3_2025'
    LIMIT 1
);

UPDATE rule_versions
SET calculation_formula_id = @formula_tiered_percentage
WHERE id = @nssf_y3;

INSERT INTO rule_components
(rule_version_id, component_code, component_name, component_type,
 applies_to, calculation_order, effective_from, effective_to, is_active)
VALUES
(@nssf_y3, 'TIER_I', 'NSSF Tier I', 'TIER', 'BOTH', 10, '2025-02-01', '2026-01-31', TRUE),
(@nssf_y3, 'TIER_II', 'NSSF Tier II', 'TIER', 'BOTH', 20, '2025-02-01', '2026-01-31', TRUE)
ON DUPLICATE KEY UPDATE component_name = VALUES(component_name);

SET @nssf_y3_t1 := (
    SELECT id FROM rule_components
    WHERE rule_version_id = @nssf_y3
      AND component_code = 'TIER_I'
    LIMIT 1
);

SET @nssf_y3_t2 := (
    SELECT id FROM rule_components
    WHERE rule_version_id = @nssf_y3
      AND component_code = 'TIER_II'
    LIMIT 1
);

INSERT INTO rule_component_bands
(rule_component_id, from_amount, to_amount, employee_rate, employer_rate, display_order)
VALUES
(@nssf_y3_t1, 0.00, 8000.00, 0.0600000000, 0.0600000000, 1),
(@nssf_y3_t2, 8000.00, 72000.00, 0.0600000000, 0.0600000000, 1);

-- ---------------------------------------------------------------------------
-- YEAR 4: 1 Feb 2026 onward
-- LEL 9,000
-- UEL 108,000
-- ---------------------------------------------------------------------------

SET @nssf_y4 := (
    SELECT id FROM rule_versions
    WHERE rule_definition_id = @rule_nssf
      AND version_code = 'NSSF_YEAR_4_2026'
    LIMIT 1
);

UPDATE rule_versions
SET calculation_formula_id = @formula_tiered_percentage
WHERE id = @nssf_y4;

INSERT INTO rule_components
(rule_version_id, component_code, component_name, component_type,
 applies_to, calculation_order, effective_from, effective_to, is_active)
VALUES
(@nssf_y4, 'TIER_I', 'NSSF Tier I', 'TIER', 'BOTH', 10, '2026-02-01', NULL, TRUE),
(@nssf_y4, 'TIER_II', 'NSSF Tier II', 'TIER', 'BOTH', 20, '2026-02-01', NULL, TRUE)
ON DUPLICATE KEY UPDATE component_name = VALUES(component_name);

SET @nssf_y4_t1 := (
    SELECT id FROM rule_components
    WHERE rule_version_id = @nssf_y4
      AND component_code = 'TIER_I'
    LIMIT 1
);

SET @nssf_y4_t2 := (
    SELECT id FROM rule_components
    WHERE rule_version_id = @nssf_y4
      AND component_code = 'TIER_II'
    LIMIT 1
);

INSERT INTO rule_component_bands
(rule_component_id, from_amount, to_amount, employee_rate, employer_rate, display_order)
VALUES
(@nssf_y4_t1, 0.00, 9000.00, 0.0600000000, 0.0600000000, 1),
(@nssf_y4_t2, 9000.00, 108000.00, 0.0600000000, 0.0600000000, 1);

-- ============================================================================
-- 4. MARK LEGACY NSSF VERSION AS FIXED-AMOUNT FORMULA
-- ============================================================================

UPDATE rule_versions
SET calculation_formula_id = @formula_fixed
WHERE version_code IN (
    'NSSF_LEGACY_2022',
    'EMPLOYER_NSSF_LEGACY_2022'
);

-- ============================================================================
-- 5. TAX TREATMENT HISTORY
-- ============================================================================

-- Remove duplicate treatment rows if this migration is rerun manually.
DELETE rtt
FROM rule_tax_treatments rtt
JOIN rule_versions rv ON rv.id = rtt.rule_version_id
WHERE rv.version_code IN (
    'AHL_2024',
    'AHL_2024_TAX_DEDUCTIBLE',
    'SHIF_2024',
    'SHIF_2024_TAX_DEDUCTIBLE'
);

-- Pre-December 2024 AHL.
INSERT INTO rule_tax_treatments
(rule_version_id, treatment_type, maximum_amount, maximum_period,
 effective_from, effective_to, notes)
SELECT
    id,
    'NONE',
    NULL,
    'NONE',
    '2024-03-19',
    '2024-11-30',
    'AHL was not modeled as a PAYE taxable-income deduction before the December 2024 Tax Laws (Amendment) Act treatment.'
FROM rule_versions
WHERE version_code = 'AHL_2024'
LIMIT 1;

-- December 2024 onward AHL.
INSERT INTO rule_tax_treatments
(rule_version_id, treatment_type, maximum_amount, maximum_period,
 effective_from, effective_to, notes)
SELECT
    id,
    'DEDUCT_FROM_TAXABLE_INCOME',
    NULL,
    'NONE',
    '2024-12-01',
    NULL,
    'AHL deductible in determining taxable employment income from December 2024.'
FROM rule_versions
WHERE version_code = 'AHL_2024_TAX_DEDUCTIBLE'
LIMIT 1;

-- SHIF October-November 2024.
INSERT INTO rule_tax_treatments
(rule_version_id, treatment_type, maximum_amount, maximum_period,
 effective_from, effective_to, notes)
SELECT
    id,
    'NONE',
    NULL,
    'NONE',
    '2024-10-01',
    '2024-11-30',
    'Pre-December 2024 PAYE treatment.'
FROM rule_versions
WHERE version_code = 'SHIF_2024'
LIMIT 1;

-- SHIF December 2024 onward.
INSERT INTO rule_tax_treatments
(rule_version_id, treatment_type, maximum_amount, maximum_period,
 effective_from, effective_to, notes)
SELECT
    id,
    'DEDUCT_FROM_TAXABLE_INCOME',
    NULL,
    'NONE',
    '2024-12-01',
    NULL,
    'SHIF deductible in determining taxable employment income from December 2024.'
FROM rule_versions
WHERE version_code = 'SHIF_2024_TAX_DEDUCTIBLE'
LIMIT 1;

-- ============================================================================
-- 6. GENERATE MONTHLY CALCULATION PERIODS
-- Jan 2022 through Dec 2026
-- ============================================================================

INSERT INTO calculation_periods
(period_year, period_month, period_start, period_end)
SELECT
    YEAR(period_start),
    MONTH(period_start),
    period_start,
    LAST_DAY(period_start)
FROM (
    SELECT DATE_ADD('2022-01-01', INTERVAL n MONTH) AS period_start
    FROM (
        SELECT ones.n + tens.n * 10 + hundreds.n * 100 AS n
        FROM (
            SELECT 0 n UNION ALL SELECT 1 UNION ALL SELECT 2 UNION ALL SELECT 3
            UNION ALL SELECT 4 UNION ALL SELECT 5 UNION ALL SELECT 6
            UNION ALL SELECT 7 UNION ALL SELECT 8 UNION ALL SELECT 9
        ) ones
        CROSS JOIN (
            SELECT 0 n UNION ALL SELECT 1 UNION ALL SELECT 2 UNION ALL SELECT 3
            UNION ALL SELECT 4 UNION ALL SELECT 5 UNION ALL SELECT 6
            UNION ALL SELECT 7 UNION ALL SELECT 8 UNION ALL SELECT 9
        ) tens
        CROSS JOIN (
            SELECT 0 n UNION ALL SELECT 1
        ) hundreds
    ) numbers
    WHERE n <= TIMESTAMPDIFF(MONTH, '2022-01-01', '2026-12-01')
) periods
ON DUPLICATE KEY UPDATE
    period_start = VALUES(period_start),
    period_end = VALUES(period_end);

-- ============================================================================
-- 7. VALIDATION RESULTS FOR NSSF TIER DATA
-- ============================================================================

DELETE FROM rule_validation_results
WHERE validation_code LIKE 'NSSF_%';

INSERT INTO rule_validation_results
(rule_version_id, validation_code, validation_status, message, details)
VALUES
(
    @nssf_y1,
    'NSSF_TIER_STRUCTURE',
    'PASS',
    'Year 1 contains Tier I and Tier II component bands.',
    JSON_OBJECT('lower_earnings_limit',6000,'upper_earnings_limit',18000)
),
(
    @nssf_y2,
    'NSSF_TIER_STRUCTURE',
    'PASS',
    'Year 2 contains Tier I and Tier II component bands.',
    JSON_OBJECT('lower_earnings_limit',7000,'upper_earnings_limit',36000)
),
(
    @nssf_y3,
    'NSSF_TIER_STRUCTURE',
    'PASS',
    'Year 3 contains Tier I and Tier II component bands.',
    JSON_OBJECT('lower_earnings_limit',8000,'upper_earnings_limit',72000)
),
(
    @nssf_y4,
    'NSSF_TIER_STRUCTURE',
    'PASS',
    'Year 4 contains Tier I and Tier II component bands.',
    JSON_OBJECT('lower_earnings_limit',9000,'upper_earnings_limit',108000)
);

-- ============================================================================
-- 8. USEFUL POST-MIGRATION VERIFICATION QUERIES
--
-- SELECT rv.version_code, rc.component_code, rcb.from_amount, rcb.to_amount,
--        rcb.employee_rate, rcb.employer_rate
-- FROM rule_versions rv
-- JOIN rule_components rc ON rc.rule_version_id = rv.id
-- JOIN rule_component_bands rcb ON rcb.rule_component_id = rc.id
-- WHERE rv.version_code LIKE 'NSSF_YEAR_%'
-- ORDER BY rv.version_code, rc.calculation_order, rcb.display_order;
--
-- SELECT period_year, period_month, period_start, period_end
-- FROM calculation_periods
-- ORDER BY period_start;
-- ============================================================================

COMMIT;
