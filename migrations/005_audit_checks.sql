-- ============================================================================
-- 005_audit_checks.sql
-- Budget254 PAYE Calculator - Post-migration consistency audit
-- MySQL 8.0+
--
-- Run after migrations 001-004.
-- The script is read-only: it returns result sets that should be empty for
-- failure queries, except for explicit summary queries.
-- ============================================================================

-- 1. Published versions with invalid date ranges.
SELECT 'INVALID_RULE_DATE_RANGE' AS audit, rv.*
FROM rule_versions rv
WHERE rv.status = 'PUBLISHED'
  AND rv.effective_to IS NOT NULL
  AND rv.effective_to < rv.effective_from;

-- 2. Overlapping published versions of the same rule definition.
SELECT
    'OVERLAPPING_PUBLISHED_RULES' AS audit,
    a.id AS rule_version_a_id,
    a.version_code AS rule_version_a,
    a.effective_from AS a_from,
    a.effective_to AS a_to,
    b.id AS rule_version_b_id,
    b.version_code AS rule_version_b,
    b.effective_from AS b_from,
    b.effective_to AS b_to
FROM rule_versions a
JOIN rule_versions b
  ON a.rule_definition_id = b.rule_definition_id
 AND a.id < b.id
WHERE a.status = 'PUBLISHED'
  AND b.status = 'PUBLISHED'
  AND a.effective_from <= COALESCE(b.effective_to, '9999-12-31')
  AND b.effective_from <= COALESCE(a.effective_to, '9999-12-31');

-- 3. Published rule versions missing a canonical formula mapping.
SELECT 'MISSING_FORMULA_MAPPING' AS audit, rd.code, rv.version_code
FROM rule_versions rv
JOIN rule_definitions rd ON rd.id = rv.rule_definition_id
LEFT JOIN calculation_formulas cf ON cf.id = rv.calculation_formula_id
WHERE rv.status = 'PUBLISHED'
  AND rv.calculation_formula_id IS NULL;

-- 4. NSSF phased versions must have exactly two active components.
SELECT
    'NSSF_COMPONENT_COUNT' AS audit,
    rv.version_code,
    COUNT(rc.id) AS component_count
FROM rule_versions rv
LEFT JOIN rule_components rc
  ON rc.rule_version_id = rv.id
 AND rc.is_active = TRUE
WHERE rv.version_code IN (
    'NSSF_YEAR_1_2023',
    'NSSF_YEAR_2_2024',
    'NSSF_YEAR_3_2025',
    'NSSF_YEAR_4_2026'
)
GROUP BY rv.id, rv.version_code
HAVING COUNT(rc.id) <> 2;

-- 5. NSSF components missing bands.
SELECT
    'NSSF_COMPONENT_MISSING_BAND' AS audit,
    rv.version_code,
    rc.component_code
FROM rule_versions rv
JOIN rule_components rc ON rc.rule_version_id = rv.id
LEFT JOIN rule_component_bands rcb ON rcb.rule_component_id = rc.id
WHERE rv.version_code IN (
    'NSSF_YEAR_1_2023',
    'NSSF_YEAR_2_2024',
    'NSSF_YEAR_3_2025',
    'NSSF_YEAR_4_2026'
)
GROUP BY rv.version_code, rc.id, rc.component_code
HAVING COUNT(rcb.id) = 0;

-- 6. NSSF exact expected limits/rates.
SELECT
    'NSSF_UNEXPECTED_LIMIT_OR_RATE' AS audit,
    rv.version_code,
    rc.component_code,
    rcb.from_amount,
    rcb.to_amount,
    rcb.employee_rate,
    rcb.employer_rate
FROM rule_versions rv
JOIN rule_components rc ON rc.rule_version_id = rv.id
JOIN rule_component_bands rcb ON rcb.rule_component_id = rc.id
WHERE
    (rv.version_code = 'NSSF_YEAR_1_2023' AND NOT (
        (rc.component_code='TIER_I' AND rcb.from_amount=0 AND rcb.to_amount=6000 AND rcb.employee_rate=0.06 AND rcb.employer_rate=0.06) OR
        (rc.component_code='TIER_II' AND rcb.from_amount=6000 AND rcb.to_amount=18000 AND rcb.employee_rate=0.06 AND rcb.employer_rate=0.06)
    ))
 OR (rv.version_code = 'NSSF_YEAR_2_2024' AND NOT (
        (rc.component_code='TIER_I' AND rcb.from_amount=0 AND rcb.to_amount=7000 AND rcb.employee_rate=0.06 AND rcb.employer_rate=0.06) OR
        (rc.component_code='TIER_II' AND rcb.from_amount=7000 AND rcb.to_amount=36000 AND rcb.employee_rate=0.06 AND rcb.employer_rate=0.06)
    ))
 OR (rv.version_code = 'NSSF_YEAR_3_2025' AND NOT (
        (rc.component_code='TIER_I' AND rcb.from_amount=0 AND rcb.to_amount=8000 AND rcb.employee_rate=0.06 AND rcb.employer_rate=0.06) OR
        (rc.component_code='TIER_II' AND rcb.from_amount=8000 AND rcb.to_amount=72000 AND rcb.employee_rate=0.06 AND rcb.employer_rate=0.06)
    ))
 OR (rv.version_code = 'NSSF_YEAR_4_2026' AND NOT (
        (rc.component_code='TIER_I' AND rcb.from_amount=0 AND rcb.to_amount=9000 AND rcb.employee_rate=0.06 AND rcb.employer_rate=0.06) OR
        (rc.component_code='TIER_II' AND rcb.from_amount=9000 AND rcb.to_amount=108000 AND rcb.employee_rate=0.06 AND rcb.employer_rate=0.06)
    ));

-- 7. SHIF/AHL tax-treatment periods must be non-overlapping.
SELECT
    'OVERLAPPING_TAX_TREATMENT' AS audit,
    a.rule_version_id AS version_a,
    b.rule_version_id AS version_b,
    a.effective_from,
    a.effective_to,
    b.effective_from,
    b.effective_to
FROM rule_tax_treatments a
JOIN rule_tax_treatments b
  ON a.rule_version_id = b.rule_version_id
 AND a.id < b.id
WHERE a.effective_from <= COALESCE(b.effective_to, '9999-12-31')
  AND b.effective_from <= COALESCE(a.effective_to, '9999-12-31');

-- 8. Monthly calculation periods must be contiguous.
WITH ordered AS (
    SELECT
        id,
        period_start,
        period_end,
        LAG(period_end) OVER (ORDER BY period_start) AS previous_end
    FROM calculation_periods
)
SELECT
    'CALCULATION_PERIOD_GAP_OR_OVERLAP' AS audit,
    id,
    period_start,
    period_end,
    previous_end
FROM ordered
WHERE previous_end IS NOT NULL
  AND period_start <> DATE_ADD(previous_end, INTERVAL 1 DAY);

-- 9. Custom deductions should not be tax-allowable merely because of category.
SELECT
    'CUSTOM_DEDUCTION_CATEGORY_MISMATCH' AS audit,
    dt.id,
    dt.name,
    dt.tax_treatment,
    dc.code AS category_code,
    dc.affects_taxable_income
FROM deduction_templates dt
LEFT JOIN deduction_categories dc ON dc.id = dt.deduction_category_id
WHERE dt.tax_treatment = 'TAX_ALLOWABLE'
  AND (dc.id IS NULL OR dc.affects_taxable_income = FALSE);

-- 10. Summary.
SELECT
    (SELECT COUNT(*) FROM rule_versions WHERE status='PUBLISHED') AS published_rule_versions,
    (SELECT COUNT(*) FROM calculation_periods) AS calculation_periods,
    (SELECT COUNT(*) FROM rule_components) AS rule_components,
    (SELECT COUNT(*) FROM rule_component_bands) AS rule_component_bands,
    (SELECT COUNT(*) FROM rule_tax_treatments) AS rule_tax_treatments;
