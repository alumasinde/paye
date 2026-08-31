# Budget254 PAYE Calculator — Database Package

This package contains the audited MySQL migration sequence for the dynamic,
historical Kenya PAYE calculator.

## Migration order

1. `001_budget254_paye.sql`
2. `002_seed_kenya_payroll_rules.sql`
3. `003_improve_budget254_paye_schema.sql`
4. `004_correct_and_refine_kenya_payroll_rules.sql`

After importing migrations 001–004, run:

5. `005_audit_checks.sql`

The audit script is read-only. Failure queries should return zero rows.

## Important architecture

- Rule definitions are stable concepts such as PAYE, NSSF, SHIF and AHL.
- Rule versions are effective-dated historical versions.
- Calculation formulas are implemented by Go and selected by database data.
- NSSF phased rates are represented using explicit Tier I and Tier II components.
- Contribution calculation and PAYE tax treatment are stored separately.
- Calculation periods prevent historical-month date ambiguity.
- Calculation rule sets and rule-set items preserve exactly which published
  rule versions were used for a saved calculation.
- Calculation snapshots preserve input, resolved rules, steps and output.

## Audit improvements applied

The consistency audit identified and corrected a schema mismatch in the earlier
draft where calculation rule sets and snapshots used `CHAR(36)` calculation IDs
while the base `calculations.id` is `BIGINT UNSIGNED`.

The audited migration now:

- uses `BIGINT UNSIGNED calculation_id`;
- adds foreign keys to `calculations(id)`;
- adds `calculation_rule_set_items` so exact resolved rule versions are stored,
  rather than storing only a hash;
- includes `005_audit_checks.sql` for overlap, mapping, NSSF, tax-treatment and
  calculation-period consistency checks.

## Production note

Run this package first on a clean MySQL 8.0+ database. Keep the migration files
immutable after production use. Future statutory changes should be new numbered
migrations and new effective-dated rule versions.

## Recommended next backend milestone

Build the Go Rule Resolver first, then the formula engine, then test cases for
historical payroll months before exposing the Android API.
