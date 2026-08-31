-- ============================================================================
-- 002_seed_kenya_payroll_rules.sql
-- Budget254 PAYE Calculator
-- Database: budget254_paye
-- MySQL 8.0+
--
-- VERIFIED BASELINE SEED DATA
-- Period covered: 2022-01-01 through current known rules at 2026-08-30.
--
-- IMPORTANT:
-- This migration is intentionally versioned and effective-dated.
-- Do not UPDATE published historical rows to change history. Insert a new
-- rule_version instead.
--
-- Sources represented in rule_sources include KRA, Kenya Law and NSSF/SHA.
-- Source URLs are stored as references for the admin portal.
--
-- IMPORTANT IMPLEMENTATION NOTE:
-- Affordable Housing Levy (AHL) 2023 enforcement was interrupted by litigation.
-- This seed marks the initial 2023 period as historical and separates the 2024
-- statutory regime. Before using the 2023 period for production payroll
-- reconciliation, verify the exact employer/payroll treatment for the specific
-- month being reconstructed.
-- ============================================================================

START TRANSACTION;

-- ============================================================================
-- HELPER VARIABLES
-- ============================================================================

SET @method_fixed := (
    SELECT id FROM calculation_methods WHERE code = 'FIXED_AMOUNT' LIMIT 1
);

SET @method_pct := (
    SELECT id FROM calculation_methods WHERE code = 'PERCENTAGE' LIMIT 1
);

SET @method_pct_min := (
    SELECT id FROM calculation_methods WHERE code = 'PERCENTAGE_WITH_MINIMUM' LIMIT 1
);

SET @method_capped_pct := (
    SELECT id FROM calculation_methods WHERE code = 'CAPPED_PERCENTAGE' LIMIT 1
);

SET @method_progressive := (
    SELECT id FROM calculation_methods WHERE code = 'PROGRESSIVE_BANDS' LIMIT 1
);

SET @method_tiered_fixed := (
    SELECT id FROM calculation_methods WHERE code = 'TIERED_FIXED_AMOUNT' LIMIT 1
);

-- ============================================================================
-- SOURCE DOCUMENTS
-- ============================================================================

INSERT INTO rule_sources
(authority, title, source_url, reference_number, published_on, accessed_on, notes)
VALUES
(
    'Kenya Revenue Authority',
    'Change of Tax Rates',
    'https://www.kra.go.ke/news-center/public-notices/1042-change-of-tax-rates',
    'Tax Laws (Amendment) No. 2 Act of 2020',
    '2020-12-24',
    '2026-08-30',
    'PAYE rates effective 1 January 2021 and therefore applicable at the start of the Budget254 historical range.'
),
(
    'Kenya Revenue Authority',
    'Individual Income Tax',
    'https://www.kra.go.ke/individual/filing-paying/types-of-taxes/individual-income-tax',
    'Finance Act 2023 PAYE guidance',
    '2023-07-01',
    '2026-08-30',
    'Current KRA individual income tax bands and personal relief.'
),
(
    'Kenya Revenue Authority',
    'Collection of Affordable Housing Levy by Kenya Revenue Authority',
    'https://www.kra.go.ke/news-center/public-notices/1974-collection-of-affordable-housing-levy-by-kenya-revenue-authority',
    'Finance Act 2023 AHL',
    '2023-08-04',
    '2026-08-30',
    'KRA notice stating 1.5% employee and 1.5% employer contributions effective 1 July 2023.'
),
(
    'Kenya Revenue Authority',
    'Collection of Affordable Housing Levy by Kenya Revenue Authority',
    'https://www.kra.go.ke/news-center/public-notices/2099-',
    'Affordable Housing Act 2024 AHL',
    '2024-03-21',
    '2026-08-30',
    'KRA notice stating the Affordable Housing Act 2024 collection effective 19 March 2024.'
),
(
    'Kenya Revenue Authority',
    'Amendments to PAYE Computation Pursuant to the Tax Laws (Amendment) Act, 2024',
    'https://www.kra.go.ke/news-center/public-notices/2157-amendments-to-paye-computation-pursuant-to-the-tax-laws-amendment-act%2C-2024',
    'Tax Laws (Amendment) Act 2024',
    '2024-12-27',
    '2026-08-30',
    'PAYE computation changes applicable to December 2024 and subsequent periods.'
),
(
    'Kenya Law',
    'National Health Insurance Fund (Standard and Special Contributions) Regulations',
    'https://new.kenyalaw.org/akn/ke/act/ln/2015/14/eng%402022-12-31',
    'Legal Notice 14 of 2015',
    '2015-02-13',
    '2026-08-30',
    'Historical graduated NHIF employee contribution schedule.'
),
(
    'Kenya Law',
    'Social Health Insurance Regulations, 2024',
    'https://new.kenyalaw.org/akn/ke/act/ln/2024/49/eng%402024-03-08',
    'Legal Notice 49 of 2024',
    '2024-03-08',
    '2026-08-30',
    'Salaried household contribution is 2.75% of gross salary/wage with minimum KES 300.'
),
(
    'NSSF Kenya',
    'New NSSF Rates',
    'https://www.nssf.or.ke/download/new-nssf-rates',
    'NSSF Act 2013 phased implementation',
    '2023-01-13',
    '2026-08-30',
    'Official NSSF rates download reference.'
),
(
    'NSSF Kenya',
    'New NSSF Rates Computation Guidelines',
    'https://www.nssf.or.ke/download/new-nssf-rates-computation-guidelines',
    'NSSF computation guidelines',
    '2023-01-13',
    '2026-08-30',
    'Official NSSF computation guideline reference.'
);

-- Source IDs
SET @src_paye_2021 := (
    SELECT id FROM rule_sources
    WHERE source_url = 'https://www.kra.go.ke/news-center/public-notices/1042-change-of-tax-rates'
    LIMIT 1
);

SET @src_paye_2023 := (
    SELECT id FROM rule_sources
    WHERE source_url = 'https://www.kra.go.ke/individual/filing-paying/types-of-taxes/individual-income-tax'
    LIMIT 1
);

SET @src_ahl_2023 := (
    SELECT id FROM rule_sources
    WHERE source_url = 'https://www.kra.go.ke/news-center/public-notices/1974-collection-of-affordable-housing-levy-by-kenya-revenue-authority'
    LIMIT 1
);

SET @src_ahl_2024 := (
    SELECT id FROM rule_sources
    WHERE source_url = 'https://www.kra.go.ke/news-center/public-notices/2099-'
    LIMIT 1
);

SET @src_tax_2024 := (
    SELECT id FROM rule_sources
    WHERE source_url LIKE '%2157-amendments-to-paye-computation%'
    LIMIT 1
);

SET @src_nhif := (
    SELECT id FROM rule_sources
    WHERE authority = 'Kenya Law'
      AND title LIKE 'National Health Insurance Fund%'
    LIMIT 1
);

SET @src_shif := (
    SELECT id FROM rule_sources
    WHERE authority = 'Kenya Law'
      AND title LIKE 'Social Health Insurance Regulations%'
    LIMIT 1
);

SET @src_nssf := (
    SELECT id FROM rule_sources
    WHERE authority = 'NSSF Kenya'
      AND title = 'New NSSF Rates'
    LIMIT 1
);

-- ============================================================================
-- RULE DEFINITION IDS
-- ============================================================================

SET @rule_paye := (SELECT id FROM rule_definitions WHERE code = 'PAYE' LIMIT 1);
SET @rule_personal_relief := (SELECT id FROM rule_definitions WHERE code = 'PERSONAL_RELIEF' LIMIT 1);
SET @rule_nssf := (SELECT id FROM rule_definitions WHERE code = 'NSSF' LIMIT 1);
SET @rule_nhif := (SELECT id FROM rule_definitions WHERE code = 'NHIF' LIMIT 1);
SET @rule_shif := (SELECT id FROM rule_definitions WHERE code = 'SHIF' LIMIT 1);
SET @rule_ahl := (SELECT id FROM rule_definitions WHERE code = 'AHL' LIMIT 1);
SET @rule_employer_ahl := (SELECT id FROM rule_definitions WHERE code = 'EMPLOYER_AHL' LIMIT 1);
SET @rule_employer_nssf := (SELECT id FROM rule_definitions WHERE code = 'EMPLOYER_NSSF' LIMIT 1);
SET @rule_pension := (SELECT id FROM rule_definitions WHERE code = 'QUALIFYING_PENSION' LIMIT 1);
SET @rule_mortgage := (SELECT id FROM rule_definitions WHERE code = 'QUALIFYING_MORTGAGE_INTEREST' LIMIT 1);
SET @rule_medical_fund := (SELECT id FROM rule_definitions WHERE code = 'POST_RETIREMENT_MEDICAL_FUND' LIMIT 1);

-- ============================================================================
-- PAYE: 1 JANUARY 2021 TO 30 JUNE 2023
-- Applicable to all Budget254 historical calculations from Jan 2022 through
-- June 2023.
--
-- First 24,000 @ 10%
-- Next 8,333 @ 25%
-- Balance above 32,333 @ 30%
-- ============================================================================

INSERT INTO rule_versions
(
    rule_definition_id,
    calculation_method_id,
    version_code,
    version_name,
    status,
    effective_from,
    effective_to,
    calculation_order,
    affects_taxable_income,
    affects_net_pay,
    source_summary,
    notes,
    published_at
)
VALUES
(
    @rule_paye,
    @method_progressive,
    'PAYE_2021',
    'PAYE Bands Effective 1 January 2021',
    'PUBLISHED',
    '2021-01-01',
    '2023-06-30',
    100,
    0,
    1,
    'KRA Change of Tax Rates effective 1 January 2021',
    'Historical PAYE version used for all calculations from January 2022 to June 2023.',
    CURRENT_TIMESTAMP
);

SET @paye_2021 := LAST_INSERT_ID();

INSERT INTO rule_bands
(rule_version_id, from_amount, to_amount, rate, fixed_amount, display_order, label)
VALUES
(@paye_2021, 0.00, 24000.00, 0.10000000, NULL, 1, 'First KES 24,000 at 10%'),
(@paye_2021, 24000.00, 32333.00, 0.25000000, NULL, 2, 'Next KES 8,333 at 25%'),
(@paye_2021, 32333.00, NULL, 0.30000000, NULL, 3, 'Income above KES 32,333 at 30%');

INSERT INTO rule_version_sources (rule_version_id, rule_source_id)
VALUES (@paye_2021, @src_paye_2021);

-- ============================================================================
-- PAYE: EFFECTIVE 1 JULY 2023
--
-- First 24,000 @ 10%
-- Next 8,333 @ 25%
-- Next 467,667 @ 30%
-- Next 300,000 @ 32.5%
-- Balance above 800,000 @ 35%
-- ============================================================================

INSERT INTO rule_versions
(
    rule_definition_id,
    calculation_method_id,
    version_code,
    version_name,
    status,
    effective_from,
    effective_to,
    calculation_order,
    affects_taxable_income,
    affects_net_pay,
    source_summary,
    notes,
    published_at
)
VALUES
(
    @rule_paye,
    @method_progressive,
    'PAYE_2023_JULY',
    'PAYE Bands Effective 1 July 2023',
    'PUBLISHED',
    '2023-07-01',
    NULL,
    100,
    0,
    1,
    'KRA individual income tax rates effective 1 July 2023',
    'Current PAYE band structure as represented by KRA.',
    CURRENT_TIMESTAMP
);

SET @paye_2023 := LAST_INSERT_ID();

INSERT INTO rule_bands
(rule_version_id, from_amount, to_amount, rate, fixed_amount, display_order, label)
VALUES
(@paye_2023, 0.00, 24000.00, 0.10000000, NULL, 1, 'First KES 24,000 at 10%'),
(@paye_2023, 24000.00, 32333.00, 0.25000000, NULL, 2, 'Next KES 8,333 at 25%'),
(@paye_2023, 32333.00, 500000.00, 0.30000000, NULL, 3, 'Next KES 467,667 at 30%'),
(@paye_2023, 500000.00, 800000.00, 0.32500000, NULL, 4, 'Next KES 300,000 at 32.5%'),
(@paye_2023, 800000.00, NULL, 0.35000000, NULL, 5, 'Income above KES 800,000 at 35%');

INSERT INTO rule_version_sources (rule_version_id, rule_source_id)
VALUES (@paye_2023, @src_paye_2023);

-- ============================================================================
-- PERSONAL RELIEF
-- KES 2,400 monthly.
-- ============================================================================

INSERT INTO rule_versions
(
    rule_definition_id,
    calculation_method_id,
    version_code,
    version_name,
    status,
    effective_from,
    effective_to,
    calculation_order,
    affects_taxable_income,
    affects_net_pay,
    source_summary,
    notes,
    published_at
)
VALUES
(
    @rule_personal_relief,
    @method_fixed,
    'PERSONAL_RELIEF_2021',
    'Personal Relief KES 2,400 Monthly',
    'PUBLISHED',
    '2021-01-01',
    NULL,
    900,
    0,
    1,
    'KRA personal relief guidance',
    'Applicable to resident individuals subject to eligibility.',
    CURRENT_TIMESTAMP
);

SET @personal_relief := LAST_INSERT_ID();

INSERT INTO rule_parameters
(rule_version_id, parameter_name, parameter_type, value_decimal)
VALUES
(@personal_relief, 'amount', 'DECIMAL', 2400.00);

INSERT INTO rule_version_sources (rule_version_id, rule_source_id)
VALUES
(@personal_relief, @src_paye_2021),
(@personal_relief, @src_paye_2023);

-- ============================================================================
-- NSSF LEGACY FLAT CONTRIBUTION
-- Used as the baseline historical model before phased NSSF implementation.
--
-- Employee: KES 200
-- Employer: KES 200
--
-- IMPORTANT:
-- If a specific employer had a lawful contracted-out or different historical
-- arrangement, this generic calculator result may differ.
-- ============================================================================

INSERT INTO rule_versions
(
    rule_definition_id, calculation_method_id, version_code, version_name,
    status, effective_from, effective_to, calculation_order,
    affects_taxable_income, affects_net_pay, source_summary, notes, published_at
)
VALUES
(
    @rule_nssf, @method_fixed, 'NSSF_LEGACY_2022',
    'Legacy NSSF Employee Contribution',
    'PUBLISHED', '2022-01-01', '2023-01-31', 200,
    1, 1,
    'Historical legacy NSSF payroll contribution baseline',
    'Generic historical employee contribution baseline before phased NSSF rates.',
    CURRENT_TIMESTAMP
);

SET @nssf_legacy_employee := LAST_INSERT_ID();

INSERT INTO rule_parameters
(rule_version_id, parameter_name, parameter_type, value_decimal)
VALUES
(@nssf_legacy_employee, 'amount', 'DECIMAL', 200.00),
(@nssf_legacy_employee, 'calculation_base', 'TEXT', NULL);

INSERT INTO rule_versions
(
    rule_definition_id, calculation_method_id, version_code, version_name,
    status, effective_from, effective_to, calculation_order,
    affects_taxable_income, affects_net_pay, source_summary, notes, published_at
)
VALUES
(
    @rule_employer_nssf, @method_fixed, 'EMPLOYER_NSSF_LEGACY_2022',
    'Legacy Employer NSSF Contribution',
    'PUBLISHED', '2022-01-01', '2023-01-31', 200,
    0, 0,
    'Historical legacy NSSF employer contribution baseline',
    'Employer contribution shown separately and does not reduce employee net pay.',
    CURRENT_TIMESTAMP
);

SET @nssf_legacy_employer := LAST_INSERT_ID();

INSERT INTO rule_parameters
(rule_version_id, parameter_name, parameter_type, value_decimal)
VALUES
(@nssf_legacy_employer, 'amount', 'DECIMAL', 200.00);

-- ============================================================================
-- NSSF PHASED CONTRIBUTIONS
--
-- Employee and employer are stored separately so the employee net-pay engine
-- does not accidentally deduct the employer contribution.
--
-- Contribution = 6% of pensionable earnings subject to the configured upper
-- earnings limit. The lower earnings limit is retained as a rule parameter for
-- statutory transparency and future calculation-engine validation.
-- ============================================================================

-- YEAR 1: Feb 2023 - Jan 2024
INSERT INTO rule_versions
(
    rule_definition_id, calculation_method_id, version_code, version_name,
    status, effective_from, effective_to, calculation_order,
    affects_taxable_income, affects_net_pay, source_summary, notes, published_at
)
VALUES
(
    @rule_nssf, @method_capped_pct, 'NSSF_YEAR_1_2023',
    'NSSF Year 1 Employee Contribution',
    'PUBLISHED', '2023-02-01', '2024-01-31', 200,
    1, 1,
    'NSSF phased implementation Year 1',
    '6% employee contribution using the Year 1 upper earnings limit.',
    CURRENT_TIMESTAMP
);

SET @nssf_y1_employee := LAST_INSERT_ID();

INSERT INTO rule_parameters
(rule_version_id, parameter_name, parameter_type, value_decimal)
VALUES
(@nssf_y1_employee, 'rate', 'DECIMAL', 0.06000000),
(@nssf_y1_employee, 'lower_earnings_limit', 'DECIMAL', 6000.00),
(@nssf_y1_employee, 'upper_earnings_limit', 'DECIMAL', 18000.00),
(@nssf_y1_employee, 'maximum_contribution', 'DECIMAL', 1080.00),
(@nssf_y1_employee, 'calculation_base', 'TEXT', NULL);

UPDATE rule_parameters
SET value_text = 'PENSIONABLE_EARNINGS'
WHERE rule_version_id = @nssf_y1_employee
  AND parameter_name = 'calculation_base';

INSERT INTO rule_version_sources (rule_version_id, rule_source_id) VALUES (@nssf_y1_employee, @src_nssf);

INSERT INTO rule_versions
(
    rule_definition_id, calculation_method_id, version_code, version_name,
    status, effective_from, effective_to, calculation_order,
    affects_taxable_income, affects_net_pay, source_summary, notes, published_at
)
VALUES
(
    @rule_employer_nssf, @method_capped_pct, 'EMPLOYER_NSSF_YEAR_1_2023',
    'NSSF Year 1 Employer Contribution',
    'PUBLISHED', '2023-02-01', '2024-01-31', 200,
    0, 0,
    'NSSF phased implementation Year 1',
    'Employer contribution stored separately.',
    CURRENT_TIMESTAMP
);

SET @nssf_y1_employer := LAST_INSERT_ID();

INSERT INTO rule_parameters
(rule_version_id, parameter_name, parameter_type, value_decimal, value_text)
VALUES
(@nssf_y1_employer, 'rate', 'DECIMAL', 0.06000000, NULL),
(@nssf_y1_employer, 'lower_earnings_limit', 'DECIMAL', 6000.00, NULL),
(@nssf_y1_employer, 'upper_earnings_limit', 'DECIMAL', 18000.00, NULL),
(@nssf_y1_employer, 'maximum_contribution', 'DECIMAL', 1080.00, NULL),
(@nssf_y1_employer, 'calculation_base', 'TEXT', NULL, 'PENSIONABLE_EARNINGS');

INSERT INTO rule_version_sources (rule_version_id, rule_source_id) VALUES (@nssf_y1_employer, @src_nssf);

-- YEAR 2: Feb 2024 - Jan 2025
INSERT INTO rule_versions
(
    rule_definition_id, calculation_method_id, version_code, version_name,
    status, effective_from, effective_to, calculation_order,
    affects_taxable_income, affects_net_pay, source_summary, notes, published_at
)
VALUES
(
    @rule_nssf, @method_capped_pct, 'NSSF_YEAR_2_2024',
    'NSSF Year 2 Employee Contribution',
    'PUBLISHED', '2024-02-01', '2025-01-31', 200,
    1, 1,
    'NSSF phased implementation Year 2',
    '6% employee contribution using the Year 2 upper earnings limit.',
    CURRENT_TIMESTAMP
);

SET @nssf_y2_employee := LAST_INSERT_ID();

INSERT INTO rule_parameters
(rule_version_id, parameter_name, parameter_type, value_decimal, value_text)
VALUES
(@nssf_y2_employee, 'rate', 'DECIMAL', 0.06000000, NULL),
(@nssf_y2_employee, 'lower_earnings_limit', 'DECIMAL', 7000.00, NULL),
(@nssf_y2_employee, 'upper_earnings_limit', 'DECIMAL', 36000.00, NULL),
(@nssf_y2_employee, 'maximum_contribution', 'DECIMAL', 2160.00, NULL),
(@nssf_y2_employee, 'calculation_base', 'TEXT', NULL, 'PENSIONABLE_EARNINGS');

INSERT INTO rule_version_sources (rule_version_id, rule_source_id) VALUES (@nssf_y2_employee, @src_nssf);

INSERT INTO rule_versions
(
    rule_definition_id, calculation_method_id, version_code, version_name,
    status, effective_from, effective_to, calculation_order,
    affects_taxable_income, affects_net_pay, source_summary, notes, published_at
)
VALUES
(
    @rule_employer_nssf, @method_capped_pct, 'EMPLOYER_NSSF_YEAR_2_2024',
    'NSSF Year 2 Employer Contribution',
    'PUBLISHED', '2024-02-01', '2025-01-31', 200,
    0, 0,
    'NSSF phased implementation Year 2',
    'Employer contribution stored separately.',
    CURRENT_TIMESTAMP
);

SET @nssf_y2_employer := LAST_INSERT_ID();

INSERT INTO rule_parameters
(rule_version_id, parameter_name, parameter_type, value_decimal, value_text)
VALUES
(@nssf_y2_employer, 'rate', 'DECIMAL', 0.06000000, NULL),
(@nssf_y2_employer, 'lower_earnings_limit', 'DECIMAL', 7000.00, NULL),
(@nssf_y2_employer, 'upper_earnings_limit', 'DECIMAL', 36000.00, NULL),
(@nssf_y2_employer, 'maximum_contribution', 'DECIMAL', 2160.00, NULL),
(@nssf_y2_employer, 'calculation_base', 'TEXT', NULL, 'PENSIONABLE_EARNINGS');

INSERT INTO rule_version_sources (rule_version_id, rule_source_id) VALUES (@nssf_y2_employer, @src_nssf);

-- YEAR 3: Feb 2025 - Jan 2026
INSERT INTO rule_versions
(
    rule_definition_id, calculation_method_id, version_code, version_name,
    status, effective_from, effective_to, calculation_order,
    affects_taxable_income, affects_net_pay, source_summary, notes, published_at
)
VALUES
(
    @rule_nssf, @method_capped_pct, 'NSSF_YEAR_3_2025',
    'NSSF Year 3 Employee Contribution',
    'PUBLISHED', '2025-02-01', '2026-01-31', 200,
    1, 1,
    'NSSF phased implementation Year 3',
    '6% employee contribution using the Year 3 upper earnings limit.',
    CURRENT_TIMESTAMP
);

SET @nssf_y3_employee := LAST_INSERT_ID();

INSERT INTO rule_parameters
(rule_version_id, parameter_name, parameter_type, value_decimal, value_text)
VALUES
(@nssf_y3_employee, 'rate', 'DECIMAL', 0.06000000, NULL),
(@nssf_y3_employee, 'lower_earnings_limit', 'DECIMAL', 8000.00, NULL),
(@nssf_y3_employee, 'upper_earnings_limit', 'DECIMAL', 72000.00, NULL),
(@nssf_y3_employee, 'maximum_contribution', 'DECIMAL', 4320.00, NULL),
(@nssf_y3_employee, 'calculation_base', 'TEXT', NULL, 'PENSIONABLE_EARNINGS');

INSERT INTO rule_version_sources (rule_version_id, rule_source_id) VALUES (@nssf_y3_employee, @src_nssf);

INSERT INTO rule_versions
(
    rule_definition_id, calculation_method_id, version_code, version_name,
    status, effective_from, effective_to, calculation_order,
    affects_taxable_income, affects_net_pay, source_summary, notes, published_at
)
VALUES
(
    @rule_employer_nssf, @method_capped_pct, 'EMPLOYER_NSSF_YEAR_3_2025',
    'NSSF Year 3 Employer Contribution',
    'PUBLISHED', '2025-02-01', '2026-01-31', 200,
    0, 0,
    'NSSF phased implementation Year 3',
    'Employer contribution stored separately.',
    CURRENT_TIMESTAMP
);

SET @nssf_y3_employer := LAST_INSERT_ID();

INSERT INTO rule_parameters
(rule_version_id, parameter_name, parameter_type, value_decimal, value_text)
VALUES
(@nssf_y3_employer, 'rate', 'DECIMAL', 0.06000000, NULL),
(@nssf_y3_employer, 'lower_earnings_limit', 'DECIMAL', 8000.00, NULL),
(@nssf_y3_employer, 'upper_earnings_limit', 'DECIMAL', 72000.00, NULL),
(@nssf_y3_employer, 'maximum_contribution', 'DECIMAL', 4320.00, NULL),
(@nssf_y3_employer, 'calculation_base', 'TEXT', NULL, 'PENSIONABLE_EARNINGS');

INSERT INTO rule_version_sources (rule_version_id, rule_source_id) VALUES (@nssf_y3_employer, @src_nssf);

-- YEAR 4: Feb 2026 onward
INSERT INTO rule_versions
(
    rule_definition_id, calculation_method_id, version_code, version_name,
    status, effective_from, effective_to, calculation_order,
    affects_taxable_income, affects_net_pay, source_summary, notes, published_at
)
VALUES
(
    @rule_nssf, @method_capped_pct, 'NSSF_YEAR_4_2026',
    'NSSF Year 4 Employee Contribution',
    'PUBLISHED', '2026-02-01', NULL, 200,
    1, 1,
    'NSSF phased implementation Year 4',
    '6% employee contribution using the Year 4 upper earnings limit.',
    CURRENT_TIMESTAMP
);

SET @nssf_y4_employee := LAST_INSERT_ID();

INSERT INTO rule_parameters
(rule_version_id, parameter_name, parameter_type, value_decimal, value_text)
VALUES
(@nssf_y4_employee, 'rate', 'DECIMAL', 0.06000000, NULL),
(@nssf_y4_employee, 'lower_earnings_limit', 'DECIMAL', 9000.00, NULL),
(@nssf_y4_employee, 'upper_earnings_limit', 'DECIMAL', 108000.00, NULL),
(@nssf_y4_employee, 'maximum_contribution', 'DECIMAL', 6480.00, NULL),
(@nssf_y4_employee, 'calculation_base', 'TEXT', NULL, 'PENSIONABLE_EARNINGS');

INSERT INTO rule_version_sources (rule_version_id, rule_source_id) VALUES (@nssf_y4_employee, @src_nssf);

INSERT INTO rule_versions
(
    rule_definition_id, calculation_method_id, version_code, version_name,
    status, effective_from, effective_to, calculation_order,
    affects_taxable_income, affects_net_pay, source_summary, notes, published_at
)
VALUES
(
    @rule_employer_nssf, @method_capped_pct, 'EMPLOYER_NSSF_YEAR_4_2026',
    'NSSF Year 4 Employer Contribution',
    'PUBLISHED', '2026-02-01', NULL, 200,
    0, 0,
    'NSSF phased implementation Year 4',
    'Employer contribution stored separately.',
    CURRENT_TIMESTAMP
);

SET @nssf_y4_employer := LAST_INSERT_ID();

INSERT INTO rule_parameters
(rule_version_id, parameter_name, parameter_type, value_decimal, value_text)
VALUES
(@nssf_y4_employer, 'rate', 'DECIMAL', 0.06000000, NULL),
(@nssf_y4_employer, 'lower_earnings_limit', 'DECIMAL', 9000.00, NULL),
(@nssf_y4_employer, 'upper_earnings_limit', 'DECIMAL', 108000.00, NULL),
(@nssf_y4_employer, 'maximum_contribution', 'DECIMAL', 6480.00, NULL),
(@nssf_y4_employer, 'calculation_base', 'TEXT', NULL, 'PENSIONABLE_EARNINGS');

INSERT INTO rule_version_sources (rule_version_id, rule_source_id) VALUES (@nssf_y4_employer, @src_nssf);

-- ============================================================================
-- NHIF: HISTORICAL GRADUATED EMPLOYEE CONTRIBUTION
-- Used for Budget254 historical periods through the end of September 2024.
--
-- The effective end date is modeled as 30 September 2024 for the app's
-- month-based PAYE history, with SHIF beginning October 2024.
-- ============================================================================

INSERT INTO rule_versions
(
    rule_definition_id, calculation_method_id, version_code, version_name,
    status, effective_from, effective_to, calculation_order,
    affects_taxable_income, affects_net_pay, source_summary, notes, published_at
)
VALUES
(
    @rule_nhif, @method_tiered_fixed, 'NHIF_2015_HISTORICAL',
    'NHIF Historical Graduated Contribution Bands',
    'PUBLISHED', '2015-04-01', '2024-09-30', 300,
    0, 1,
    'Legal Notice 14 of 2015 NHIF graduated contribution schedule',
    'Used for the Budget254 historical PAYE range until SHIF became the payroll health deduction model in October 2024.',
    CURRENT_TIMESTAMP
);

SET @nhif_historical := LAST_INSERT_ID();

INSERT INTO rule_bands
(rule_version_id, from_amount, to_amount, rate, fixed_amount, display_order, label)
VALUES
(@nhif_historical, 0.00, 5999.00, NULL, 150.00, 1, 'Up to KES 5,999'),
(@nhif_historical, 6000.00, 7999.00, NULL, 300.00, 2, 'KES 6,000 to 7,999'),
(@nhif_historical, 8000.00, 11999.00, NULL, 400.00, 3, 'KES 8,000 to 11,999'),
(@nhif_historical, 12000.00, 14999.00, NULL, 500.00, 4, 'KES 12,000 to 14,999'),
(@nhif_historical, 15000.00, 19999.00, NULL, 600.00, 5, 'KES 15,000 to 19,999'),
(@nhif_historical, 20000.00, 24999.00, NULL, 750.00, 6, 'KES 20,000 to 24,999'),
(@nhif_historical, 25000.00, 29999.00, NULL, 850.00, 7, 'KES 25,000 to 29,999'),
(@nhif_historical, 30000.00, 34999.00, NULL, 900.00, 8, 'KES 30,000 to 34,999'),
(@nhif_historical, 35000.00, 39999.00, NULL, 950.00, 9, 'KES 35,000 to 39,999'),
(@nhif_historical, 40000.00, 44999.00, NULL, 1000.00, 10, 'KES 40,000 to 44,999'),
(@nhif_historical, 45000.00, 49999.00, NULL, 1100.00, 11, 'KES 45,000 to 49,999'),
(@nhif_historical, 50000.00, 59999.00, NULL, 1200.00, 12, 'KES 50,000 to 59,999'),
(@nhif_historical, 60000.00, 69999.00, NULL, 1300.00, 13, 'KES 60,000 to 69,999'),
(@nhif_historical, 70000.00, 79999.00, NULL, 1400.00, 14, 'KES 70,000 to 79,999'),
(@nhif_historical, 80000.00, 89999.00, NULL, 1500.00, 15, 'KES 80,000 to 89,999'),
(@nhif_historical, 90000.00, 99999.00, NULL, 1600.00, 16, 'KES 90,000 to 99,999'),
(@nhif_historical, 100000.00, NULL, NULL, 1700.00, 17, 'KES 100,000 and above');

INSERT INTO rule_version_sources (rule_version_id, rule_source_id) VALUES (@nhif_historical, @src_nhif);

-- ============================================================================
-- SHIF: OCTOBER 2024 ONWARD
-- 2.75% of gross salary/wage
-- Minimum KES 300
-- ============================================================================

INSERT INTO rule_versions
(
    rule_definition_id, calculation_method_id, version_code, version_name,
    status, effective_from, effective_to, calculation_order,
    affects_taxable_income, affects_net_pay, source_summary, notes, published_at
)
VALUES
(
    @rule_shif, @method_pct_min, 'SHIF_2024',
    'SHIF Salaried Contribution',
    'PUBLISHED', '2024-10-01', NULL, 300,
    1, 1,
    'Social Health Insurance Regulations, 2024',
    '2.75% of gross salary or wage with a minimum contribution of KES 300.',
    CURRENT_TIMESTAMP
);

SET @shif_2024 := LAST_INSERT_ID();

INSERT INTO rule_parameters
(rule_version_id, parameter_name, parameter_type, value_decimal, value_text)
VALUES
(@shif_2024, 'rate', 'DECIMAL', 0.02750000, NULL),
(@shif_2024, 'minimum_amount', 'DECIMAL', 300.00, NULL),
(@shif_2024, 'calculation_base', 'TEXT', NULL, 'GROSS_PAY');

INSERT INTO rule_version_sources (rule_version_id, rule_source_id) VALUES (@shif_2024, @src_shif);

-- ============================================================================
-- AFFORDABLE HOUSING LEVY: INITIAL FINANCE ACT 2023 REGIME
-- Employee 1.5% of gross monthly salary
-- Employer 1.5% of gross monthly salary
--
-- The initial period is kept separate because of subsequent litigation.
-- ============================================================================

INSERT INTO rule_versions
(
    rule_definition_id, calculation_method_id, version_code, version_name,
    status, effective_from, effective_to, calculation_order,
    affects_taxable_income, affects_net_pay, source_summary, notes, published_at
)
VALUES
(
    @rule_ahl, @method_pct, 'AHL_2023_INITIAL',
    'Affordable Housing Levy Initial 2023 Regime',
    'PUBLISHED', '2023-07-01', '2023-11-27', 400,
    0, 1,
    'KRA notice: 1.5% employee contribution effective 1 July 2023',
    'Historical initial regime. Production reconciliation for this period should retain the rule source and legal-history context.',
    CURRENT_TIMESTAMP
);

SET @ahl_2023_employee := LAST_INSERT_ID();

INSERT INTO rule_parameters
(rule_version_id, parameter_name, parameter_type, value_decimal, value_text)
VALUES
(@ahl_2023_employee, 'rate', 'DECIMAL', 0.01500000, NULL),
(@ahl_2023_employee, 'calculation_base', 'TEXT', NULL, 'GROSS_MONTHLY_SALARY');

INSERT INTO rule_version_sources (rule_version_id, rule_source_id) VALUES (@ahl_2023_employee, @src_ahl_2023);

INSERT INTO rule_versions
(
    rule_definition_id, calculation_method_id, version_code, version_name,
    status, effective_from, effective_to, calculation_order,
    affects_taxable_income, affects_net_pay, source_summary, notes, published_at
)
VALUES
(
    @rule_employer_ahl, @method_pct, 'EMPLOYER_AHL_2023_INITIAL',
    'Employer Affordable Housing Levy Initial 2023 Regime',
    'PUBLISHED', '2023-07-01', '2023-11-27', 400,
    0, 0,
    'KRA notice: 1.5% employer contribution effective 1 July 2023',
    'Employer contribution is not deducted from employee take-home pay.',
    CURRENT_TIMESTAMP
);

SET @ahl_2023_employer := LAST_INSERT_ID();

INSERT INTO rule_parameters
(rule_version_id, parameter_name, parameter_type, value_decimal, value_text)
VALUES
(@ahl_2023_employer, 'rate', 'DECIMAL', 0.01500000, NULL),
(@ahl_2023_employer, 'calculation_base', 'TEXT', NULL, 'GROSS_MONTHLY_SALARY');

INSERT INTO rule_version_sources (rule_version_id, rule_source_id) VALUES (@ahl_2023_employer, @src_ahl_2023);

-- ============================================================================
-- AFFORDABLE HOUSING LEVY: AFFORDABLE HOUSING ACT 2024 REGIME
-- Employee 1.5%
-- Employer 1.5%
-- Effective 19 March 2024.
--
-- For a monthly calculator, the engine should apply the version selected by
-- the official payroll-period treatment. This row stores the legal effective
-- date and the admin UI can expose that exact date.
-- ============================================================================

INSERT INTO rule_versions
(
    rule_definition_id, calculation_method_id, version_code, version_name,
    status, effective_from, effective_to, calculation_order,
    affects_taxable_income, affects_net_pay, source_summary, notes, published_at
)
VALUES
(
    @rule_ahl, @method_pct, 'AHL_2024',
    'Affordable Housing Levy Under Affordable Housing Act 2024',
    'PUBLISHED', '2024-03-19', NULL, 400,
    1, 1,
    'KRA notice: Affordable Housing Act 2024 collection effective 19 March 2024',
    'From the Tax Laws (Amendment) Act 2024 PAYE computation changes, AHL is deductible in determining taxable employment income for December 2024 and subsequent periods.',
    CURRENT_TIMESTAMP
);

SET @ahl_2024_employee := LAST_INSERT_ID();

INSERT INTO rule_parameters
(rule_version_id, parameter_name, parameter_type, value_decimal, value_text)
VALUES
(@ahl_2024_employee, 'rate', 'DECIMAL', 0.01500000, NULL),
(@ahl_2024_employee, 'calculation_base', 'TEXT', NULL, 'GROSS_MONTHLY_SALARY');

INSERT INTO rule_version_sources (rule_version_id, rule_source_id) VALUES (@ahl_2024_employee, @src_ahl_2024);

INSERT INTO rule_versions
(
    rule_definition_id, calculation_method_id, version_code, version_name,
    status, effective_from, effective_to, calculation_order,
    affects_taxable_income, affects_net_pay, source_summary, notes, published_at
)
VALUES
(
    @rule_employer_ahl, @method_pct, 'EMPLOYER_AHL_2024',
    'Employer Affordable Housing Levy Under Affordable Housing Act 2024',
    'PUBLISHED', '2024-03-19', NULL, 400,
    0, 0,
    'KRA notice: Affordable Housing Act 2024 collection effective 19 March 2024',
    'Employer contribution shown separately and not deducted from employee net pay.',
    CURRENT_TIMESTAMP
);

SET @ahl_2024_employer := LAST_INSERT_ID();

INSERT INTO rule_parameters
(rule_version_id, parameter_name, parameter_type, value_decimal, value_text)
VALUES
(@ahl_2024_employer, 'rate', 'DECIMAL', 0.01500000, NULL),
(@ahl_2024_employer, 'calculation_base', 'TEXT', NULL, 'GROSS_MONTHLY_SALARY');

INSERT INTO rule_version_sources (rule_version_id, rule_source_id) VALUES (@ahl_2024_employer, @src_ahl_2024);

-- ============================================================================
-- TAX-ALLOWABLE DEDUCTIONS: BASELINE PENSION AND MORTGAGE LIMITS
--
-- Current KRA PAYE guidance lists:
-- Registered pension/provident/retirement fund: KES 30,000 per month
-- Mortgage interest: KES 30,000 per month
--
-- These are configured as rule versions so the admin can replace them if an
-- official historical limit changes for a later migration.
-- ============================================================================

INSERT INTO rule_versions
(
    rule_definition_id, calculation_method_id, version_code, version_name,
    status, effective_from, effective_to, calculation_order,
    affects_taxable_income, affects_net_pay, source_summary, notes, published_at
)
VALUES
(
    @rule_pension, @method_capped_pct, 'QUALIFYING_PENSION_BASELINE',
    'Qualifying Pension Contribution Monthly Limit',
    'PUBLISHED', '2022-01-01', NULL, 50,
    1, 1,
    'KRA PAYE guidance: qualifying pension/provident/retirement contributions',
    'Maximum deductible amount configured at KES 30,000 per month. Engine must use actual qualifying contribution as input and apply the cap.',
    CURRENT_TIMESTAMP
);

SET @pension_baseline := LAST_INSERT_ID();

INSERT INTO rule_parameters
(rule_version_id, parameter_name, parameter_type, value_decimal, value_text)
VALUES
(@pension_baseline, 'maximum_amount', 'DECIMAL', 30000.00, NULL),
(@pension_baseline, 'calculation_base', 'TEXT', NULL, 'ACTUAL_QUALIFYING_CONTRIBUTION');

INSERT INTO rule_version_sources (rule_version_id, rule_source_id) VALUES (@pension_baseline, @src_paye_2023);

INSERT INTO rule_versions
(
    rule_definition_id, calculation_method_id, version_code, version_name,
    status, effective_from, effective_to, calculation_order,
    affects_taxable_income, affects_net_pay, source_summary, notes, published_at
)
VALUES
(
    @rule_mortgage, @method_capped_pct, 'QUALIFYING_MORTGAGE_BASELINE',
    'Qualifying Mortgage Interest Monthly Limit',
    'PUBLISHED', '2022-01-01', NULL, 60,
    1, 1,
    'KRA PAYE guidance: qualifying mortgage interest',
    'Maximum deductible amount configured at KES 30,000 per month. Eligibility must be validated by application/business rules.',
    CURRENT_TIMESTAMP
);

SET @mortgage_baseline := LAST_INSERT_ID();

INSERT INTO rule_parameters
(rule_version_id, parameter_name, parameter_type, value_decimal, value_text)
VALUES
(@mortgage_baseline, 'maximum_amount', 'DECIMAL', 30000.00, NULL),
(@mortgage_baseline, 'calculation_base', 'TEXT', NULL, 'ACTUAL_QUALIFYING_INTEREST');

INSERT INTO rule_version_sources (rule_version_id, rule_source_id) VALUES (@mortgage_baseline, @src_paye_2023);

-- ============================================================================
-- TAX LAWS (AMENDMENT) ACT 2024 ADDITIONS
-- Applicable to December 2024 and subsequent periods according to KRA notice.
--
-- Post-retirement medical fund:
-- maximum KES 15,000 per month
-- ============================================================================

INSERT INTO rule_versions
(
    rule_definition_id, calculation_method_id, version_code, version_name,
    status, effective_from, effective_to, calculation_order,
    affects_taxable_income, affects_net_pay, source_summary, notes, published_at
)
VALUES
(
    @rule_medical_fund, @method_capped_pct, 'POST_RETIREMENT_MEDICAL_FUND_2024',
    'Post-Retirement Medical Fund Allowable Deduction',
    'PUBLISHED', '2024-12-01', NULL, 70,
    1, 1,
    'Tax Laws (Amendment) Act 2024 PAYE computation changes',
    'KRA states a limit of KES 15,000 per month.',
    CURRENT_TIMESTAMP
);

SET @medical_fund_2024 := LAST_INSERT_ID();

INSERT INTO rule_parameters
(rule_version_id, parameter_name, parameter_type, value_decimal, value_text)
VALUES
(@medical_fund_2024, 'maximum_amount', 'DECIMAL', 15000.00, NULL),
(@medical_fund_2024, 'calculation_base', 'TEXT', NULL, 'ACTUAL_QUALIFYING_CONTRIBUTION');

INSERT INTO rule_version_sources (rule_version_id, rule_source_id) VALUES (@medical_fund_2024, @src_tax_2024);

-- ============================================================================
-- TAX TREATMENT CHANGE METADATA
--
-- The existing AHL and SHIF contribution rules become tax-deductible inputs in
-- determining taxable employment income for December 2024 and subsequent
-- periods according to KRA.
--
-- Because rule_versions themselves cannot express a mid-history change in
-- tax treatment without splitting versions, we create successor versions with
-- affects_taxable_income = 1 from 2024-12-01.
-- ============================================================================

-- End pre-December 2024 AHL tax-treatment version.
UPDATE rule_versions
SET effective_to = '2024-11-30'
WHERE id = @ahl_2024_employee;

INSERT INTO rule_versions
(
    rule_definition_id, calculation_method_id, version_code, version_name,
    status, effective_from, effective_to, calculation_order,
    affects_taxable_income, affects_net_pay, source_summary, notes, published_at
)
VALUES
(
    @rule_ahl, @method_pct, 'AHL_2024_TAX_DEDUCTIBLE',
    'Affordable Housing Levy Tax-Deductible PAYE Treatment',
    'PUBLISHED', '2024-12-01', NULL, 40,
    1, 1,
    'KRA PAYE computation notice for Tax Laws (Amendment) Act 2024',
    'AHL deducted in determining taxable employment income for December 2024 and subsequent periods.',
    CURRENT_TIMESTAMP
);

SET @ahl_tax_deductible := LAST_INSERT_ID();

INSERT INTO rule_parameters
(rule_version_id, parameter_name, parameter_type, value_decimal, value_text)
VALUES
(@ahl_tax_deductible, 'rate', 'DECIMAL', 0.01500000, NULL),
(@ahl_tax_deductible, 'calculation_base', 'TEXT', NULL, 'GROSS_MONTHLY_SALARY');

INSERT INTO rule_version_sources (rule_version_id, rule_source_id) VALUES
(@ahl_tax_deductible, @src_ahl_2024),
(@ahl_tax_deductible, @src_tax_2024);

-- SHIF already starts October 2024. Split the rule for explicit tax treatment.
UPDATE rule_versions
SET effective_to = '2024-11-30'
WHERE id = @shif_2024;

INSERT INTO rule_versions
(
    rule_definition_id, calculation_method_id, version_code, version_name,
    status, effective_from, effective_to, calculation_order,
    affects_taxable_income, affects_net_pay, source_summary, notes, published_at
)
VALUES
(
    @rule_shif, @method_pct_min, 'SHIF_2024_TAX_DEDUCTIBLE',
    'SHIF Contribution Tax-Deductible PAYE Treatment',
    'PUBLISHED', '2024-12-01', NULL, 30,
    1, 1,
    'KRA PAYE computation notice for Tax Laws (Amendment) Act 2024',
    'SHIF contribution deductible in determining taxable employment income for December 2024 and subsequent periods.',
    CURRENT_TIMESTAMP
);

SET @shif_tax_deductible := LAST_INSERT_ID();

INSERT INTO rule_parameters
(rule_version_id, parameter_name, parameter_type, value_decimal, value_text)
VALUES
(@shif_tax_deductible, 'rate', 'DECIMAL', 0.02750000, NULL),
(@shif_tax_deductible, 'minimum_amount', 'DECIMAL', 300.00, NULL),
(@shif_tax_deductible, 'calculation_base', 'TEXT', NULL, 'GROSS_PAY');

INSERT INTO rule_version_sources (rule_version_id, rule_source_id) VALUES
(@shif_tax_deductible, @src_shif),
(@shif_tax_deductible, @src_tax_2024);

-- ============================================================================
-- RULE DEPENDENCY / ORDERING HINTS
-- ============================================================================

INSERT INTO rule_dependencies
(rule_version_id, depends_on_rule_code, dependency_type)
VALUES
(@paye_2021, 'NSSF', 'CALCULATION_BASE'),
(@paye_2023, 'NSSF', 'CALCULATION_BASE'),
(@ahl_tax_deductible, 'PAYE', 'MUST_RUN_BEFORE'),
(@shif_tax_deductible, 'PAYE', 'MUST_RUN_BEFORE');

-- ============================================================================
-- SANITY CHECKS
-- ============================================================================

-- There should be one current published version for current rules.
-- These queries are informational and can be run after import:
--
-- SELECT rd.code, rv.version_code, rv.effective_from, rv.effective_to
-- FROM rule_versions rv
-- JOIN rule_definitions rd ON rd.id = rv.rule_definition_id
-- WHERE rv.status = 'PUBLISHED'
-- ORDER BY rd.code, rv.effective_from;

COMMIT;
