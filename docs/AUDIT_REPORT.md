# Database Consistency Audit Report

## Scope

Audited migration sequence:

- 001 base schema
- 002 Kenya payroll seed data
- 003 schema improvements
- 004 NSSF and historical rule refinement

## Findings

### PASS — Foreign-key base types

The base schema uses `BIGINT UNSIGNED` primary keys for `calculations.id`.
The earlier draft of migration 003 used `CHAR(36)` for snapshot and rule-set
calculation IDs. That was inconsistent and prevented safe foreign-key linkage.

**Correction applied:** migration 003 now uses `BIGINT UNSIGNED calculation_id`
and references `calculations(id)`.

### PASS — Historical reproducibility

A rules hash alone cannot prove which versions were used.

**Correction applied:** `calculation_rule_set_items` now stores each resolved
`rule_version_id` and `rule_definition_id`.

### PASS — NSSF phased structure

The seed contains four phased employee NSSF versions:

- Year 1: LEL 6,000 / UEL 18,000
- Year 2: LEL 7,000 / UEL 36,000
- Year 3: LEL 8,000 / UEL 72,000
- Year 4: LEL 9,000 / UEL 108,000

Migration 004 adds explicit Tier I and Tier II components at 6% employee and 6%
employer rates.

### PASS — AHL and SHIF historical tax treatment split

Migration 002 already closes the pre-December 2024 versions and creates successor
versions from 1 December 2024 for the PAYE tax-treatment change. Migration 004
adds explicit tax-treatment records to make that history queryable separately.

### PASS — Calculation periods

Monthly periods from January 2022 through December 2026 are seeded. Audit query
005 checks for gaps or overlaps.

### REMAINING REQUIRED EXECUTION CHECK

A live MySQL 8.0+ execution was not available in this packaging environment.
Therefore, `005_audit_checks.sql` is included and must be executed immediately
after import. It performs database-level checks against the actual imported
schema and seed data.

## Result

**STATIC AUDIT: PASS WITH CORRECTIONS APPLIED**

The package is ready for clean-database import and live MySQL audit execution.
